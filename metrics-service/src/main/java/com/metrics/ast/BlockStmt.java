package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class BlockStmt implements Statement {
    private String type;
    private List<Statement> statements;
    private ASTLocation location;
}
