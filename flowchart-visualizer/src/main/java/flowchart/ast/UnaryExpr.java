package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class UnaryExpr implements Expression {
    @JsonProperty("op")
    private String op;
    
    @JsonProperty("operand")
    private Expression operand;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public String getOp() { return op; }
    public Expression getOperand() { return operand; }
    public ASTLocation getLocation() { return location; }
}
