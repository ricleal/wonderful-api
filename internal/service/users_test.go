package service_test

import (
	"context"
	"net/http"
	"testing"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/test"
	"wonderful/internal/service"
	"wonderful/internal/store"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

//nolint:lll // this is a test file
var insertStatement = `INSERT INTO users (id, name, email, phone, cell, picture, registration) VALUES 
('0ujsszwN8NRY24YaXiTIE2VWDT0', 'Mr. John Doe', 'jd@mail.com', '123-456-7890', '123-456-7890', '{"large": "http://example.com/large.jpg", "medium": "http://example.com/medium.jpg", "thumbnail": "http://example.com/thumbnail.jpg"}', '2021-01-01T00:00:00Z'),
('0ujsszwN8NRY24YaXiTIE2VWDT1', 'Mrs. Jane Doe1', 'jane@mail.com', '123-456-7890', '123-456-7890', '{"large": "http://example.com/large.jpg", "medium": "http://example.com/medium.jpg", "thumbnail": "http://example.com/thumbnail.jpg"}', '2022-01-01T00:00:00Z'),
('0ujsszwN8NRY24YaXiTIE2VWDT2', 'Mrs. Jane Doe2', 'jane@mail.com', '123-456-7890', '123-456-7890', '{"large": "http://example.com/large.jpg", "medium": "http://example.com/medium.jpg", "thumbnail": "http://example.com/thumbnail.jpg"}', '2022-01-01T00:00:00Z'),
('0ujsszwN8NRY24YaXiTIE2VWDT3', 'Mr. John Smith', 'js@mail.com', '123-456-7890', '123-456-7890', '{"large": "http://example.com/large.jpg", "medium": "http://example.com/medium.jpg", "thumbnail": "http://example.com/thumbnail.jpg"}', '2023-01-01T00:00:00Z')`

var deleteStatement = `DELETE FROM users`

type UsersTestSuite struct {
	suite.Suite
	container *testcontainers.PostgresContainer
	s         *db.Storage
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestUsersTestSuite(t *testing.T) {
	suite.Run(t, new(UsersTestSuite))
}

func (ts *UsersTestSuite) SetupSuite() {
	var err error
	ctx := context.Background()
	ts.container, err = test.SetupDB(ctx)
	require.NoError(ts.T(), err)
	ts.s, err = db.NewStorage(ctx)
	require.NoError(ts.T(), err)
}

func (ts *UsersTestSuite) TearDownSuite() {
	ctx := context.Background()
	err := test.TeardownDB(ctx, ts.container)
	require.NoError(ts.T(), err)
	ts.s.Close()
}

func (ts *UsersTestSuite) TestBulkLoad() {
	s := store.NewPersistentStore(ts.s.Pool())
	c := http.Client{}
	su := service.NewUserService(s, c)
	ctx := context.Background()

	// get all users empty DB
	err := su.Create(ctx)
	ts.Require().NoError(err)

	_, err = ts.s.Pool().Exec(ctx, deleteStatement)
	require.NoError(ts.T(), err)
}

func (ts *UsersTestSuite) TestListUsers() {
	// insert some users
	ctx := context.Background()
	_, err := ts.s.Pool().Exec(ctx, insertStatement)
	require.NoError(ts.T(), err)

	s := store.NewPersistentStore(ts.s.Pool())
	c := http.Client{}
	su := service.NewUserService(s, c)

	// get all users with limit
	users, err := su.ListUsers(ctx, repository.Params{Limit: 2})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 2)

	// get all users
	users, err = su.ListUsers(ctx, repository.Params{})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 4)
	// Make sure the users are sorted by registration date
	require.Equal(ts.T(), "Mr. John Smith", users[0].Name)
	require.Equal(ts.T(), "Mrs. Jane Doe2", users[1].Name)
	require.Equal(ts.T(), "Mrs. Jane Doe1", users[2].Name)
	require.Equal(ts.T(), "Mr. John Doe", users[3].Name)
	janeDoe1ID, err := ksuid.Parse(users[2].ID)
	require.NoError(ts.T(), err)
	// get all users starting after Jane Doe1
	users, err = su.ListUsers(ctx, repository.Params{StartingAfter: &janeDoe1ID})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 1)
	require.Equal(ts.T(), "Mr. John Doe", users[0].Name)
	// get all users ending before Jane
	users, err = su.ListUsers(ctx, repository.Params{EndingBefore: &janeDoe1ID})
	require.NoError(ts.T(), err)
	require.Len(ts.T(), users, 2)
	require.Equal(ts.T(), "Mr. John Smith", users[0].Name)
	require.Equal(ts.T(), "Mrs. Jane Doe2", users[1].Name)
}
