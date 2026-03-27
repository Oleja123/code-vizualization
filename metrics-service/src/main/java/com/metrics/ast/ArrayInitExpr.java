package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;
import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ArrayInitExpr implements Expression {
    private String type;
    private List<Expression> elements;
    private ASTLocation location;
}