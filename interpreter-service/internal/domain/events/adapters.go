package events

import (
	"encoding/json"
	"fmt"
)

type EventDTO struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func MarshalEvent(e Event) (EventDTO, error) {
	var typeStr string
	var data interface{}

	switch v := e.(type) {
	case EnterScope:
		typeStr = "EnterScope"
		data = v
	case ExitScope:
		typeStr = "ExitScope"
		data = v
	case DeclareVar:
		typeStr = "DeclareVar"
		data = v
	case DeclareArray:
		typeStr = "DeclareArray"
		data = v
	case DeclareArray2D:
		typeStr = "DeclareArray2D"
		data = v
	case VarChanged:
		typeStr = "VarChanged"
		data = v
	case ArrayElementChanged:
		typeStr = "ArrayElementChanged"
		data = v
	case Array2DElementChanged:
		typeStr = "Array2DElementChanged"
		data = v
	case FunctionCall:
		typeStr = "FunctionCall"
		data = v
	case FunctionReturn:
		typeStr = "FunctionReturn"
		data = v
	case LineChanged:
		typeStr = "LineChanged"
		data = v
	case UndefinedBehavior:
		typeStr = "UndefinedBehavior"
		data = v
	case RuntimeError:
		typeStr = "RuntimeError"
		data = v
	default:
		return EventDTO{}, fmt.Errorf("unknown event type: %T", e)
	}

	rawData, err := json.Marshal(data)
	if err != nil {
		return EventDTO{}, err
	}

	return EventDTO{
		Type: typeStr,
		Data: rawData,
	}, nil
}

func UnmarshalEvent(dto EventDTO) (Event, error) {
	switch dto.Type {
	case "EnterScope":
		var e EnterScope
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "ExitScope":
		var e ExitScope
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "DeclareVar":
		var e DeclareVar
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "DeclareArray":
		var e DeclareArray
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "DeclareArray2D":
		var e DeclareArray2D
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "VarChanged":
		var e VarChanged
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "ArrayElementChanged":
		var e ArrayElementChanged
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "Array2DElementChanged":
		var e Array2DElementChanged
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "FunctionCall":
		var e FunctionCall
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "FunctionReturn":
		var e FunctionReturn
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "LineChanged":
		var e LineChanged
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "UndefinedBehavior":
		var e UndefinedBehavior
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "RuntimeError":
		var e RuntimeError
		if err := json.Unmarshal(dto.Data, &e); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", dto.Type)
	}
}
