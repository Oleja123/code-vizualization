package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class GotoStmt implements Statement {
    private String type;
    private String label;
    private ASTLocation location;
}
