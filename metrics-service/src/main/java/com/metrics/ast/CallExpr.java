package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class CallExpr implements Expression {
    private String type;
    private String functionName;
    private List<Expression> arguments;
    private ASTLocation location;
}
