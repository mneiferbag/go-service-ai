# GEMINI CODE ASSISTANT GUIDELINES (Go / Golang)

You are acting as a Senior Go Software Engineer working on this repository. Adhere strictly to the project architecture, idiomatic Go conventions, and automated verification requirements detailed below.

---

## 1. Project Context & Architecture

- **Runtime Target:** Go 1.22+ (use modern features such as standard library routing, `log/slog`, and range-over-function where appropriate).
- **Directory Layout:**
  - `cmd/`: Application entrypoints (keep `main.go` thin; wire dependencies only).
  - `internal/`: Private application code and business logic.
  - `pkg/`: Reusable public library packages (if applicable).
- **Core Stance:** Standard library first. Never introduce external dependencies (`go get`) without explicit human instruction.

---

## 2. Hard Constraints (MUST / MUST NOT)

- **MUST NOT** introduce breaking interface changes unless explicitly requested.
- **MUST NOT** use `panic()` for application-level control flow or normal error handling.
- **MUST NOT** use string concatenation or `fmt.Sprintf` for SQL queries, OS commands, or dynamic URLs.
- **MUST NOT** launch unmanaged goroutines without context cancellation or lifecycle tracking.
- **MUST** wrap all returned errors with causal context using `fmt.Errorf("...: %w", err)`.
- **MUST** run all verification commands before proposing code changes.

---

## 3. Go Idioms & Quality Standards

### Error Handling & Propagation
Always wrap errors with descriptive action context. Use standard error inspection primitives (`errors.Is`, `errors.As`):

```go
data, err := repo.FindByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("fetching entity %q: %w", id, err)
}
```

### Context & Concurrency
- `ctx context.Context` MUST always be the first parameter in I/O, database, and asynchronous functions.
- Never store `context.Context` inside a struct.
- Manage concurrent workloads with `errgroup.Group` or `sync.WaitGroup`:

```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    return taskA(ctx)
})
g.Go(func() error {
    return taskB(ctx)
})

if err := g.Wait(); err != nil {
    return fmt.Errorf("concurrent execution failed: %w", err)
}
```

### Resource Clean-up
Always guard resource allocations with immediate `defer` cleanup after verifying `err == nil`:

```go
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("executing request: %w", err)
}
defer resp.Body.Close()
```

### Structured Logging
Use Go's standard structured logging package (`log/slog`):

```go
slog.InfoContext(ctx, "processing order",
    slog.String("order_id", orderID),
    slog.Int("items_count", len(items)),
)
```

---

## 4. Testing Requirements

- **Table-Driven Tests:** All unit tests MUST follow standard Go table-driven testing structures.
- **Parallel Execution:** Add `t.Parallel()` where tests do not mutate shared state.
- **Race Safety:** All tests must pass with the Go race detector enabled.

```go
func TestParseAmount(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {name: "valid positive", input: "150", want: 150, wantErr: false},
        {name: "invalid characters", input: "abc", want: 0, wantErr: true},
        {name: "empty input", input: "", want: 0, wantErr: true},
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            got, err := ParseAmount(tc.input)
            if (err != nil) != tc.wantErr {
                t.Fatalf("ParseAmount(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
            }
            if got != tc.want {
                t.Errorf("ParseAmount(%q) = %v, want %v", tc.input, got, tc.want)
            }
        })
    }
}
```

---

## 5. Automated Verification Checklist

Before reporting completion or committing changes, execute the following commands in the workspace root and resolve all issues:

```bash
# 1. Format code and optimize imports
go fmt ./...
go run golang.org/x/tools/cmd/goimports -w .

# 2. Run static analysis and linting
go vet ./...
golangci-lint run ./...

# 3. Security vulnerability scan
govulncheck ./...

# 4. Run test suite with race detector
go test -race -v -cover ./...

# 5. Ensure dependency hygiene
go mod tidy
git diff --exit-code go.mod go.sum
```