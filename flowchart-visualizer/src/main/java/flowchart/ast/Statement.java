package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo; /**
 * Базовый интерфейс для операторов
 */
@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, property = "type")
@JsonSubTypes({
    @JsonSubTypes.Type(value = VariableDecl.class, name = "VariableDecl"),
    @JsonSubTypes.Type(value = FunctionDecl.class, name = "FunctionDecl"),
    @JsonSubTypes.Type(value = IfStmt.class, name = "IfStmt"),
    @JsonSubTypes.Type(value = WhileStmt.class, name = "WhileStmt"),
    @JsonSubTypes.Type(value = ForStmt.class, name = "ForStmt"),
    @JsonSubTypes.Type(value = ReturnStmt.class, name = "ReturnStmt"),
    @JsonSubTypes.Type(value = BlockStmt.class, name = "BlockStmt"),
    @JsonSubTypes.Type(value = ExprStmt.class, name = "ExprStmt"),
    @JsonSubTypes.Type(value = BreakStmt.class, name = "BreakStmt"),
    @JsonSubTypes.Type(value = ContinueStmt.class, name = "ContinueStmt")
})
public interface Statement {
    ASTLocation getLocation();
}
