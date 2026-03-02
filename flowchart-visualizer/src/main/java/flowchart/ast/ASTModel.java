package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;
import java.util.List;

// ============= Statements =============

// ============= Expressions =============

class ArrayInitExpr implements Expression {
    @JsonProperty("elements")
    private List<Expression> elements;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public List<Expression> getElements() { return elements; }
    public ASTLocation getLocation() { return location; }
}
