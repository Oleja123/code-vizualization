package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class UnaryExpr implements Expression {
    private String type;
    private String operator;
    private Expression operand;
    private boolean isPostfix;
    private ASTLocation location;
}
