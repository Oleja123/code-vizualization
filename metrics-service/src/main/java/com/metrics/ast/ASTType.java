package com.metrics.ast;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class ASTType {
    private String baseType;
    private int pointerLevel;
    private List<Integer> arraySizes;
}
