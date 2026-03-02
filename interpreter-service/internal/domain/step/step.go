package step

import "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"

type Step struct {
	Events     []events.Event `json:"events"`
	StepNumber int            `json:"stepNumber"`
}
