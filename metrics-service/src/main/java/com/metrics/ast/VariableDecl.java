package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class VariableDecl implements Statement {
    private String type;
    private ASTType varType;
    private String name;
    private Expression initExpr;
    private ASTLocation location;
}
