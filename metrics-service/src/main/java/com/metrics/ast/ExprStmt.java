package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ExprStmt implements Statement {
    private String type;
    private Expression expression;
    private ASTLocation location;
}
