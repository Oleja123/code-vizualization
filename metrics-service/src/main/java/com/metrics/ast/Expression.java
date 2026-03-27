package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, property = "type")
@JsonSubTypes({
        @JsonSubTypes.Type(value = BinaryExpr.class,     name = "BinaryExpr"),
        @JsonSubTypes.Type(value = UnaryExpr.class,      name = "UnaryExpr"),
        @JsonSubTypes.Type(value = AssignmentExpr.class, name = "AssignmentExpr"),
        @JsonSubTypes.Type(value = CallExpr.class,       name = "CallExpr"),
        @JsonSubTypes.Type(value = VariableExpr.class,   name = "VariableExpr"),
        @JsonSubTypes.Type(value = IntLiteral.class,     name = "IntLiteral"),
        @JsonSubTypes.Type(value = ArrayAccessExpr.class,name = "ArrayAccessExpr"),
        @JsonSubTypes.Type(value = ArrayInitExpr.class,   name = "ArrayInitExpr"),
})
@JsonIgnoreProperties(ignoreUnknown = true)
public interface Expression {}