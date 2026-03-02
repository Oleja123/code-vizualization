package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Cacher = &RedisCacher{}

type RedisCacher struct {
	client     *redis.Client
	expiration time.Duration
}

func NewRedisCacher(client *redis.Client, expiration time.Duration) *RedisCacher {
	return &RedisCacher{
		client:     client,
		expiration: expiration,
	}
}

func (r *RedisCacher) Set(ctx context.Context, key string, value CachedInfo) error {
	dto, err := MarshalCachedInfo(value)
	if err != nil {
		return err
	}

	data, marshalErr := json.Marshal(dto)
	if marshalErr != nil {
		return marshalErr
	}

	return r.client.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisCacher) Get(ctx context.Context, key string) (CachedInfo, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return CachedInfo{}, err
	}

	var dto CachedInfoDTO
	if err := json.Unmarshal([]byte(data), &dto); err != nil {
		return CachedInfo{}, err
	}

	return UnmarshalCachedInfo(dto)
}
