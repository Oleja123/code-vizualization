package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ReturnStmt implements Statement {
    @JsonProperty("value")
    private Expression value;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getValue() { return value; }
    public ASTLocation getLocation() { return location; }
}
