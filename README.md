# Wonderful API

Link to the [problem statement](docs/README.md).

## Architecture

### Structure

```bash
.
├── cmd
│   └── wonderful - Executable to start the API
├── docs - The problem statement
├── internal
│   ├── api - The API layer
│   │   ├── testhelpers - The test helpers (GET, POST, etc) for the API layer
│   │   └── v1 - The API V1 layer
│   │       └── openapi - The generated code from the OpenAPI spec
│   ├── entities - The data entities used in the business logic
│   ├── repository - The data layer
│   │   └── db - The Postgres database specific code
│   │       ├── sqlc - The auto generated code from the database schema and the SQL queries
│   │       └── test - The test containers auxiliary code
│   ├── service - The business logic
│   └── store - The store to chain multiple repository operations in a single transaction
├── load-test - The load test using Vegeta
├── migrations - The database migrations
└── open-api - The OpenAPI spec file
```

### Backend

I have used an architecture following the same principles as the [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Robert C. Martin. I have included the [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html) to abstract the data layer from the business logic. The packages worth mentioning are:

- `internal/repository/db/sqlc`: the auto generated code from the database schema and the SQL queries. This is generated using the [sqlc](https://sqlc.dev/). This package is used to interact with the database.
- `internal/repository/db`: the Postgres specific code. This package is used to connect to the database and execute the `sqlc` queries. If we want to have more databases, for example `mem`, we would have a `mem` package here: `internal/repository/mem`. Note that these packages should implement the same interfaces defined in `internal/repository/interfaces.go`.


To interact with this repositories, there is a [Store](internal/store). The final objective of this store is to chain multiple repository operations in a single transaction. It sort of follows the same principles as in the [Unit of Work](https://martinfowler.com/eaaCatalog/unitOfWork.html) pattern. The service layer uses the store to interact with the repositories. It never interacts directly with the repositories.

The business logic is implemented in the `internal/service` package. This represents the use cases of the application - the domain service. Note that the business logic always uses data entities defined in the `internal/entities` package. This ensures business logic is decoupled from the data layer defined in the `internal/repository` package.

### Frontend API

The use cases are exposed via the API layer. The API layer is implemented using the [go-chi](https://github.com/go-chi/chi) router. 

The endpoints routes were generated from a [Open-API](https://www.openapis.org/) [spec file](open-api/v1.yaml) using [oapi-codegen](https://github.com/deepmap/oapi-codegen). The code generated is in `internal/api/v1/openapi`.

The following endpoints are exposed:
```bash
# Get the API spec
GET /api/v1/api.json
# Get all users (see the problem statement for the query parameters)
GET /api/v1/wonderfuls
# Create users (copy users from the `https://randomuser.me/api/` endpoint and store them in the database)
POST /api/v1/populate
```

## Running the application

### Prerequisites
- docker
- docker-compose

### Running the application

Ideally, the application should be launched from the makefile. This makes sure `docker-compose` is run with the correct environment variables. Otherwise, set the environment variables defined in the `envrc-template` file.

To start the application, run:
```bash
make docker-up
```
It runs in attached mode, so you can see the logs of the application. 

When you are done, stop the application in another terminal with:

```bash
make docker-down
```

## Development

### Makefile targets

The `make help` command lists all the available targets.

```bash
# Open a Postgres CLI
db-cli                         Start the Postgres CLI
# Manage the database schema
db-migrate-create              Create a new migration file
db-migrate-down                Run database downgrade the last migration
db-migrate-force               Force mark the migration version
db-migrate-up                  Run database upgrade migrations
db-migrate-version             Print the current migration version
# Start / stop the database running in docker
db-start                       Postgres start
db-stop                        Postgres stop
# Development targets
dev                            Run development server
lint                           Lint and format source code based on golangci configuration
# Runs tests
test                           Run unit and integration tests
# Generate code
db-models                      Generate Go database models
openapi-generate               Generate OpenAPI client
# Starts the API in docker (starts the database and runs the migrations if needed)
docker-down                    Stop docker container
docker-up                      Run docker container
```


## Tests

All tests are integration tests. I use [Test Containers](https://www.testcontainers.org/) to start a PostgreSQL container and run the tests against it. The tests are run with the `make` command:

```bash
make test
```

In a real-world scenario, I would separate the tests into unit tests and integration tests. The unit tests would test the business logic and the integration tests would test the database access.

### Load tests

I have included the [load-tests](load-tests/) folder. 

I used [Vegeta](https://github.com/tsenart/vegeta) to load test the API.

Using 1000 req/sec for 5 seconds, the API appears to serve successfully 99.98% of requests. The image below shows the results of the load test.

![vegeta](load-test/vegeta-plot.png)

To avoid DB caching queries, the uses the GET /api/v1/wonderfuls endpoint with varying limits.

## Decisions

- Every user is identified by a unique `id` field. This is a good practice to avoid exposing internal IDs and to avoid exposing the number of users in the system.
  - I used the [ksuid library](https://github.com/segmentio/ksuid) to generate the `id` field. This is a good option for primary keys because it is a good balance between being unique, sortable and being human-readable.
  > KSUID is for K-Sortable Unique IDentifier. It is a kind of globally unique identifier similar to a RFC 4122 UUID, built from the ground-up to be "naturally" sorted by generation timestamp without any special type-aware logic. 
- Every user `name` is stored in the DB and presented in the API as a single string: `<title>  <first> <last>`. For more formal salutations, we can use the full string. For more informal salutations, we can strip the title from the string.
- To provide a better UX, the API will provide the `picture` field with nested `large`, `medium`, and `thumbnail` fields. This way, the client can choose the best image for the current context. For example, the `large` image can be used in the user profile page, the `medium` image can be used in the user list, and the `thumbnail` image can be used in the user search results.
  ```json
    "picture": {
        "large": "https://randomuser.me/api/portraits/men/75.jpg",
        "medium": "https://randomuser.me/api/portraits/med/men/75.jpg",
        "thumbnail": "https://randomuser.me/api/portraits/thumb/men/75.jpg"
    },
    ```
- The customer phone number will include both the landline and a cell phone number so that the customer can be contacted in any situation. In the case of a cell phone we can even provide an SMS service. The structure of the phone number will be:
  ```json
  "phone": {
    "main": "1234567890",
    "cell": "1234567890
  }
  ```


## Notes

I tried to chunk the data when downloading it from the `https://randomuser.me/api/` endpoint but it seems that the endpoint does not like concurrent requests. I was receiving `429 Too Many Requests` errors. I decided to download the data in a single request.
