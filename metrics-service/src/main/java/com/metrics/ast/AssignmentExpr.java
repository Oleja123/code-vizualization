package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class AssignmentExpr implements Expression {
    private String type;
    private Expression left;
    private String operator;
    private Expression right;
    private ASTLocation location;
}
