package queue

import "context"

type SubscriberFn func(ctx context.Context, data []byte) error
