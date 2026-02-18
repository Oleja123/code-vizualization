package event_dispatcher

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/interpreter-service/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/snapshot"
)

type Step struct {
	Events     []events.Event
	StepNumber int
}

type EventDispatcher struct {
	Snapshot         *snapshot.Snapshot
	Steps            []Step
	currentStepIndex int
	pendingEvents    []events.Event
	stepInProgress   bool
}

func NewEventDispatcher(globalScope *runtime.Scope) *EventDispatcher {
	return &EventDispatcher{
		Snapshot:         snapshot.NewSnapshot(globalScope),
		Steps:            make([]Step, 0),
		currentStepIndex: -1,
		pendingEvents:    make([]events.Event, 0),
		stepInProgress:   false,
	}
}

func (ed *EventDispatcher) BeginStep() {
	if ed.stepInProgress {
		return
	}
	ed.pendingEvents = make([]events.Event, 0)
	ed.stepInProgress = true
}

func (ed *EventDispatcher) Emit(event events.Event) error {
	if !ed.stepInProgress {
		return fmt.Errorf("no step in progress, call BeginStep() first")
	}
	ed.pendingEvents = append(ed.pendingEvents, event)
	return nil
}

func (ed *EventDispatcher) EndStep() (Step, error) {
	if !ed.stepInProgress {
		return Step{}, fmt.Errorf("no step in progress")
	}

	stepNumber := len(ed.Steps)

	for _, event := range ed.pendingEvents {
		if err := ed.Snapshot.Apply(event, stepNumber); err != nil {
			return Step{}, err
		}
	}

	step := Step{
		Events:     ed.pendingEvents,
		StepNumber: stepNumber,
	}
	ed.Steps = append(ed.Steps, step)
	ed.currentStepIndex = stepNumber

	ed.pendingEvents = make([]events.Event, 0)
	ed.stepInProgress = false

	return step, nil
}

func (ed *EventDispatcher) ApplyStep(stepIndex int) error {
	if stepIndex < 0 || stepIndex >= len(ed.Steps) {
		return fmt.Errorf("invalid step index: %d (total steps: %d)", stepIndex, len(ed.Steps))
	}

	if stepIndex < ed.currentStepIndex {
		ed.Snapshot.Reset()
		ed.currentStepIndex = -1
	}

	startIndex := ed.currentStepIndex + 1
	for i := startIndex; i <= stepIndex; i++ {
		for _, event := range ed.Steps[i].Events {
			if err := ed.Snapshot.Apply(event, i); err != nil {
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

func (ed *EventDispatcher) IsStepInProgress() bool {
	return ed.stepInProgress
}

func (ed *EventDispatcher) GetPendingEventsCount() int {
	return len(ed.pendingEvents)
}
