package cache

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
)

type CachedInfoDTO struct {
	Value     []eventdispatcher.StepDTO `json:"value"`
	StepBegin int                       `json:"stepBegin"`
	Result    *int                      `json:"result"`
	Err       string                    `json:"err,omitempty"`
}

func MarshalCachedInfo(c CachedInfo) (CachedInfoDTO, error) {
	stepDTOs := make([]eventdispatcher.StepDTO, len(c.Value))
	for i, s := range c.Value {
		dto, err := eventdispatcher.MarshalStep(s)
		if err != nil {
			return CachedInfoDTO{}, err
		}
		stepDTOs[i] = dto
	}

	var errStr string
	if c.Err != nil {
		errStr = c.Err.Error()
	}

	return CachedInfoDTO{
		Value:     stepDTOs,
		StepBegin: c.StepBegin,
		Result:    c.Result,
		Err:       errStr,
	}, nil
}

func UnmarshalCachedInfo(dto CachedInfoDTO) (CachedInfo, error) {
	steps := make([]eventdispatcher.Step, len(dto.Value))
	for i, stepDTO := range dto.Value {
		s, err := eventdispatcher.UnmarshalStep(stepDTO)
		if err != nil {
			return CachedInfo{}, err
		}
		steps[i] = s
	}

	var err error
	if dto.Err != "" {
		err = fmt.Errorf(dto.Err)
	}

	return CachedInfo{
		Value:     steps,
		StepBegin: dto.StepBegin,
		Result:    dto.Result,
		Err:       err,
	}, nil
}
