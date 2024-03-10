package db_test

import (
	"context"
	"testing"
	"time"

	"wonderful/internal/repository"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/test"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var usersRaw = []repository.User{
	{
		Name:         "Mr. John Doe",
		Email:        "john@xpto.com",
		Phone:        "123456789",
		Picture:      map[string]string{"url": "http://xpto.com/john.jpg"},
		Registration: time.Now().Add(-time.Hour * 48),
	},
	{
		Name:         "Mrs. Jane Doe",
		Email:        "jane@xpto.com",
		Phone:        "987654321",
		Picture:      map[string]string{"url": "http://xpto.com/jane.jpg"},
		Registration: time.Now().Add(-time.Hour * 24),
	},
	{
		Name:         "Mr. John Smith",
		Email:        "smith@xpto.com",
		Phone:        "123456789",
		Picture:      map[string]string{"url": "http://xpto.com/smith.jpg"},
		Registration: time.Now(),
	},
}

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

func (ts *UsersTestSuite) TestData() {
	ctx := context.Background()

	pool := ts.s.Pool()
	require.NotNil(ts.T(), pool)

	u := db.NewUserStorage(pool)

	// Find all users empty DB
	{
		p := repository.Params{}
		users, err := u.ListUsers(ctx, p)
		ts.Require().NoError(err)
		ts.Require().Len(users, 0)
	}

	// Insert users
	{
		err := u.Create(ctx, usersRaw)
		ts.Require().NoError(err)
	}

	// Find all users after insert
	{
		p := repository.Params{}
		users, err := u.ListUsers(ctx, p)
		ts.Require().NoError(err)
		ts.Require().Len(users, 3)
		// make sure the order is correct:
		ts.Require().Equal(users[0].Name, "Mr. John Smith")
		ts.Require().Equal(users[1].Name, "Mrs. Jane Doe")
		ts.Require().Equal(users[2].Name, "Mr. John Doe")
		ts.Require().Equal(users[2].Picture, usersRaw[0].Picture)
	}

	// Find the next and previous user
	{
		p := repository.Params{Limit: 1}
		users, err := u.ListUsers(ctx, p)
		ts.Require().NoError(err)
		ts.Require().Len(users, 1)
		ts.Require().Equal(users[0].Name, "Mr. John Smith")
		// Find the previous user
		p = repository.Params{Limit: 1, StartingAfter: &users[0].ID}
		users, err = u.ListUsers(ctx, p)
		ts.Require().NoError(err)
		ts.Require().Len(users, 1)
		ts.Require().Equal(users[0].Name, "Mrs. Jane Doe")
		// Find the next user
		p = repository.Params{Limit: 1, EndingBefore: &users[0].ID}
		users, err = u.ListUsers(ctx, p)
		ts.Require().NoError(err)
		ts.Require().Len(users, 1)
		ts.Require().Equal(users[0].Name, "Mr. John Smith")
	}
}
