package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class UnaryExpr implements Expression {
    @JsonProperty("operator")
    private String op;

    @JsonProperty("operand")
    private Expression operand;

    @JsonProperty("isPostfix")
    private boolean isPostfix;

    @JsonProperty("location")
    private ASTLocation location;

    public String getOp() { return op; }
    public Expression getOperand() { return operand; }
    public boolean isPostfix() { return isPostfix; }
    public ASTLocation getLocation() { return location; }
}