package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class IntLiteral implements Expression {
    private String type;
    private int value;
    private ASTLocation location;
}
