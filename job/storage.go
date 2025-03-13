package job

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pavelpuchok/vocabforge/db/sqlc"
)

type DBStorage struct {
	Queries *sqlc.Queries
}

var ErrJobNotFound = errors.New("job not found")

func (s DBStorage) Push(ctx context.Context, group string, value []byte) error {
	err := s.Queries.CreateJob(ctx, sqlc.CreateJobParams{
		GroupName: group,
		Item:      string(value),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("job.DBStorage.Push failed to create job. %w", err)
	}
	return nil
}

func (s DBStorage) Pop(ctx context.Context, group string) ([]byte, error) {
	job, err := s.Queries.GetOldestJobByGroup(ctx, group)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("job.DBStorage.Pop failed to get job. %w", err)
	}

	err = s.Queries.DeleteJob(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("job.DBStorage.Pop failed to delete job after fetching. %w", err)
	}

	return []byte(job.Item), nil
}
