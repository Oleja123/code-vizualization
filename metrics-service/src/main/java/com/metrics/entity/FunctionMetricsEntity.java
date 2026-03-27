package com.metrics.entity;

import jakarta.persistence.*;
import lombok.Data;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "function_metrics")
public class FunctionMetricsEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    // Имя пользователя из auth-service (из /api/auth/me)
    @Column(name = "username", nullable = false)
    private String username;

    @Column(name = "function_name", nullable = false)
    private String functionName;

    @Column(name = "loc")
    private int loc;

    @Column(name = "cyclomatic_complexity")
    private int cyclomaticComplexity;

    @Column(name = "parameter_count")
    private int parameterCount;

    @Column(name = "max_nesting_depth")
    private int maxNestingDepth;

    @Column(name = "call_count")
    private int callCount;

    @Column(name = "return_count")
    private int returnCount;

    @Column(name = "goto_count")
    private int gotoCount;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;
}
