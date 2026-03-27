package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ArrayAccessExpr implements Expression {
    private String type;
    private Expression array;
    private Expression index;
    private ASTLocation location;
}
