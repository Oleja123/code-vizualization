package com.metrics.api;

import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import com.metrics.service.MetricsService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.*;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;

import java.util.Collections;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@DisplayName("MetricsController — юнит-тесты")
class MetricsControllerTest {

    @Mock
    private MetricsService metricsService;

    @InjectMocks
    private MetricsController controller;

    private Authentication auth;

    @BeforeEach
    void setUp() {
        auth = new UsernamePasswordAuthenticationToken("alice", null, Collections.emptyList());
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/metrics/calculate
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /calculate")
    class Calculate {

        @Test
        @DisplayName("Успешный расчёт: 200 OK с метриками")
        void success() throws Exception {
            ProgramMetrics metrics = new ProgramMetrics();
            metrics.setFunctionCount(2);
            when(metricsService.countByUsername("alice")).thenReturn(0L);
            when(metricsService.calculateAndSave("int main(){}", "alice")).thenReturn(metrics);

            ResponseEntity<?> resp = controller.calculate(Map.of("code", "int main(){}"), auth);

            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(resp.getBody()).isInstanceOf(ProgramMetrics.class);
            assertThat(((ProgramMetrics) resp.getBody()).getFunctionCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("Пустой code → 400 Bad Request")
        void emptyCode() {
            ResponseEntity<?> resp = controller.calculate(Map.of("code", "  "), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("Отсутствующий code → 400 Bad Request")
        void missingCode() {
            ResponseEntity<?> resp = controller.calculate(Map.of(), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("Лимит записей превышен → 429")
        void limitExceeded() {
            when(metricsService.countByUsername("alice")).thenReturn((long) MetricsService.MAX_RECORDS_PER_USER);

            ResponseEntity<?> resp = controller.calculate(Map.of("code", "int main(){}"), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.TOO_MANY_REQUESTS);

            @SuppressWarnings("unchecked")
            Map<String, Object> body = (Map<String, Object>) resp.getBody();
            assertThat(body).containsKey("limitExceeded");
            assertThat(body.get("limitExceeded")).isEqualTo(true);
        }

        @Test
        @DisplayName("calculateAndSave бросает исключение → 400")
        void serviceException() throws Exception {
            when(metricsService.countByUsername("alice")).thenReturn(0L);
            when(metricsService.calculateAndSave(any(), any())).thenThrow(new RuntimeException("parse error"));

            ResponseEntity<?> resp = controller.calculate(Map.of("code", "bad code"), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/metrics/latest
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("GET /latest")
    class Latest {

        @Test
        @DisplayName("Возвращает последние метрики пользователя")
        void success() {
            FunctionMetrics fm = new FunctionMetrics();
            fm.setFunctionName("main");
            when(metricsService.getLatest("alice")).thenReturn(List.of(fm));

            ResponseEntity<List<FunctionMetrics>> resp = controller.latest(auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(resp.getBody()).hasSize(1);
            assertThat(resp.getBody().get(0).getFunctionName()).isEqualTo("main");
        }

        @Test
        @DisplayName("Пустой список если нет записей")
        void empty() {
            when(metricsService.getLatest("alice")).thenReturn(List.of());
            ResponseEntity<List<FunctionMetrics>> resp = controller.latest(auth);
            assertThat(resp.getBody()).isEmpty();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/metrics/count
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("GET /count")
    class Count {

        @Test
        @DisplayName("Возвращает count и limit")
        void success() {
            when(metricsService.countByUsername("alice")).thenReturn(10L);

            ResponseEntity<Map<String, Long>> resp = controller.count(auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(resp.getBody()).containsEntry("count", 10L);
            assertThat(resp.getBody()).containsEntry("limit", (long) MetricsService.MAX_RECORDS_PER_USER);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  DELETE /api/metrics/{id}
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("DELETE /{id}")
    class Delete {

        @Test
        @DisplayName("Успешное удаление: 200 с id")
        void success() {
            when(metricsService.deleteById(5L, "alice")).thenReturn(true);

            ResponseEntity<?> resp = controller.delete(5L, auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
            @SuppressWarnings("unchecked")
            Map<String, Object> body = (Map<String, Object>) resp.getBody();
            assertThat(body).containsEntry("deleted", 5L);
        }

        @Test
        @DisplayName("Запись не найдена или чужая: 403")
        void forbidden() {
            when(metricsService.deleteById(99L, "alice")).thenReturn(false);
            ResponseEntity<?> resp = controller.delete(99L, auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.FORBIDDEN);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  DELETE /api/metrics/batch
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("DELETE /batch")
    class DeleteBatch {

        @Test
        @DisplayName("Успешное пакетное удаление")
        void success() {
            when(metricsService.deleteByIds(List.of(1L, 2L), "alice")).thenReturn(2);

            ResponseEntity<?> resp = controller.deleteBatch(Map.of("ids", List.of(1L, 2L)), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
        }

        @Test
        @DisplayName("Пустой список ids → 400")
        void emptyIds() {
            ResponseEntity<?> resp = controller.deleteBatch(Map.of("ids", List.of()), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("Отсутствующий ids → 400")
        void missingIds() {
            ResponseEntity<?> resp = controller.deleteBatch(Map.of(), auth);
            assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/metrics/health
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("GET /health → 200 OK с status=ok")
    void health() {
        ResponseEntity<Map<String, String>> resp = controller.health();
        assertThat(resp.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(resp.getBody()).containsEntry("status", "ok");
    }
}