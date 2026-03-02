package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;
import java.util.List;

/**
 * Тип данных
 */
class ASTType {
    @JsonProperty("baseType")
    private String baseType;
    
    @JsonProperty("pointerLevel")
    private int pointerLevel;
    
    @JsonProperty("arraySizes")
    private List<Integer> arraySizes;
    
    public String getBaseType() { return baseType; }
    public int getPointerLevel() { return pointerLevel; }
    public List<Integer> getArraySizes() { return arraySizes; }
    
    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder(baseType);
        for (int i = 0; i < pointerLevel; i++) sb.append("*");
        for (Integer size : arraySizes) sb.append("[").append(size).append("]");
        return sb.toString();
    }
}

// ============= Statements =============

class Parameter {
    @JsonProperty("type")
    private ASTType type;
    
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public ASTType getType() { return type; }
    public String getName() { return name; }
    public ASTLocation getLocation() { return location; }
}

// ============= Expressions =============

class ArrayInitExpr implements Expression {
    @JsonProperty("elements")
    private List<Expression> elements;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public List<Expression> getElements() { return elements; }
    public ASTLocation getLocation() { return location; }
}
