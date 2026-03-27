package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

// ── Program ───────────────────────────────────────────────────────────────────

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ASTNodes {
    private ASTType type;
    private String name;
    private ASTLocation location;
}

// ── Statements ────────────────────────────────────────────────────────────────

// ── Expressions ───────────────────────────────────────────────────────────────

