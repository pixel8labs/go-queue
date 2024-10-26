package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/pixel8labs/go-queue/redisqueue"
	"github.com/pixel8labs/logtrace/log"
	"github.com/pixel8labs/logtrace/trace"
	"github.com/redis/go-redis/v9"
)

const appName = "example-redis-queue"
const appEnv = "local"
const exampleQueueName = "example-queue"

type doSomethingMessage struct {
	DoWhat string `json:"do_what"`
}

func main() {
	// Load .env file.
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	// Init logger & tracer.
	log.Init(appName, appEnv, log.WithPrettyPrint())
	trace.InitTracer()

	// Init redis client.
	redisClient := initRedisClient(os.Getenv("REDIS_URL"))
	asynqClient, asynqOpt := initAsynqClient(redisClient)

	// Subscriber part.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		startSubscriber(ctx, asynqOpt)
		wg.Done()
	}()

	// Publisher part.
	queuePublisher := redisqueue.NewPublisher(
		asynqClient,
		redisqueue.PublisherConfig{MaxRetry: 5},
	)

	// Publish a message.
	message := doSomethingMessage{DoWhat: "Do something"}
	if err := queuePublisher.Publish(ctx, exampleQueueName, message); err != nil {
		log.Error(ctx, err, nil, "Failed to publish message")
	}

	// Sleep to wait for the subscriber to process the message.
	time.Sleep(5 * time.Second)

	cancel()

	wg.Wait()
}

func initRedisClient(redisUrl string) *redis.Client {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}
	opt.ReadTimeout = time.Second * 5
	opt.WriteTimeout = time.Second * 5

	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Panic(ctx, err, nil, "Failed to connect to Redis")
	}

	return client
}

func initAsynqClient(redisClient *redis.Client) (*asynq.Client, asynq.RedisClientOpt) {
	asynqRedisClientOpt := asynq.RedisClientOpt{
		Network:      redisClient.Options().Network,
		Addr:         redisClient.Options().Addr,
		Username:     redisClient.Options().Username,
		Password:     redisClient.Options().Password,
		DB:           redisClient.Options().DB,
		DialTimeout:  redisClient.Options().DialTimeout,
		ReadTimeout:  redisClient.Options().ReadTimeout,
		WriteTimeout: redisClient.Options().WriteTimeout,
		PoolSize:     redisClient.Options().PoolSize,
		TLSConfig:    redisClient.Options().TLSConfig,
	}

	return asynq.NewClient(asynqRedisClientOpt), asynqRedisClientOpt
}

func startSubscriber(ctx context.Context, asynqOpt asynq.RedisClientOpt) {
	subscriber := redisqueue.NewSubscriber(
		asynqOpt,
		appName,
		exampleQueueName,
		func(ctx context.Context, content json.RawMessage) error {
			// Your handler here.
			var payload doSomethingMessage
			if err := json.Unmarshal(content, &payload); err != nil {
				return fmt.Errorf("json.Unmarshal content: %w", err)
			}

			log.Info(ctx, log.Fields{"payload": payload}, "doSomething")

			return nil
		},
		redisqueue.SubscriberConfig{Concurrency: 10},
	)

	done := make(chan struct{})
	log.Info(ctx, nil, "Starting worker...")
	go func() {
		if err := subscriber.Start(); err != nil {
			log.Panic(ctx, err, nil, "Run worker failed")
		}
		done <- struct{}{}
	}()
	<-ctx.Done()

	log.Info(ctx, nil, "Shutting down worker...")
	subscriber.Stop()
	<-done
	log.Info(ctx, nil, "Worker shut down")
}
