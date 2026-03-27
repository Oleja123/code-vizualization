package com.metrics.api;

import com.metrics.entity.FunctionMetricsEntity;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import com.metrics.service.MetricsService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/api/metrics")
@RequiredArgsConstructor
public class MetricsController {

    private final MetricsService metricsService;

    @PostMapping("/calculate")
    public ResponseEntity<?> calculate(@RequestBody Map<String, String> body, Authentication auth) {
        String code = body.get("code");
        if (code == null || code.isBlank())
            return ResponseEntity.badRequest().body(Map.of("error", "Field 'code' is required"));

        String username = auth.getName();
        long count = metricsService.countByUsername(username);
        if (count >= MetricsService.MAX_RECORDS_PER_USER) {
            return ResponseEntity.status(429).body(Map.of(
                    "error", "Достигнут лимит " + MetricsService.MAX_RECORDS_PER_USER +
                            " записей. Удалите часть истории, чтобы сохранить новые метрики.",
                    "limitExceeded", true,
                    "count", count
            ));
        }

        try {
            ProgramMetrics metrics = metricsService.calculateAndSave(code, username);
            return ResponseEntity.ok(metrics);
        } catch (Exception e) {
            log.warn("Metrics calculation failed for user={}: {}", username, e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping("/latest")
    public ResponseEntity<List<FunctionMetrics>> latest(Authentication auth) {
        return ResponseEntity.ok(metricsService.getLatest(auth.getName()));
    }

    @GetMapping("/history")
    public ResponseEntity<List<FunctionMetricsEntity>> history(Authentication auth) {
        return ResponseEntity.ok(metricsService.getHistory(auth.getName()));
    }

    @GetMapping("/count")
    public ResponseEntity<Map<String, Long>> count(Authentication auth) {
        long count = metricsService.countByUsername(auth.getName());
        return ResponseEntity.ok(Map.of("count", count, "limit", (long) MetricsService.MAX_RECORDS_PER_USER));
    }

    /** Удалить одну запись */
    @DeleteMapping("/{id}")
    public ResponseEntity<?> delete(@PathVariable Long id, Authentication auth) {
        boolean deleted = metricsService.deleteById(id, auth.getName());
        if (!deleted)
            return ResponseEntity.status(403).body(Map.of("error", "Not found or access denied"));
        return ResponseEntity.ok(Map.of("deleted", id));
    }

    /** Удалить несколько записей сразу */
    @DeleteMapping("/batch")
    public ResponseEntity<?> deleteBatch(@RequestBody Map<String, List<Long>> body, Authentication auth) {
        List<Long> ids = body.get("ids");
        if (ids == null || ids.isEmpty())
            return ResponseEntity.badRequest().body(Map.of("error", "ids required"));
        int deleted = metricsService.deleteByIds(ids, auth.getName());
        return ResponseEntity.ok(Map.of("deleted", deleted));
    }

    /** Удалить все записи за конкретную дату (YYYY-MM-DD) */
    @DeleteMapping("/by-date/{date}")
    public ResponseEntity<?> deleteByDate(@PathVariable String date, Authentication auth) {
        int deleted = metricsService.deleteByDate(date, auth.getName());
        return ResponseEntity.ok(Map.of("deleted", deleted));
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok", "service", "metrics-service"));
    }
}