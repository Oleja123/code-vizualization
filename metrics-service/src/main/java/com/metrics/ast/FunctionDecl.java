package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class FunctionDecl implements Statement {
    private String type;
    private String name;
    private ASTType returnType;
    private List<ASTNodes> parameters;
    private BlockStmt body;
    private ASTLocation location;
}
