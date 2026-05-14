package worker

import (
	"context"
	"fmt"
	"github.com/user/streaming/internal/analytics"
	"github.com/user/streaming/internal/domain"
	"github.com/user/streaming/internal/redis"
)

type Pool struct {
	stream    *redis.RedisStream
	analytics *analytics.Engine
	numWorkers int
}

func NewPool(stream *redis.RedisStream, analytics *analytics.Engine, numWorkers int) *Pool {
	return &Pool{
		stream:     stream,
		analytics:  analytics,
		numWorkers: numWorkers,
	}
}

func (p *Pool) Start(ctx context.Context) {
	groupName := "analytics-group"
	p.stream.CreateGroup(ctx, groupName)

	// Launch multiple workers to process events concurrently
	// This helps in scaling when the event frequency is very high
	for i := 0; i < p.numWorkers; i++ {
		consumerName := fmt.Sprintf("consumer-%d", i)
		go p.stream.Consume(ctx, groupName, consumerName, func(event domain.Event) error {
			// In a real app, you might do database inserts or complex validation here
			p.analytics.ProcessEvent(event)
			return nil
		})
	}
}
