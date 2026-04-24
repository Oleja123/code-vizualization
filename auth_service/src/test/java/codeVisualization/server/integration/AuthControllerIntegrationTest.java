package codeVisualization.server.integration;

import codeVisualization.server.db.entity.User;
import codeVisualization.server.db.jpaRepository.UserRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.mock.web.MockHttpSession;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Интеграционные тесты auth_service.
 *
 * Поднимает PostgreSQL через Testcontainers, запускает Liquibase-миграции,
 * проверяет полный HTTP-цикл (через MockMvc): регистрация → логин → /me → выход.
 *
 * Зависимости (добавить в pom.xml auth_service, scope test):
 *   org.testcontainers:junit-jupiter:1.19.7
 *   org.testcontainers:postgresql:1.19.7
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@Testcontainers
@DisplayName("AuthController — интеграционные тесты")
class AuthControllerIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres =
            new PostgreSQLContainer<>("postgres:16-alpine")
                    .withDatabaseName("testdb")
                    .withUsername("test")
                    .withPassword("test");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        // Отключаем Redis и Kafka для теста
        registry.add("spring.autoconfigure.exclude",
                () -> "org.springframework.boot.autoconfigure.data.redis.RedisAutoConfiguration," +
                        "org.springframework.boot.autoconfigure.kafka.KafkaAutoConfiguration");
    }

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private PasswordEncoder passwordEncoder;

    @Autowired
    private ObjectMapper objectMapper;

    @BeforeEach
    void cleanUp() {
        userRepository.deleteAll();
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/auth/register
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/register")
    class Register {

        @Test
        @DisplayName("Регистрация нового пользователя → 201, пользователь сохранён в БД")
        void registerNewUser() throws Exception {
            mockMvc.perform(post("/api/auth/register")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("alice", "password123")))
                    .andExpect(status().isCreated())
                    .andExpect(content().string(org.hamcrest.Matchers.containsString("создан")));

            assertThat(userRepository.findByName("alice")).isPresent();
            User saved = userRepository.findByName("alice").get();
            assertThat(passwordEncoder.matches("password123", saved.getPassword())).isTrue();
        }

        @Test
        @DisplayName("Повторная регистрация того же пользователя → 409 Conflict")
        void registerDuplicateUser() throws Exception {
            createUser("alice", "pass");

            mockMvc.perform(post("/api/auth/register")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("alice", "newpass")))
                    .andExpect(status().isConflict());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/auth/login
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/login")
    class Login {

        @Test
        @DisplayName("Верные credentials → 200, сессия установлена")
        void successfulLogin() throws Exception {
            createUser("bob", "secret");

            MvcResult result = mockMvc.perform(post("/api/auth/login")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("bob", "secret")))
                    .andExpect(status().isOk())
                    .andReturn();

            assertThat(result.getResponse().getContentAsString()).contains("Успешный");
        }

        @Test
        @DisplayName("Несуществующий пользователь → 401")
        void unknownUser() throws Exception {
            mockMvc.perform(post("/api/auth/login")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("ghost", "any")))
                    .andExpect(status().isUnauthorized());
        }

        @Test
        @DisplayName("Неверный пароль → 401")
        void wrongPassword() throws Exception {
            createUser("charlie", "correctpass");

            mockMvc.perform(post("/api/auth/login")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("charlie", "wrongpass")))
                    .andExpect(status().isUnauthorized());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/auth/me
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("GET /api/auth/me")
    class Me {

        @Test
        @DisplayName("С активной сессией → 200, имя пользователя")
        void withSession() throws Exception {
            createUser("diana", "pass");

            // Логинимся и берём сессию
            MvcResult loginResult = mockMvc.perform(post("/api/auth/login")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("diana", "pass")))
                    .andExpect(status().isOk())
                    .andReturn();

            MockHttpSession session = (MockHttpSession) loginResult.getRequest().getSession();

            // Используем ту же сессию для /me
            mockMvc.perform(get("/api/auth/me").session(session))
                    .andExpect(status().isOk())
                    .andExpect(content().string("diana"));
        }

        @Test
        @DisplayName("Без сессии → 401")
        void withoutSession() throws Exception {
            mockMvc.perform(get("/api/auth/me"))
                    .andExpect(status().isUnauthorized());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/auth/logout
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/logout")
    class Logout {

        @Test
        @DisplayName("Выход инвалидирует сессию: /me после → 401")
        void logoutInvalidatesSession() throws Exception {
            createUser("eve", "pass");

            // Логин
            MvcResult loginResult = mockMvc.perform(post("/api/auth/login")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(jsonCreds("eve", "pass")))
                    .andExpect(status().isOk())
                    .andReturn();
            MockHttpSession session = (MockHttpSession) loginResult.getRequest().getSession();

            // /me работает
            mockMvc.perform(get("/api/auth/me").session(session))
                    .andExpect(status().isOk());

            // Логаут
            mockMvc.perform(post("/api/auth/logout").session(session))
                    .andExpect(status().isOk());

            // /me без сессии → 401
            mockMvc.perform(get("/api/auth/me"))
                    .andExpect(status().isUnauthorized());
        }

        @Test
        @DisplayName("Выход без сессии → 200 (идемпотентный)")
        void logoutWithoutSession() throws Exception {
            mockMvc.perform(post("/api/auth/logout"))
                    .andExpect(status().isOk());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Полный сценарий
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("Полный цикл: register → login → /me → logout → /me (401)")
    void fullCycle() throws Exception {
        // 1. Регистрация
        mockMvc.perform(post("/api/auth/register")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(jsonCreds("frank", "mypassword")))
                .andExpect(status().isCreated());

        // 2. Логин
        MvcResult loginResult = mockMvc.perform(post("/api/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(jsonCreds("frank", "mypassword")))
                .andExpect(status().isOk())
                .andReturn();
        MockHttpSession session = (MockHttpSession) loginResult.getRequest().getSession();

        // 3. Проверка сессии
        mockMvc.perform(get("/api/auth/me").session(session))
                .andExpect(status().isOk())
                .andExpect(content().string("frank"));

        // 4. Выход
        mockMvc.perform(post("/api/auth/logout").session(session))
                .andExpect(status().isOk());

        // 5. Убеждаемся что сессия недействительна
        mockMvc.perform(get("/api/auth/me"))
                .andExpect(status().isUnauthorized());
    }

    // ─────────────────────────────────────────────────────────────
    //  Хелперы
    // ─────────────────────────────────────────────────────────────

    private void createUser(String name, String rawPassword) {
        User u = new User();
        u.setName(name);
        u.setPassword(passwordEncoder.encode(rawPassword));
        userRepository.save(u);
    }

    private String jsonCreds(String username, String password) throws Exception {
        return objectMapper.writeValueAsString(
                Map.of("username", username, "rawPassword", password));
    }
}