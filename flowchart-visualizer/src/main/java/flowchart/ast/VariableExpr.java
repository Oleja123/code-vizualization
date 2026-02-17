package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class VariableExpr implements Expression {
    @JsonProperty("name")
    private String name;

    @JsonProperty("location")
    private ASTLocation location;

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public ASTLocation getLocation() { return location; }
}