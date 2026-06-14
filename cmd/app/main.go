package main

import (
	"context"
	"log"
	"net/http"
	"time"

	postgresrepo "github.com/example/registration/internal/adapters/postgres"
	"github.com/example/registration/internal/adapters/security"
	"github.com/example/registration/internal/app"
	httptransport "github.com/example/registration/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Hardcoded only for local development.
// NOTE: Do not use hardcoded values in production:
//
//	we need to make this configurable (via yaml files for each env) and add secrets file to store db password for each env.
const (
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
	defaultJWTSecret   = "replace-this-with-a-long-random-secret"
	defaultAddr        = ":8080"
)

// main starts the HTTP server and wires up dependencies.
func main() {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, defaultDatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	userRepository := postgresrepo.NewUserRepository(db)
	passwordHasher := security.NewBcryptHasher()
	tokenIssuer := security.NewJWTIssuer(
		defaultJWTSecret,
		"registration",
		24*time.Hour,
	)

	appService := app.NewService(
		userRepository,
		passwordHasher,
		tokenIssuer,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	httptransport.Register(mux, appService)

	server := &http.Server{
		Addr:              defaultAddr,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server listening on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// healthHandler returns a simple liveness check.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// logRequests logs each incoming request before passing it along.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
