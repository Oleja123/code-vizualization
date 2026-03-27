package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ForStmt implements Statement {
    private String type;
    private Statement init;
    private Expression condition;
    private Statement post;
    private Statement body;
    private ASTLocation location;
}
