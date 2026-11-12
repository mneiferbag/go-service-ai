package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHelloHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	HelloHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	trimmedBody := strings.TrimSpace(string(body))
	expectedBody := `{"message":"hello"}`
	if trimmedBody != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, trimmedBody)
	}

	var response HelloResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response.Message != "hello" {
		t.Errorf("expected message %q, got %q", "hello", response.Message)
	}
}

func TestVocabHandler(t *testing.T) {
	t.Parallel()

	validPairs := make(map[string]string, len(defaultVocabulary))
	for _, entry := range defaultVocabulary {
		validPairs[entry.De] = entry.Es
	}

	tests := []struct {
		name string
	}{
		{name: "returns valid German-Spanish vocabulary pair"},
		{name: "consistent valid schema on repeated request"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/vocab", nil)
			rec := httptest.NewRecorder()

			VocabHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
			}

			contentType := res.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", contentType)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			var entry VocabularyEntry
			if err := json.Unmarshal(body, &entry); err != nil {
				t.Fatalf("failed to unmarshal JSON response %q: %v", string(body), err)
			}

			if entry.De == "" || entry.Es == "" {
				t.Fatalf("expected non-empty vocabulary fields, got %+v", entry)
			}

			expectedEs, ok := validPairs[entry.De]
			if !ok {
				t.Errorf("unexpected German word %q in response", entry.De)
			} else if entry.Es != expectedEs {
				t.Errorf("for German %q expected Spanish %q, got %q", entry.De, expectedEs, entry.Es)
			}
		})
	}
}

func TestServerRouting(t *testing.T) {
	t.Parallel()

	server := NewServer()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET /hello returns 200 OK",
			method:         http.MethodGet,
			path:           "/hello",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /hello returns 405 Method Not Allowed",
			method:         http.MethodPost,
			path:           "/hello",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET /vocab returns 200 OK",
			method:         http.MethodGet,
			path:           "/vocab",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /vocab returns 405 Method Not Allowed",
			method:         http.MethodPost,
			path:           "/vocab",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET /nonexistent returns 404 Not Found",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}

func TestServerIntegration(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(NewServer())
	defer ts.Close()

	t.Run("GET /hello", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/hello")
		if err != nil {
			t.Fatalf("GET /hello failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		trimmedBody := strings.TrimSpace(string(body))
		if trimmedBody != `{"message":"hello"}` {
			t.Errorf("expected body %q, got %q", `{"message":"hello"}`, trimmedBody)
		}
	})

	t.Run("GET /vocab", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/vocab")
		if err != nil {
			t.Fatalf("GET /vocab failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var entry VocabularyEntry
		if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
			t.Fatalf("failed to decode VocabularyEntry: %v", err)
		}

		if entry.De == "" || entry.Es == "" {
			t.Errorf("expected non-empty VocabularyEntry fields, got %+v", entry)
		}
	})
}

func TestConfigFromEnv(t *testing.T) {
	t.Run("default port when PORT not set", func(t *testing.T) {
		t.Setenv("PORT", "")
		cfg := ConfigFromEnv()
		if cfg.Port != "8080" {
			t.Errorf("expected default port 8080, got %q", cfg.Port)
		}
	})

	t.Run("custom port when PORT set", func(t *testing.T) {
		t.Setenv("PORT", "3000")
		cfg := ConfigFromEnv()
		if cfg.Port != "3000" {
			t.Errorf("expected port 3000, got %q", cfg.Port)
		}
	})
}

func TestSetupServer(t *testing.T) {
	handler := NewServer()
	srv := SetupServer(":9090", handler)

	if srv.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", srv.Addr)
	}
	if srv.Handler != handler {
		t.Errorf("handler not properly set")
	}
	if srv.ReadTimeout != 5*time.Second {
		t.Errorf("expected ReadTimeout 5s, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("expected IdleTimeout 120s, got %v", srv.IdleTimeout)
	}
}

func TestRunServerGracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	addr := listener.Addr().String()
	srv := SetupServer(addr, NewServer())
	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		errChan <- RunServer(ctx, srv, listener)
	}()

	resp, err := http.Get("http://" + addr + "/hello")
	if err != nil {
		cancel()
		t.Fatalf("failed to make request to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Trigger graceful shutdown
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("expected clean shutdown without error, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for server to shut down")
	}
}
