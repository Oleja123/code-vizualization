package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ContinueStmt implements Statement {
    private String type;
    private ASTLocation location;
}
