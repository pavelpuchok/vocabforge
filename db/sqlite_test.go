package db

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSQLiteWithMigrations(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "db.sqlite")

	sqlDb, err := NewSQLiteWithMigrations(dbPath)
	require.NoError(t, err, "unexpected fail on empty db opening")
	sqlDb.Close()

	sqlDb, err = NewSQLiteWithMigrations(dbPath)
	require.NoError(t, err, "unexpected fail on non empty db opening")
	sqlDb.Close()
}
