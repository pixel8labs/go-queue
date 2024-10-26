package redisqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/pixel8labs/logtrace/log"
	"github.com/pixel8labs/logtrace/trace"
)

// Publisher is a wrapper around asynq.Client to publish tasks.
// We use the term "queue" to refer to both task name and queue name, and use the same one for both.
type Publisher struct {
	asynqClient *asynq.Client
	config      PublisherConfig
}

type PublisherConfig struct {
	// MaxRetry is the maximum number of retry upon consume. Default is 5.
	MaxRetry int
}

func NewPublisher(asynqClient *asynq.Client, config PublisherConfig) *Publisher {
	return &Publisher{asynqClient: asynqClient, config: config}
}

// Publish publishes a message to the queue.
// content is expected to be a struct or map that can be marshalled to JSON.
func (p *Publisher) Publish(ctx context.Context, queue string, content any) error {
	// Put the content to the message struct so we can have headers object.
	msg := PublishedMessage{
		Content: content,
		Headers: p.constructHeaders(ctx),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json.Marshal msg: %w", err)
	}
	task := asynq.NewTask(queue, payload)

	logFields := log.Fields{
		"queue":   queue,
		"payload": string(payload),
	}
	log.Info(ctx, logFields, "Publishing to queue...")

	taskInfo, err := p.asynqClient.EnqueueContext(
		ctx,
		task,
		asynq.MaxRetry(p.config.MaxRetry),
		asynq.Queue(queue),
	)
	if err != nil {
		return fmt.Errorf("asynqClient.EnqueueContext: %w", err)
	}

	logFields["asynq_task_info"] = taskInfo
	log.Info(ctx, logFields, "Published to queue")

	return nil
}

func (p *Publisher) constructHeaders(ctx context.Context) Headers {
	traceHeader := make(map[string]string)

	trace.InjectTraceToMap(ctx, traceHeader)
	return Headers{
		Trace: traceHeader,
	}
}
