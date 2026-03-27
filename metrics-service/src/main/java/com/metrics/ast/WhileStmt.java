package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class WhileStmt implements Statement {
    private String type;
    private Expression condition;
    private Statement body;
    private ASTLocation location;
}
