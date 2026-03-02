package cache

import (
	"context"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
)

type CachedInfo struct {
	Value     []eventdispatcher.Step `json:"value"`
	StepBegin int                    `json:"stepBegin"`
	Result    *int                   `json:"result"`
	Err       error                  `json:"err,omitempty"`
}

type Cacher interface {
	Set(ctx context.Context, key string, value CachedInfo) error
	Get(ctx context.Context, key string) (CachedInfo, error)
}
