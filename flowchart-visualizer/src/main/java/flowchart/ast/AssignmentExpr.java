package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class AssignmentExpr implements Expression {
    @JsonProperty("op")
    private String op;
    
    @JsonProperty("left")
    private Expression left;
    
    @JsonProperty("right")
    private Expression right;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public String getOp() { return op; }
    public Expression getLeft() { return left; }
    public Expression getRight() { return right; }
    public ASTLocation getLocation() { return location; }
}
