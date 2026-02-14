package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ForStmt implements Statement {
    @JsonProperty("init")
    private Statement init;
    
    @JsonProperty("condition")
    private Expression condition;
    
    @JsonProperty("post")
    private Statement post;
    
    @JsonProperty("body")
    private Statement body;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Statement getInit() { return init; }
    public Expression getCondition() { return condition; }
    public Statement getPost() { return post; }
    public Statement getBody() { return body; }
    public ASTLocation getLocation() { return location; }
}
