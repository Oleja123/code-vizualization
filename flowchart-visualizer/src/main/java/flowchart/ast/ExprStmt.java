package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ExprStmt implements Statement {
    @JsonProperty("expression")
    private Expression expression;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getExpression() { return expression; }
    public ASTLocation getLocation() { return location; }
}
