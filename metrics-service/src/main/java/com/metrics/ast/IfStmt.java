package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class IfStmt implements Statement {
    private String type;
    private Expression condition;
    private Statement thenBlock;
    private Statement elseBlock;
    private ASTLocation location;
}
