package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ArrayAccessExpr implements Expression {
    @JsonProperty("array")
    private Expression array;
    
    @JsonProperty("index")
    private Expression index;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getArray() { return array; }
    public Expression getIndex() { return index; }
    public ASTLocation getLocation() { return location; }
}
