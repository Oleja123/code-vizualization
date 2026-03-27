package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class DoWhileStmt implements Statement {
    private String type;
    private Statement body;
    private Expression condition;
    private ASTLocation location;
}
