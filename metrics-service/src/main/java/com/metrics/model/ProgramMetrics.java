package com.metrics.model;

import lombok.Data;
import java.util.List;

@Data
public class ProgramMetrics {
    private int functionCount;
    private int globalVarCount;
    private List<FunctionMetrics> functions;
}
