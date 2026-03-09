package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	QueueImport     = "wms:queue:import"
	QueueBulkUpdate = "wms:queue:bulk_update"
)

type Job struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type RedisQueue struct {
	Client *redis.Client
}

func NewRedisQueue() (*RedisQueue, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisQueue{Client: client}, nil
}

func (q *RedisQueue) Enqueue(ctx context.Context, queueName string, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return q.Client.LPush(ctx, queueName, data).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*Job, error) {
	result, err := q.Client.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return nil, fmt.Errorf("dequeue: %w", err)
	}
	if len(result) < 2 {
		return nil, fmt.Errorf("empty result")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}
	return &job, nil
}

func (q *RedisQueue) Close() {
	q.Client.Close()
}
