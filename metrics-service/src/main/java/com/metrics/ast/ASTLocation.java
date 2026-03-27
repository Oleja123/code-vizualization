package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ASTLocation {
    private int line;
    private int column;
    private int endLine;
    private int endColumn;
}
