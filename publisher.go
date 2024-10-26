package queue

import "context"

type Publisher interface {
	Publish(ctx context.Context, queue string, content any) error
}
