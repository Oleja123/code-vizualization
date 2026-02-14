package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class VariableExpr implements Expression {
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public String getName() { return name; }
    public ASTLocation getLocation() { return location; }
}
