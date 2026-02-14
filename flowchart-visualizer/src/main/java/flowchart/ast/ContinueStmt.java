package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ContinueStmt implements Statement {
    @JsonProperty("location")
    private ASTLocation location;
    
    public ASTLocation getLocation() { return location; }
}
