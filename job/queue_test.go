package job_test

import (
	"context"
	"testing"

	"github.com/pavelpuchok/vocabforge/db"
	"github.com/pavelpuchok/vocabforge/db/sqlc"
	"github.com/pavelpuchok/vocabforge/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestJobDescription struct {
	Value int
}

func TestQueue(t *testing.T) {
	sqlDB, err := db.NewSQLiteWithMigrations(":memory:")
	require.NoError(t, err)

	queue := job.Queue[TestJobDescription]{
		GroupName: "job_test.TestJobDescription",
		Storage: job.DBStorage{
			Queries: sqlc.New(sqlDB),
		},
	}

	ctx := context.Background()

	values := []TestJobDescription{
		{11},
		{22},
		{33},
		{44},
	}

	for _, v := range values {
		err := queue.Enqueue(ctx, v)
		assert.NoError(t, err)
	}

	for _, expected := range values {
		actual, err := queue.Dequeue(ctx)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	}
}
