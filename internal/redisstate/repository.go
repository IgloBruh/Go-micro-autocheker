package redisstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"autocheck-microservices/internal/contracts"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Client *redis.Client
	Prefix string
}

func New(client *redis.Client, prefix string) *Repository {
	return &Repository{Client: client, Prefix: prefix}
}

func (r *Repository) key(id string) string {
	return r.Prefix + id
}

func (r *Repository) Save(ctx context.Context, state contracts.PersistedRunState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, r.key(state.Id), data, 24*time.Hour).Err()
}

func (r *Repository) Get(ctx context.Context, id string) (contracts.PersistedRunState, error) {
	value, err := r.Client.Get(ctx, r.key(id)).Result()
	if err != nil {
		if err == redis.Nil {
			return contracts.PersistedRunState{}, fmt.Errorf("run %s not found", id)
		}
		return contracts.PersistedRunState{}, err
	}
	var state contracts.PersistedRunState
	if err = json.Unmarshal([]byte(value), &state); err != nil {
		return contracts.PersistedRunState{}, err
	}
	return state, nil
}
