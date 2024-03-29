package db_test

import (
	"context"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"

	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/test"
)

type PostgresTestSuite struct {
	suite.Suite
	container *testcontainers.PostgresContainer
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (ts *PostgresTestSuite) SetupTest() {
	var err error
	ctx := context.Background()
	ts.container, err = test.SetupDB(ctx)
	require.NoError(ts.T(), err)
}

func (ts *PostgresTestSuite) TearDownTest() {
	ctx := context.Background()
	err := test.TeardownDB(ctx, ts.container)
	require.NoError(ts.T(), err)
}

// This tests only that the connection to the DB is working and also the migrations.
func (ts *PostgresTestSuite) TestPostgres() {
	ctx := context.Background()
	s, err := db.NewStorage(ctx)
	require.NoError(ts.T(), err)

	pool := s.Pool()
	require.NotNil(ts.T(), pool)
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	require.NoError(ts.T(), err)
	require.NotNil(ts.T(), conn)
	defer conn.Release()

	err = conn.Ping(ctx)
	require.NoError(ts.T(), err)

	// check the existing tables
	rows, err := conn.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), rows.Err())
	defer rows.Close()
	var tables []string //nolint:prealloc // this is a test
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		require.NoError(ts.T(), err)
		tables = append(tables, table)
	}
	require.Contains(ts.T(), tables, "schema_migrations")
	require.Contains(ts.T(), tables, "users")
}
