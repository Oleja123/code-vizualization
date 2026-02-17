package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;

// Go шлёт: {"type":"CallExpr","functionName":"printf","arguments":[...]}
@JsonIgnoreProperties(ignoreUnknown = true)
public class CallExpr implements Expression {
    @JsonProperty("functionName")
    private String functionName;

    @JsonProperty("arguments")
    private List<Expression> arguments;

    @JsonProperty("location")
    private ASTLocation location;

    // Совместимость с FlowchartBuilder который вызывает getFunction()
    public Expression getFunction() {
        VariableExpr v = new VariableExpr();
        v.setName(functionName != null ? functionName : "");
        return v;
    }

    public String getFunctionName() { return functionName; }
    public List<Expression> getArguments() { return arguments; }
    public ASTLocation getLocation() { return location; }
}