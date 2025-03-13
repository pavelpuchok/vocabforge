package job

import (
	"context"
	"encoding/json"
	"fmt"
)

type Queue[T any] struct {
	GroupName string
	Storage   DBStorage
}

func (q Queue[T]) Enqueue(ctx context.Context, v T) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("job.Queue.Enqueue fail to marshal value. %w", err)
	}

	err = q.Storage.Push(ctx, q.GroupName, raw)
	if err != nil {
		return fmt.Errorf("job.Queue.Enqueue fail to insert value. %w", err)
	}

	return nil
}

func (q Queue[T]) Dequeue(ctx context.Context) (T, error) {
	var v T

	raw, err := q.Storage.Pop(ctx, q.GroupName)
	if err != nil {
		return v, fmt.Errorf("job.Queue.Dequeue fail to get value. %w", err)
	}

	err = json.Unmarshal(raw, &v)
	if err != nil {
		return v, fmt.Errorf("job.Queue.Dequeue fail to unmarshal job value. %w", err)
	}

	return v, nil
}
