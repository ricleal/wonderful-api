package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	omiddleware "github.com/oapi-codegen/nethttp-middleware"

	apiv1 "wonderful/internal/api/v1"
	openapiv1 "wonderful/internal/api/v1/openapi"
	"wonderful/internal/repository/db"
	"wonderful/internal/service"
	"wonderful/internal/store"
)

func apiV1Router(root *chi.Mux, su service.UserService) error {
	wonderfulAPI := apiv1.New(su)

	swagger, err := openapiv1.GetSwagger()
	if err != nil {
		return fmt.Errorf("error getting swagger: %w", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	r := chi.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(omiddleware.OapiRequestValidator(swagger))
	r.Use(middleware.AllowContentType("application/json"))          //nolint:goconst //ignore
	r.Use(middleware.SetHeader("Content-Type", "application/json")) //nolint:goconst //ignore

	root.Mount("/api/v1", http.StripPrefix("/api/v1", openapiv1.HandlerFromMux(wonderfulAPI, r)))

	apiJSON, err := json.Marshal(swagger)
	if err != nil {
		return fmt.Errorf("error marshaling swagger: %w", err)
	}
	root.Get("/api/v1/api.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(apiJSON)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return nil
}

func printRoutes(_ context.Context, r chi.Router) {
	walkFunc := func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		slog.Info("registered route", "method", method, "route", route)
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		slog.Error("error walking routes", "error", err)
	}
}

func main() {
	ctx := context.Background()

	port := flag.Int("port", 8888, "Port for the HTTP server")
	flag.Parse()

	// Set up our data store
	dbServer, err := db.NewStorage(ctx)
	if err != nil {
		slog.Error("error connecting to database", "error", err)
	}
	defer dbServer.Close()

	// we need to create a http client to fetch random users.
	c := http.Client{Timeout: 10 * time.Second}
	s := store.NewPersistentStore(dbServer.Pool())
	su := service.NewUserService(s, c)

	// Set up the root router
	root := chi.NewRouter()
	root.Use(middleware.Logger)
	root.Use(middleware.Recoverer)
	root.Use(middleware.StripSlashes)

	// Set up API v1
	if err := apiV1Router(root, su); err != nil {
		slog.Error("error setting up api v1 router", "error", err)
		return
	}

	// Print out the routes if we're in debug mode
	printRoutes(ctx, root)

	// Start the server
	if err := serve(ctx, root, *port); err != nil {
		slog.Error("error serving http", "error", err)
		return
	}
}

func serve(ctx context.Context, handler http.Handler, port int) error {
	srv := &http.Server{
		Handler:     handler,
		Addr:        fmt.Sprintf(":%d", port),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		ReadTimeout: 10 * time.Second,
	}

	errChan := make(chan error)
	go func() {
		slog.Info("serving http on port", "port", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start server: %w", err)
		}
	}()
	ctx, stop := signal.NotifyContext(ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown gracefully: %w", err)
	}
	slog.Info("server shutdown gracefully")
	return nil
}
