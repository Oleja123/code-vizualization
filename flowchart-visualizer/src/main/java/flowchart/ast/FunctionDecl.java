package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public class FunctionDecl implements Statement {
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("returnType")
    private ASTType returnType;
    
    @JsonProperty("parameters")
    private List<Parameter> parameters;
    
    @JsonProperty("body")
    private BlockStmt body;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public String getName() { return name; }
    public ASTType getReturnType() { return returnType; }
    public List<Parameter> getParameters() { return parameters; }
    public BlockStmt getBody() { return body; }
    public ASTLocation getLocation() { return location; }
}
