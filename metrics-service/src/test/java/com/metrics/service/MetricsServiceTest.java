package com.metrics.service;

import com.metrics.ast.Program;
import com.metrics.calculator.MetricsCalculator;
import com.metrics.entity.FunctionMetricsEntity;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import com.metrics.repository.FunctionMetricsRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.*;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@DisplayName("MetricsService — юнит-тесты")
class MetricsServiceTest {

    @Mock
    private MetricsCalculator calculator;

    @Mock
    private SemanticAnalyzerClient astClient;

    @Mock
    private FunctionMetricsRepository repository;

    @InjectMocks
    private MetricsService service;

    // ─────────────────────────────────────────────────────────────
    //  calculateAndSave
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("calculateAndSave")
    class CalculateAndSave {

        @Test
        @DisplayName("Вызывает astClient.parse → calculator.calculate → repository.saveAll")
        void happyPath() throws Exception {
            Program program = new Program();
            ProgramMetrics metrics = new ProgramMetrics();
            FunctionMetrics fm = new FunctionMetrics();
            fm.setFunctionName("main");
            fm.setLoc(10);
            fm.setCyclomaticComplexity(3);
            metrics.setFunctions(List.of(fm));
            metrics.setFunctionCount(1);

            when(astClient.parse("code")).thenReturn(program);
            when(calculator.calculate(program)).thenReturn(metrics);
            when(repository.saveAll(anyList())).thenAnswer(inv -> inv.getArgument(0));

            ProgramMetrics result = service.calculateAndSave("code", "alice");

            assertThat(result.getFunctionCount()).isEqualTo(1);
            verify(astClient).parse("code");
            verify(calculator).calculate(program);

            ArgumentCaptor<List<FunctionMetricsEntity>> captor = ArgumentCaptor.forClass(List.class);
            verify(repository).saveAll(captor.capture());
            assertThat(captor.getValue()).hasSize(1);
            assertThat(captor.getValue().get(0).getFunctionName()).isEqualTo("main");
            assertThat(captor.getValue().get(0).getUsername()).isEqualTo("alice");
            assertThat(captor.getValue().get(0).getLoc()).isEqualTo(10);
        }

        @Test
        @DisplayName("Если функций нет, saveAll вызывается с пустым списком")
        void noFunctions() throws Exception {
            Program program = new Program();
            ProgramMetrics metrics = new ProgramMetrics();
            metrics.setFunctions(List.of());

            when(astClient.parse(any())).thenReturn(program);
            when(calculator.calculate(program)).thenReturn(metrics);
            when(repository.saveAll(anyList())).thenReturn(List.of());

            service.calculateAndSave("empty", "user");

            ArgumentCaptor<List<FunctionMetricsEntity>> captor = ArgumentCaptor.forClass(List.class);
            verify(repository).saveAll(captor.capture());
            assertThat(captor.getValue()).isEmpty();
        }

