package eventdispatcher

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/snapshot"
)

type Step struct {
	Events     []events.Event
	StepNumber int
}

type EventDispatcher struct {
	Snapshot         *snapshot.Snapshot
	Steps            []Step
	currentStepIndex int
	stepBegin        int
}

func NewEventDispatcher(stepBegin int) *EventDispatcher {
	return &EventDispatcher{
		Snapshot:         snapshot.NewSnapshot(),
		Steps:            make([]Step, 0),
		currentStepIndex: -1,
		stepBegin:        stepBegin,
	}
}

func (ed *EventDispatcher) ApplyStep(stepIndex int) error {
	stepIndex += ed.stepBegin

	if stepIndex < 0 || stepIndex >= len(ed.Steps) {
		return fmt.Errorf("invalid step index: %d (total steps: %d)", stepIndex, len(ed.Steps))
	}

	if stepIndex < ed.currentStepIndex {
		ed.Snapshot.Reset()
		ed.currentStepIndex = ed.stepBegin - 1
	}

	startIndex := ed.currentStepIndex + 1
	for i := startIndex; i <= stepIndex; i++ {
		for _, event := range ed.Steps[i].Events {
			if err := ed.Snapshot.Apply(event, i-ed.stepBegin); err != nil {
				return err
			}
		}
		ed.currentStepIndex = i
	}

	return nil
}

func (ed *EventDispatcher) GetCurrentStep() int {
	return ed.currentStepIndex
}

func (ed *EventDispatcher) GetStepsCount() int {
	return len(ed.Steps)
}

func (ed *EventDispatcher) GetSnapshot() *snapshot.Snapshot {
	return ed.Snapshot
}

func (ed *EventDispatcher) GetStep(index int) (Step, error) {
	if index < 0 || index >= len(ed.Steps) {
		return Step{}, fmt.Errorf("invalid step index: %d", index)
	}
	return ed.Steps[index], nil
}
