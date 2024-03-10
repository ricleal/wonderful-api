package store_test

import (
	"context"
	"testing"
	"time"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/test"
	"wonderful/internal/store"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type StoreTestSuite struct {
	suite.Suite
	container *testcontainers.PostgresContainer
	s         *db.Storage
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (ts *StoreTestSuite) SetupSuite() {
	var err error
	ctx := context.Background()
	ts.container, err = test.SetupDB(ctx)
	require.NoError(ts.T(), err)
	ts.s, err = db.NewStorage(ctx)
	require.NoError(ts.T(), err)
}

func (ts *StoreTestSuite) TearDownSuite() {
	ctx := context.Background()
	err := test.TeardownDB(ctx, ts.container)
	require.NoError(ts.T(), err)
	ts.s.Close()
}

func (ts *StoreTestSuite) TestStoreOK() {
	ctx := context.Background()
	s := store.NewPersistentStore(ts.s.Pool())

	usersStore := s.Users()

	users, err := usersStore.ListUsers(ctx, repository.Params{})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 0)

	err = usersStore.Create(ctx, []repository.User{
		{
			Name:         "Mr. John Doe",
			Email:        "john@test.com",
			Phone:        "123456789",
			Picture:      map[string]string{"url": "http://xpto.com/john.jpg"},
			Registration: time.Now(),
		},
	})
	require.NoError(ts.T(), err)

	users, err = usersStore.ListUsers(ctx, repository.Params{})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 1)

	// Delete the table to reset the state of the database.
	_, err = ts.s.Pool().Exec(ctx, "DELETE FROM users")
	require.NoError(ts.T(), err)
}

func (ts *StoreTestSuite) TestStoreExecTxOK() {
	ctx := context.Background()
	s := store.NewPersistentStore(ts.s.Pool())

	err := s.ExecTx(ctx, func(st store.Store) error {
		usersStore := st.Users()
		err := usersStore.Create(ctx, []repository.User{
			{
				Name:         "Mr. John Doe",
				Email:        "john@test.com",
				Phone:        "123456789",
				Picture:      map[string]string{"url": "http://xpto.com/john.jpg"},
				Registration: time.Now(),
			},
		})
		require.NoError(ts.T(), err)

		users, err := usersStore.ListUsers(ctx, repository.Params{})
		require.NoError(ts.T(), err)
		require.Len(ts.T(), users, 1)

		return nil
	})
	require.NoError(ts.T(), err)

	// Delete the table to reset the state of the database.
	_, err = ts.s.Pool().Exec(ctx, "DELETE FROM users")
	require.NoError(ts.T(), err)
}
