package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

public class DoWhileStmt implements Statement {

    @JsonProperty("condition")
    private Expression condition;

    @JsonProperty("body")
    private Statement body;

    @JsonProperty("location")
    private ASTLocation location;

    // Пустой конструктор нужен для Jackson
    public DoWhileStmt() {}

    public DoWhileStmt(Expression condition, Statement body, ASTLocation location) {
        this.condition = condition;
        this.body      = body;
        this.location  = location;
    }

    public Expression getCondition() { return condition; }
    public Statement  getBody()      { return body;      }

    @Override public ASTLocation getLocation() { return location; }
}