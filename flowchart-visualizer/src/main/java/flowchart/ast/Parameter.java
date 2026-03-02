package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Parameter {
    @JsonProperty("type")
    private ASTType type;
    
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public ASTType getType() { return type; }
    public String getName() { return name; }
    public ASTLocation getLocation() { return location; }
}
