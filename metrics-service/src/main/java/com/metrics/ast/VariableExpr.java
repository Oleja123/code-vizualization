package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class VariableExpr implements Expression {
    private String type;
    private String name;
    private ASTLocation location;
}
