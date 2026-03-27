package com.metrics.model;

import lombok.Data;

@Data
public class FunctionMetrics {
    private String functionName;
    private int loc;
    private int cyclomaticComplexity;
    private int parameterCount;
    private int maxNestingDepth;
    private int callCount;
    private int returnCount;
    private int gotoCount;
}
