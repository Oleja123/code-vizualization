package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ElseIfClause {
    @JsonProperty("condition")
    private Expression condition;
    
    @JsonProperty("block")
    private Statement block;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getCondition() { return condition; }
    public Statement getBlock() { return block; }
    public ASTLocation getLocation() { return location; }
}
