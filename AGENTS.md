# AGENT INSTRUCTIONS & CODING GUIDELINES (Go / Golang)

This document strictly governs all code modifications, refactorings, and additions performed by AI coding agents within this repository. Adhere to idiomatic Go conventions (`Effective Go`, Go Code Review Comments) and project constraints.

---

## 1. Operating Principles & Scope Control

* **Minimal Blast Radius:** Touch only files directly related to the requested task. Do not reformat adjacent files or run blanket cleanups.
* **Dependency Discipline:**
  * Do not add external dependencies (`go get`) without explicit human instruction.
  * Prefer the standard library (`net/http`, `encoding/json`, `crypto/*`, `os`, `context`, `sync`, etc.).
  * If a dependency is approved, ensure `go.mod` and `go.sum` remain in sync (`go mod tidy`).
* **Idiomatic Code Style:**
  * Keep abstractions minimal. Accept interfaces, return concrete types (*"Accept interfaces, return structs"*).
  * Do not introduce deep inheritance-like struct embedding or Java/C#-like class patterns.

---

## 2. Go Security Standards

Treat all external inputs as untrusted and enforce defensive engineering:

* **Secrets & Configuration:**
  * Never hardcode tokens, private keys, database credentials, or environment values.
  * Inject configuration using explicit config structs populated from environment variables or flags.
* **Input Validation & Injection Prevention:**
  * Validate inputs explicitly at API and service boundaries.
  * Use parameterized queries (`database/sql` with `$1`, `?` placeholders) or ORM parameterization. **Never use `fmt.Sprintf` to compose SQL queries, shell commands, or URLs.**
  * Prevent SSRF and path traversal: validate and sanitize all file paths using `filepath.Clean` and verify against root directories before opening.
* **Safe Error Logging & Exposure:**
  * Use structured logging (`log/slog`). Log error details, request IDs, and caller context internally.
  * Return generic error messages to clients; never expose internal SQL errors, stack traces, or internal paths in HTTP/gRPC responses.
* **Cryptographic Best Practices:**
  * Use `crypto/rand` for cryptographic operations and tokens; never use `math/rand` for security-sensitive logic.
  * Use standard crypto packages (`crypto/sha256`, `golang.org/x/crypto/bcrypt` or `argon2id`).

---

## 3. Go Quality & Clean Code Standards

* **Explicit Error Handling:**
  * Check every error immediately:
    ```go
    val, err := service.DoSomething(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to do something: %w", err)
    }
    ```
  * Always wrap errors with context using `fmt.Errorf("...: %w", err)` to preserve the causal chain.
  * Never use blank identifiers to ignore errors (e.g., `_ = file.Close()` is only acceptable when safe or logged explicitly via deferred handlers).
  * **Zero Panics in Production Code:** Do not use `panic()` for control flow or expected failure conditions. Reserve `panic()` only for unrecoverable initialization bugs.
* **Context Propagation & Lifecycle:**
  * Pass `ctx context.Context` as the very first parameter of functions performing I/O, database access, or network calls.
  * Respect context cancellation (`ctx.Done()`) in long-running loops and goroutines.
  * Do not store `context.Context` inside structs; pass it explicitly down the call stack.
* **Goroutines & Concurrency Safety:**
  * Never spawn a goroutine without a clear termination guarantee or cancellation strategy (prevent goroutine leaks).
  * Guard shared mutable memory with `sync.Mutex` / `sync.RWMutex` or communicate via channels.
  * Use `sync.WaitGroup` or `golang.org/x/sync/errgroup` to manage concurrent task lifecycles.
* **Resource Cleanup:**
  * Pair allocations with `defer` immediately after successful error checks (e.g., `defer resp.Body.Close()`, `defer file.Close()`, `defer mutex.Unlock()`).

---

## 4. Testing & Verification Requirements

Every change must include automated, idiomatic Go tests:

* **Table-Driven Tests:**
  * Standardize unit tests using table-driven test patterns:
    ```go
    func TestCalculateTax(t *testing.T) {
        t.Parallel()

        tests := []struct {
            name    string
            amount  int64
            want    int64
            wantErr bool
        }{
            {name: "standard calculation", amount: 100, want: 19, wantErr: false},
            {name: "zero amount", amount: 0, want: 0, wantErr: false},
            {name: "negative amount error", amount: -10, want: 0, wantErr: true},
        }

        for _, tc := range tests {
            tc := tc
            t.Run(tc.name, func(t *testing.T) {
                t.Parallel()
                got, err := CalculateTax(tc.amount)
                if (err != nil) != tc.wantErr {
                    t.Fatalf("CalculateTax() error = %v, wantErr %v", err, tc.wantErr)
                }
                if got != tc.want {
                    t.Errorf("CalculateTax() = %v, want %v", got, tc.want)
                }
            })
        }
    }
    ```
* **No Shallow Assertions:** Verify specific values, error conditions (`errors.Is`, `errors.As`), and side effects.
* **Concurrency Testing:** Run tests with race detection enabled to catch data races early.

---

## 5. Agent Verification Checklist (Run Before Submitting)

Execute the following commands in order. Fix all reported issues before proposing changes:

1. **Format & Imports:**
   ```bash
   go fmt ./...
   go run golang.org/x/tools/cmd/goimports -w .
   ```

2. **Linting & Vetting:**
   ```bash
   go vet ./...
   golangci-lint run ./...
   ```

3. **Vulnerability Audit:**
   ```bash
   govulncheck ./...
   ```

4. **Testing with Race Detector:**
   ```bash
   go test -race -v -cover ./...
   ```

5. **Self-Review Checklist:**
   * [ ] Does every new goroutine have a controlled termination or context exit?
   * [ ] Are all errors wrapped with `%w` and appropriate context?
   * [ ] Did I avoid introducing `panic()` calls?
   * [ ] Are all resources properly closed via `defer`?
   * [ ] Did `go mod tidy` leave zero unnecessary diffs in `go.mod`/`go.sum`?