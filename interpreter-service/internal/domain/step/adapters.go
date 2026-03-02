package step

import "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"

type StepDTO struct {
	Events     []events.EventDTO `json:"events"`
	StepNumber int               `json:"stepNumber"`
}

func MarshalStep(s Step) (StepDTO, error) {
	eventDTOs := make([]events.EventDTO, len(s.Events))
	for i, e := range s.Events {
		dto, err := events.MarshalEvent(e)
		if err != nil {
			return StepDTO{}, err
		}
		eventDTOs[i] = dto
	}

	return StepDTO{
		Events:     eventDTOs,
		StepNumber: s.StepNumber,
	}, nil
}

func UnmarshalStep(dto StepDTO) (Step, error) {
	eventsSlice := make([]events.Event, len(dto.Events))
	for i, eventDTO := range dto.Events {
		e, err := events.UnmarshalEvent(eventDTO)
		if err != nil {
			return Step{}, err
		}
		eventsSlice[i] = e
	}

	return Step{
		Events:     eventsSlice,
		StepNumber: dto.StepNumber,
	}, nil
}
