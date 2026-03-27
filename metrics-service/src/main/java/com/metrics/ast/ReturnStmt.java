package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ReturnStmt implements Statement {
    private String type;
    private Expression value;
    private ASTLocation location;
}
