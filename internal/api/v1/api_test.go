package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"wonderful/internal/api/testhelpers"
	api "wonderful/internal/api/v1"
	"wonderful/internal/api/v1/openapi"
	"wonderful/internal/repository/db"
	"wonderful/internal/repository/db/test"
	"wonderful/internal/service"
	"wonderful/internal/store"

	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type APITestIntegrationSuite struct {
	suite.Suite
	container *testcontainers.PostgresContainer
	s         *db.Storage
	server    *httptest.Server
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestAPITestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(APITestIntegrationSuite))
}

func (ts *APITestIntegrationSuite) SetupSuite() {
	var err error
	ctx := context.Background()
	ts.container, err = test.SetupDB(ctx)
	require.NoError(ts.T(), err)
	ts.s, err = db.NewStorage(ctx)
	require.NoError(ts.T(), err)

	s := store.NewPersistentStore(ts.s.Pool())
	c := http.Client{}
	su := service.NewUserService(s, c)

	// set up our API
	wonderfulAPI := api.New(su)
	r := chi.NewRouter()
	swagger, err := openapi.GetSwagger()
	require.NoError(ts.T(), err)
	r.Use(middleware.OapiRequestValidator(swagger))
	openapi.HandlerFromMux(wonderfulAPI, r)
	ts.server = httptest.NewServer(r)
}

func (ts *APITestIntegrationSuite) TearDownSuite() {
	ctx := context.Background()
	err := test.TeardownDB(ctx, ts.container)
	require.NoError(ts.T(), err)
	ts.s.Close()
}

func (ts *APITestIntegrationSuite) TestUsers() {
	ctx := context.Background()
	var response []openapi.User

	statusCode, err := testhelpers.Get(ctx, ts.server.URL+"/wonderfuls", &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)

	// Populate the database
	var responseEmpty struct{}
	statusCode, err = testhelpers.Post(ctx, ts.server.URL+"/populate", "", &responseEmpty)
	ts.Require().NoError(err)
	ts.Require().Equal(201, statusCode)

	// Get default number of users
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls", &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 10)

	// Get 50 users
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?limit=50", &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 50)
	ts.Require().Greater(response[0].RegistrationDate, response[49].RegistrationDate)

	// invalid limit
	var errorResponse openapi.Error
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?limit=0", &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid limit: limit must be between 1 and 100", errorResponse.Message)
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?limit=150", &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid limit: limit must be between 1 and 100", errorResponse.Message)

	// starting_after and ending_before
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?starting_after=1&ending_before=2", &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid startingAfter and endingBefore: only one of them can be used", errorResponse.Message)

	// starting_after
	var response2ndPage []openapi.User
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?limit=50&starting_after="+response[49].Id, &response2ndPage)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response2ndPage, 50)
	ts.Require().Greater(response2ndPage[0].RegistrationDate, response2ndPage[49].RegistrationDate)
	for _, u := range response {
		require.NotContains(ts.T(), response2ndPage, u)
	}

	// ending_before
	var response1stPage []openapi.User
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?limit=50&ending_before="+response2ndPage[0].Id, &response1stPage)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response1stPage, 50)
	for _, u := range response2ndPage {
		require.NotContains(ts.T(), response1stPage, u)
	}
	for i, u := range response {
		require.Equal(ts.T(), u, response1stPage[i])
	}

	// email
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?email="+response[0].Email, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 1)

	// email not found
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?email=notfound", &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)

	// partial email
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?email="+response2ndPage[0].Email[:5], &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Greater(len(response), 0)
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?email="+response2ndPage[0].Email[5:], &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Greater(len(response), 0)

	// SQL injection and make sure the database is not affected
	// '; DROP TABLE users; --
	s := "%27%3B%20DROP%20TABLE%20users%3B%20--"
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls?email="+s, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)
	statusCode, err = testhelpers.Get(ctx, ts.server.URL+"/wonderfuls", &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 10)
}
