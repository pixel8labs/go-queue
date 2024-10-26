package queue

import (
	"context"
	"encoding/json"
)

type SubscriberFn func(ctx context.Context, data json.RawMessage) error
