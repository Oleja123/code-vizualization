package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class VariableDecl implements Statement {
    @JsonProperty("varType")
    private ASTType varType;
    
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("initExpr")
    private Expression initExpr;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public ASTType getVarType() { return varType; }
    public String getName() { return name; }
    public Expression getInitExpr() { return initExpr; }
    public ASTLocation getLocation() { return location; }
}
