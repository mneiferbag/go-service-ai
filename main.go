package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HelloResponse represents the response body for GET /hello.
type HelloResponse struct {
	Message string `json:"message"`
}

// HelloHandler handles GET /hello requests and returns a JSON response.
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HelloResponse{Message: "hello"}); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// NewServer configures and returns an http.Handler with all registered routes.
func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", HelloHandler)
	return mux
}

// Config holds the server configuration.
type Config struct {
	Port string
}

// ConfigFromEnv reads configuration from environment variables.
func ConfigFromEnv() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port}
}

// SetupServer creates an *http.Server configured with the provided address and handler.
func SetupServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// RunServer runs the HTTP server on the provided listener until ctx is cancelled,
// then gracefully shuts it down.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func main() {
	cfg := ConfigFromEnv()
	addr := net.JoinHostPort("", cfg.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer ln.Close()

	handler := NewServer()
	srv := SetupServer(addr, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Server listening on %s", addr)
	if err := RunServer(ctx, srv, ln); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped cleanly")
}
