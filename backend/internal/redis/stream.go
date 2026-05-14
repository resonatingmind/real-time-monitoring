package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/user/streaming/internal/domain"
)

type RedisStream struct {
	client     *redis.Client
	streamName string
}

func NewRedisStream(client *redis.Client, streamName string) *RedisStream {
	return &RedisStream{
		client:     client,
		streamName: streamName,
	}
}

// Publish adds an event to the Redis Stream
func (s *RedisStream) Publish(ctx context.Context, event domain.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// XAdd appends the message to the stream
	// "*" tells Redis to generate an ID automatically
	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.streamName,
		Values: map[string]interface{}{"event": data},
	}).Err()
}

// CreateGroup initializes a consumer group if it doesn't exist
func (s *RedisStream) CreateGroup(ctx context.Context, group string) {
	s.client.XGroupCreateMkStream(ctx, s.streamName, group, "$")
}

// Consume reads messages from the stream using a consumer group
func (s *RedisStream) Consume(ctx context.Context, group, consumer string, handler func(domain.Event) error) {
	for {
		// XReadGroup blocks until a message is available
		entries, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{s.streamName, ">"}, // ">" means only new messages
			Count:    10,
			Block:    0, // block indefinitely
		}).Result()

		if err != nil {
			fmt.Printf("Error reading from stream: %v\n", err)
			continue
		}

		for _, stream := range entries {
			for _, message := range stream.Messages {
				var event domain.Event
				err := json.Unmarshal([]byte(message.Values["event"].(string)), &event)
				if err == nil {
					if err := handler(event); err == nil {
						// Acknowledge the message so it's not reprocessed
						s.client.XAck(ctx, s.streamName, group, message.ID)
					}
				}
			}
		}
	}
}
