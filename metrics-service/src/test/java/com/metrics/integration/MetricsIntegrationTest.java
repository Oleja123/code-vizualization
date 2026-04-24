package com.metrics.integration;

import com.metrics.calculator.MetricsCalculator;
import com.metrics.entity.FunctionMetricsEntity;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import com.metrics.repository.FunctionMetricsRepository;
import com.metrics.service.MetricsService;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.test.web.servlet.MockMvc;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Интеграционные тесты metrics-service.
 *
 * Поднимает реальный PostgreSQL (Testcontainers).
 * SemanticAnalyzerClient мокается — внешней зависимости нет.
 * Все остальные слои (Calculator, Service, Repository) работают по-настоящему.
 *
 * Зависимости (добавить в pom.xml metrics-service, scope test):
 *   org.testcontainers:junit-jupiter:1.19.7
 *   org.testcontainers:postgresql:1.19.7
 *   org.springframework.security:spring-security-test (уже в spring-boot-starter-test)
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@Testcontainers
@DisplayName("MetricsService/Controller — интеграционные тесты")
class MetricsIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres =
            new PostgreSQLContainer<>("postgres:16-alpine")
                    .withDatabaseName("metricsdb")
                    .withUsername("test")
                    .withPassword("test");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        // Отключаем реальный auth-сервис, подставляем заглушку URL
        registry.add("auth.service.url", () -> "http://localhost:9999");
    }

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private FunctionMetricsRepository repository;

    @Autowired
    private MetricsService metricsService;

    // Мокаем только внешний HTTP-клиент
    @MockBean
    private com.metrics.service.SemanticAnalyzerClient astClient;

    @BeforeEach
    void cleanUp() {
        repository.deleteAll();
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/metrics/health
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("GET /health → 200 без аутентификации")
    void healthPublic() throws Exception {
        mockMvc.perform(get("/api/metrics/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("ok"));
    }

    // ─────────────────────────────────────────────────────────────
    //  MetricsCalculator + MetricsService + Repository (сквозной тест)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Сквозной тест: Calculator → Service → Repository")
    class EndToEnd {

        @Test
        @DisplayName("calculateAndSave сохраняет метрики в БД, getLatest возвращает их")
        void calculateAndPersist() throws Exception {
            // Подготавливаем AST-ответ от мока
            com.metrics.ast.Program program = buildSimpleProgram();
            when(astClient.parse(any())).thenReturn(program);

            // Вызываем сервис
            ProgramMetrics metrics = metricsService.calculateAndSave("int main(){return 0;}", "alice");

            // Проверяем что сохранилось в БД
            List<FunctionMetricsEntity> saved = repository.findByUsernameOrderByCreatedAtDesc("alice");
            assertThat(saved).isNotEmpty();
            assertThat(saved.get(0).getFunctionName()).isEqualTo("main");
            assertThat(saved.get(0).getUsername()).isEqualTo("alice");

            // getLatest возвращает те же данные
            List<FunctionMetrics> latest = metricsService.getLatest("alice");
            assertThat(latest).isNotEmpty();
            assertThat(latest.get(0).getFunctionName()).isEqualTo("main");
        }

        @Test
        @DisplayName("countByUsername возвращает корректное число")
        void countPersisted() throws Exception {
            com.metrics.ast.Program program = buildSimpleProgram();
            when(astClient.parse(any())).thenReturn(program);

            metricsService.calculateAndSave("code1", "bob");
            metricsService.calculateAndSave("code2", "bob");

            assertThat(metricsService.countByUsername("bob")).isEqualTo(2);
            assertThat(metricsService.countByUsername("alice")).isZero();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteById
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteById — интеграция с БД")
    class DeleteByIdIntegration {

        @Test
        @DisplayName("Удаляет свою запись, не трогает чужую")
        void deleteOwn() {
            FunctionMetricsEntity aliceEntry = save("alice");
            FunctionMetricsEntity bobEntry   = save("bob");

            boolean result = metricsService.deleteById(aliceEntry.getId(), "alice");
            assertThat(result).isTrue();
            assertThat(repository.findById(aliceEntry.getId())).isEmpty();
            assertThat(repository.findById(bobEntry.getId())).isPresent();
        }

        @Test
        @DisplayName("Нельзя удалить чужую запись")
        void cannotDeleteOthers() {
            FunctionMetricsEntity bobEntry = save("bob");

            boolean result = metricsService.deleteById(bobEntry.getId(), "alice");
            assertThat(result).isFalse();
            assertThat(repository.findById(bobEntry.getId())).isPresent();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteByDate
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteByDate — интеграция с БД")
    class DeleteByDateIntegration {

        @Test
        @DisplayName("Удаляет только записи за указанный день")
        void deleteBySpecificDate() {
            String today = LocalDate.now().toString();
            FunctionMetricsEntity e1 = save("alice");
            FunctionMetricsEntity e2 = save("alice");

            int deleted = metricsService.deleteByDate(today, "alice");
            assertThat(deleted).isEqualTo(2);
            assertThat(repository.findByUsernameOrderByCreatedAtDesc("alice")).isEmpty();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteByIds
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteByIds — интеграция с БД")
    class DeleteByIdsIntegration {

        @Test
        @DisplayName("Пакетное удаление — только своих записей")
        void batchDeleteOwned() {
            FunctionMetricsEntity a1 = save("alice");
            FunctionMetricsEntity a2 = save("alice");
            FunctionMetricsEntity b1 = save("bob");

            int deleted = metricsService.deleteByIds(
                    List.of(a1.getId(), a2.getId(), b1.getId()), "alice");

            assertThat(deleted).isEqualTo(2);
            assertThat(repository.findById(b1.getId())).isPresent();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Лимит записей
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("Лимит MAX_RECORDS_PER_USER работает через MetricsService.countByUsername")
    void limitCheck() throws Exception {
        com.metrics.ast.Program program = buildSimpleProgram();
        when(astClient.parse(any())).thenReturn(program);

        // Создаём MAX_RECORDS_PER_USER записей
        for (int i = 0; i < MetricsService.MAX_RECORDS_PER_USER; i++) {
            metricsService.calculateAndSave("code" + i, "alice");
        }

        assertThat(metricsService.countByUsername("alice"))
                .isEqualTo(MetricsService.MAX_RECORDS_PER_USER);
    }

    // ─────────────────────────────────────────────────────────────
    //  HTTP-слой с @WithMockUser
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("HTTP-эндпоинты (MockMvc + @WithMockUser)")
    class HttpEndpoints {

        @Test
        @DisplayName("GET /latest без авторизации → 401")
        void latestUnauthorized() throws Exception {
            mockMvc.perform(get("/api/metrics/latest"))
                    .andExpect(status().isUnauthorized());
        }

        @Test
        @WithMockUser(username = "carol")
        @DisplayName("GET /latest авторизованно → 200 с пустым списком")
        void latestAuthorized() throws Exception {
            mockMvc.perform(get("/api/metrics/latest"))
                    .andExpect(status().isOk())
                    .andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_JSON));
        }

        @Test
        @WithMockUser(username = "carol")
        @DisplayName("GET /count → 200 с count и limit")
        void count() throws Exception {
            mockMvc.perform(get("/api/metrics/count"))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.count").exists())
                    .andExpect(jsonPath("$.limit").value(MetricsService.MAX_RECORDS_PER_USER));
        }

        @Test
        @WithMockUser(username = "dave")
        @DisplayName("POST /calculate с пустым code → 400")
        void calculateEmptyCode() throws Exception {
            mockMvc.perform(post("/api/metrics/calculate")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"code\":\"\"}"))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @WithMockUser(username = "dave")
        @DisplayName("DELETE /batch без ids → 400")
        void deleteBatchEmpty() throws Exception {
            mockMvc.perform(delete("/api/metrics/batch")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"ids\":[]}"))
                    .andExpect(status().isBadRequest());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Хелперы
    // ─────────────────────────────────────────────────────────────

    private FunctionMetricsEntity save(String username) {
        FunctionMetricsEntity e = new FunctionMetricsEntity();
        e.setUsername(username);
        e.setFunctionName("main");
        e.setLoc(10);
        e.setCyclomaticComplexity(2);
        e.setParameterCount(0);
        e.setMaxNestingDepth(1);
        e.setCallCount(0);
        e.setReturnCount(1);
        e.setGotoCount(0);
        e.setCreatedAt(LocalDateTime.now());
        return repository.save(e);
    }

    private com.metrics.ast.Program buildSimpleProgram() {
        com.metrics.ast.Program p = new com.metrics.ast.Program();

        com.metrics.ast.FunctionDecl fn = new com.metrics.ast.FunctionDecl();
        fn.setName("main");
        fn.setParameters(List.of());

        com.metrics.ast.ASTLocation loc = new com.metrics.ast.ASTLocation();
        loc.setLine(1);
        loc.setEndLine(3);
        fn.setLocation(loc);

        com.metrics.ast.BlockStmt body = new com.metrics.ast.BlockStmt();
        body.setStatements(List.of());
        fn.setBody(body);

        p.setDeclarations(List.of(fn));
        return p;
    }
}