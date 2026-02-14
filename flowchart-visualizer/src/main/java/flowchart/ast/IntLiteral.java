package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class IntLiteral implements Expression {
    @JsonProperty("value")
    private int value;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public int getValue() { return value; }
    public ASTLocation getLocation() { return location; }
}
