package com.metrics.service;

import com.metrics.calculator.MetricsCalculator;
import com.metrics.ast.Program;
import com.metrics.entity.FunctionMetricsEntity;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import com.metrics.repository.FunctionMetricsRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class MetricsService {

    // Легко меняемый лимит
    public static final int MAX_RECORDS_PER_USER = 50;

    private final MetricsCalculator calculator;
    private final SemanticAnalyzerClient astClient;
    private final FunctionMetricsRepository repository;

    @Transactional
    public ProgramMetrics calculateAndSave(String code, String username) throws Exception {
        log.info("Calculating metrics for username={}", username);
        Program program = astClient.parse(code);
        ProgramMetrics metrics = calculator.calculate(program);

        LocalDateTime now = LocalDateTime.now();
        List<FunctionMetricsEntity> entities = metrics.getFunctions().stream()
                .map(fm -> toEntity(fm, username, now))
                .collect(Collectors.toList());
        repository.saveAll(entities);

        log.info("Saved {} function metrics for username={}", entities.size(), username);
        return metrics;
    }

    public long countByUsername(String username) {
        return repository.countByUsername(username);
    }

    public List<FunctionMetrics> getLatest(String username) {
        return repository.findLatestByUsername(username).stream()
                .map(this::toModel)
                .collect(Collectors.toList());
    }

    public List<FunctionMetricsEntity> getHistory(String username) {
        return repository.findByUsernameOrderByCreatedAtDesc(username);
    }

    public boolean deleteById(Long id, String username) {
        return repository.findById(id).map(e -> {
            if (!e.getUsername().equals(username)) return false;
            repository.delete(e);
            return true;
        }).orElse(false);
    }

    @Transactional
    public int deleteByIds(List<Long> ids, String username) {
        List<FunctionMetricsEntity> entities = repository.findAllById(ids).stream()
                .filter(e -> e.getUsername().equals(username))
                .collect(Collectors.toList());
        repository.deleteAll(entities);
        return entities.size();
    }

    @Transactional
    public int deleteByDate(String dateStr, String username) {
        LocalDate date = LocalDate.parse(dateStr);
        LocalDateTime from = date.atStartOfDay();
        LocalDateTime to = date.plusDays(1).atStartOfDay();
        List<FunctionMetricsEntity> entities =
                repository.findByUsernameAndCreatedAtBetween(username, from, to);
        repository.deleteAll(entities);
        return entities.size();
    }

    private FunctionMetricsEntity toEntity(FunctionMetrics fm, String username, LocalDateTime createdAt) {
        FunctionMetricsEntity e = new FunctionMetricsEntity();
        e.setUsername(username);
        e.setFunctionName(fm.getFunctionName());
        e.setLoc(fm.getLoc());
        e.setCyclomaticComplexity(fm.getCyclomaticComplexity());
        e.setParameterCount(fm.getParameterCount());
        e.setMaxNestingDepth(fm.getMaxNestingDepth());
        e.setCallCount(fm.getCallCount());
        e.setReturnCount(fm.getReturnCount());
        e.setGotoCount(fm.getGotoCount());
        e.setCreatedAt(createdAt);
        return e;
    }

    private FunctionMetrics toModel(FunctionMetricsEntity e) {
        FunctionMetrics fm = new FunctionMetrics();
        fm.setFunctionName(e.getFunctionName());
        fm.setLoc(e.getLoc());
        fm.setCyclomaticComplexity(e.getCyclomaticComplexity());
        fm.setParameterCount(e.getParameterCount());
        fm.setMaxNestingDepth(e.getMaxNestingDepth());
        fm.setCallCount(e.getCallCount());
        fm.setReturnCount(e.getReturnCount());
        fm.setGotoCount(e.getGotoCount());
        return fm;
    }
}