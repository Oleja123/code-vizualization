package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class BinaryExpr implements Expression {
    @JsonProperty("operator")
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