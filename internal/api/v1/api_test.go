package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"wonderful/internal/api/testhelpers"
	api "wonderful/internal/api/v1"
	"wonderful/internal/api/v1/openapi"
	authmiddleware "wonderful/internal/middleware"
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

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", testhelpers.TestJWTSecret)

	ts.container, err = test.SetupDB(ctx)
	require.NoError(ts.T(), err)
	ts.s, err = db.NewStorage(ctx)
	require.NoError(ts.T(), err)

	s := store.NewPersistentStore(ts.s.Pool())
	c := http.Client{}
	su := service.NewUserService(s, c)
	sh := service.NewHealthService(s)

	// set up our API
	wonderfulAPI := api.New(su, sh)
	r := chi.NewRouter()
	swagger, err := openapi.GetSwagger()
	require.NoError(ts.T(), err)
	r.Use(middleware.OapiRequestValidator(swagger))

	// Apply JWT authentication middleware
	r.Use(authmiddleware.JWTAuth)

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

	// Test constants
	const testUserID = "test-user-123"
	const testEmail = "test@example.com"

	// Test that requests without JWT token are rejected
	var errorResponse openapi.Error
	statusCode, err := testhelpers.Get(ctx, ts.server.URL+"/wonderfuls", &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusUnauthorized, statusCode)
	ts.Require().Equal(int32(401), errorResponse.Code)

	// Test with valid JWT token - should get empty list initially
	var response []openapi.User
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)

	// Populate the database
	var responseEmpty struct{}
	statusCode, err = testhelpers.PostWithJWT(ctx, ts.server.URL+"/populate", testUserID, testEmail, "", &responseEmpty)
	ts.Require().NoError(err)
	ts.Require().Equal(201, statusCode)

	// Get default number of users (should be 20 according to API spec)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 20) // Updated to match API spec default

	// Test boundary limits
	// Test minimum limit (1)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=1", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 1)

	// Test maximum limit (100)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=100", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 100)

	// Test moderate limit (5)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=5", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 5)

	// Get 50 users for pagination testing
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=50", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 50)
	ts.Require().Greater(response[0].RegistrationDate, response[49].RegistrationDate)

	// invalid limit
	errorResponse = openapi.Error{} // Reset error response
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=0", testUserID, testEmail, &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid limit: limit must be between 1 and 100", errorResponse.Message)

	errorResponse = openapi.Error{} // Reset error response
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=101", testUserID, testEmail, &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid limit: limit must be between 1 and 100", errorResponse.Message)

	errorResponse = openapi.Error{} // Reset error response
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=150", testUserID, testEmail, &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid limit: limit must be between 1 and 100", errorResponse.Message)

	// starting_after and ending_before
	errorResponse = openapi.Error{} // Reset error response
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?starting_after=1&ending_before=2", testUserID, testEmail, &errorResponse)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusBadRequest, statusCode)
	ts.Require().Equal("invalid startingAfter and endingBefore: only one of them can be used", errorResponse.Message)

	// starting_after
	var response2ndPage []openapi.User
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=50&starting_after="+response[49].Id, testUserID, testEmail, &response2ndPage)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response2ndPage, 50)
	ts.Require().Greater(response2ndPage[0].RegistrationDate, response2ndPage[49].RegistrationDate)
	for _, u := range response {
		require.NotContains(ts.T(), response2ndPage, u)
	}

	// ending_before
	var response1stPage []openapi.User
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?limit=50&ending_before="+response2ndPage[0].Id, testUserID, testEmail, &response1stPage)
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
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?email="+response[0].Email, testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 1)

	// email not found
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?email=notfound", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)

	// partial email
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?email="+response2ndPage[0].Email[:5], testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Greater(len(response), 0)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?email="+response2ndPage[0].Email[5:], testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Greater(len(response), 0)

	// SQL injection and make sure the database is not affected
	// '; DROP TABLE users; --
	s := "%27%3B%20DROP%20TABLE%20users%3B%20--"
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls?email="+s, testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 0)
	statusCode, err = testhelpers.GetWithJWT(ctx, ts.server.URL+"/wonderfuls", testUserID, testEmail, &response)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, statusCode)
	ts.Require().Len(response, 20)
}
