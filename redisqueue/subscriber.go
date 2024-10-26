package redisqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/pixel8labs/go-queue"
	"github.com/pixel8labs/logtrace/log"
	"github.com/pixel8labs/logtrace/trace"
)

// Subscriber is a wrapper around asynq.Client to subscribe.
type Subscriber struct {
	serveMux *asynq.ServeMux
	server   *asynq.Server
}

type SubscriberConfig struct {
	// Concurrency is the number of concurrent workers to use. Default is 10.
	Concurrency int
}

// NewSubscriber creates a new subscriber of a queue.
// asynqOpt is the asynq.RedisClientOpt to be used.
// queue is the queue name to subscribe to.
// config contains the optional configurations for the subscriber.
func NewSubscriber(
	asynqOpt asynq.RedisClientOpt,
	appName string,
	queue string,
	handler queue.SubscriberFn,
	config SubscriberConfig,
) *Subscriber {
	if config.Concurrency == 0 {
		config.Concurrency = 10
	}

	srv := asynq.NewServer(asynqOpt, asynq.Config{
		Concurrency: config.Concurrency, // Specify how many concurrent workers to use.
		Queues: map[string]int{
			queue: 1,
		},
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue, func(ctx context.Context, task *asynq.Task) error {
		logFields := log.Fields{
			"queue":   queue,
			"payload": string(task.Payload()),
		}
		log.Info(ctx, logFields, "Subscriber: Processing message")

		var msg Message
		if err := json.Unmarshal(task.Payload(), &msg); err != nil {
			log.Error(ctx, err, logFields, "Subscriber: Error when unmarshalling message")
			return fmt.Errorf("json.Unmarshal: %w", err)
		}

		// Propagate the trace.
		ctx = trace.ExtractTraceFromMap(ctx, msg.Headers.Trace)

		ctx, span := trace.StartSpan(ctx, appName+"-"+queue+"-worker", task.Type())
		defer span.End()

		if err := handler(ctx, msg.Content); err != nil {
			log.Error(ctx, err, logFields, "Subscriber: Error when processing message")
			return err
		}
		log.Info(ctx, logFields, "Subscriber: Processed message successfully")

		return nil
	})

	return &Subscriber{
		serveMux: mux,
		server:   srv,
	}
}

func (s *Subscriber) Start() error {
	return s.server.Run(s.serveMux)
}

func (s *Subscriber) Stop() {
	s.server.Shutdown()
}
