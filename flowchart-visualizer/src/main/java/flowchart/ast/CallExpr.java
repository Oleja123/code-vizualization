package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public class CallExpr implements Expression {
    @JsonProperty("function")
    private Expression function;
    
    @JsonProperty("arguments")
    private List<Expression> arguments;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getFunction() { return function; }
    public List<Expression> getArguments() { return arguments; }
    public ASTLocation getLocation() { return location; }
}
