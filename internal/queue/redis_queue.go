package queue

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	Client        *redis.Client
	QueueName     string
	EventsChannel string
}

func New(client *redis.Client, queueName, eventsChannel string) *RedisQueue {
	return &RedisQueue{Client: client, QueueName: queueName, EventsChannel: eventsChannel}
}

func (q *RedisQueue) EnqueueJob(ctx context.Context, payload []byte) error {
	return q.Client.LPush(ctx, q.QueueName, payload).Err()
}

func (q *RedisQueue) DequeueJob(ctx context.Context, timeout time.Duration) ([]byte, error) {
	res, err := q.Client.BRPop(ctx, timeout, q.QueueName).Result()
	if err != nil {
		return nil, err
	}
	if len(res) != 2 {
		return nil, redis.Nil
	}
	return []byte(res[1]), nil
}

func (q *RedisQueue) PublishEvent(ctx context.Context, payload []byte) error {
	return q.Client.Publish(ctx, q.EventsChannel, payload).Err()
}

func (q *RedisQueue) Subscribe(ctx context.Context) *redis.PubSub {
	return q.Client.Subscribe(ctx, q.EventsChannel)
}