        @Test
        @DisplayName("Если astClient.parse бросает исключение — оно прокидывается наверх")
        void parseException() throws Exception {
            when(astClient.parse(any())).thenThrow(new RuntimeException("parse error"));

            org.junit.jupiter.api.Assertions.assertThrows(Exception.class,
                    () -> service.calculateAndSave("bad", "user"));
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  countByUsername
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("countByUsername")
    class CountByUsername {

        @Test
        @DisplayName("Делегирует в repository.countByUsername")
        void delegates() {
            when(repository.countByUsername("alice")).thenReturn(7L);
            assertThat(service.countByUsername("alice")).isEqualTo(7L);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteById
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteById")
    class DeleteById {

        @Test
        @DisplayName("Удаляет если запись принадлежит пользователю, возвращает true")
        void deleteOwned() {
            FunctionMetricsEntity entity = new FunctionMetricsEntity();
            entity.setId(1L);
            entity.setUsername("alice");
            when(repository.findById(1L)).thenReturn(Optional.of(entity));

            boolean result = service.deleteById(1L, "alice");
            assertThat(result).isTrue();
            verify(repository).delete(entity);
        }

        @Test
        @DisplayName("Не удаляет если запись чужая, возвращает false")
        void deleteNotOwned() {
            FunctionMetricsEntity entity = new FunctionMetricsEntity();
            entity.setId(1L);
            entity.setUsername("bob");
            when(repository.findById(1L)).thenReturn(Optional.of(entity));

            boolean result = service.deleteById(1L, "alice");
            assertThat(result).isFalse();
            verify(repository, never()).delete(any());
        }

        @Test
        @DisplayName("Запись не найдена — возвращает false")
        void deleteNotFound() {
            when(repository.findById(99L)).thenReturn(Optional.empty());
            assertThat(service.deleteById(99L, "alice")).isFalse();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteByIds
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteByIds")
    class DeleteByIds {

        @Test
        @DisplayName("Удаляет только записи текущего пользователя из списка")
        void deletesOnlyOwned() {
            FunctionMetricsEntity own1 = entity(1L, "alice");
            FunctionMetricsEntity own2 = entity(2L, "alice");
            FunctionMetricsEntity other = entity(3L, "bob");

            when(repository.findAllById(List.of(1L, 2L, 3L)))
                    .thenReturn(List.of(own1, own2, other));

            int deleted = service.deleteByIds(List.of(1L, 2L, 3L), "alice");
            assertThat(deleted).isEqualTo(2);
            verify(repository).deleteAll(argThat(l -> ((java.util.Collection<?>) l).size() == 2));
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  getHistory
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("getHistory")
    class GetHistory {

        @Test
        @DisplayName("Делегирует в repository.findByUsernameOrderByCreatedAtDesc")
        void delegates() {
            List<FunctionMetricsEntity> expected = List.of(entity(1L, "alice"));
            when(repository.findByUsernameOrderByCreatedAtDesc("alice")).thenReturn(expected);

            List<FunctionMetricsEntity> result = service.getHistory("alice");
            assertThat(result).isSameAs(expected);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  deleteByDate
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("deleteByDate")
    class DeleteByDate {

        @Test
        @DisplayName("Парсит дату и удаляет записи за день")
        void deletesByDate() {
            FunctionMetricsEntity e = entity(5L, "alice");
            when(repository.findByUsernameAndCreatedAtBetween(
                    eq("alice"), any(LocalDateTime.class), any(LocalDateTime.class)))
                    .thenReturn(List.of(e));

            int deleted = service.deleteByDate("2024-01-15", "alice");
            assertThat(deleted).isEqualTo(1);
            verify(repository).deleteAll(List.of(e));
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  getLatest
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("getLatest")
    class GetLatest {

        @Test
        @DisplayName("Конвертирует entity в модель")
        void convertsToModel() {
            FunctionMetricsEntity e = entity(1L, "alice");
            e.setFunctionName("foo");
            e.setLoc(15);
            e.setCyclomaticComplexity(4);
            e.setParameterCount(2);
            e.setMaxNestingDepth(3);
            e.setCallCount(5);
            e.setReturnCount(1);
            e.setGotoCount(0);
            when(repository.findLatestByUsername("alice")).thenReturn(List.of(e));

            List<FunctionMetrics> result = service.getLatest("alice");
            assertThat(result).hasSize(1);
            FunctionMetrics fm = result.get(0);
            assertThat(fm.getFunctionName()).isEqualTo("foo");
            assertThat(fm.getLoc()).isEqualTo(15);
            assertThat(fm.getCyclomaticComplexity()).isEqualTo(4);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Хелперы
    // ─────────────────────────────────────────────────────────────

    private FunctionMetricsEntity entity(Long id, String username) {
        FunctionMetricsEntity e = new FunctionMetricsEntity();
        e.setId(id);
        e.setUsername(username);
        e.setFunctionName("func" + id);
        e.setCreatedAt(LocalDateTime.now());
        return e;
    }
}