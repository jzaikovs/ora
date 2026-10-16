# Repository Guidelines

## Project Structure & Module Organization

`github.com/jzaikovs/ora` implements Go's `database/sql` Oracle driver, registered as `ora`. Source and colocated `*_test.go` files live in the repository root. Connection, statement, transaction, and row handling are in `conn.go`, `stmt.go`, `tx.go`, and `rows.go`; conversions and LOBs are in `types.go` and `lob.go`. `oci.go` declares native procedures; `oci_linux.go` and `oci_windows.go` load OCI dynamically. Agent tooling lives under `.agents/`, `.claude/`, and `agent/`.

## Build, Test, and Development Commands

Run from the repository root:

- `go build .` — compile the driver.
- `go vet .` — check for suspicious Go constructs.
- `golangci-lint run` — apply `.golangci.yml` checks.
- `go test .` — run default tests without database configuration.
- `go test -run TestParseConnectString .` — run a focused test.
- `go test -tags oci_integration .` — run native-client tests; set `ORA_TEST_NATIVE_HANDLES=1` to enable them.
- `go test -tags oracle_integration -count=1 .` — run read-only database tests without schema setup or DML.
- `go test -tags oracle_write_integration -count=1 .` — run write tests on a disposable schema.

Install Oracle Instant Client matching the Go toolchain's architecture. Windows needs `oci.dll` on `PATH`; Linux loads `libclntsh.so` at package initialization, including for unit tests.

## Coding Style & Naming Conventions

Format changed Go files with `gofmt` (tab indentation). Use package `ora`, exported `MixedCaps` names, and unexported `mixedCaps` names. Follow existing receiver names such as `conn`, `stmt`, and `rows`. Keep platform-specific loading in the OS-suffixed files.

## Testing & Configuration

Use Go's `testing` package with `TestXxx` functions in `*_test.go`. Add regression coverage for changed behavior; no numerical coverage threshold is configured.

The Oracle tags enable the database `TestMain`. Copy `.env.example` to `.env` and supply `ORA_TEST_USER`, `ORA_TEST_PASSWORD`, and `ORA_TEST_HOST`, or set environment variables (which take precedence). Use read-only credentials for `oracle_integration`; `oracle_write_integration` creates/drops `go_test` and requires a disposable schema. See `README.md` for setup and grants. Keep credentials uncommitted.

## Architecture Guardrails

Preserve `Statement.binds` and `Descriptor.valPtr` references while OCI retains buffer addresses. Free allocated handles. Autocommit depends on `stmt.conn.tx`. Context operations use `Conn.runContext` and `interruptible`; cancellation invokes `OCIBreak` and invalidates the connection. Preserve `Rows.ownsStmt` ownership so prepared statements remain reusable.

## Commit & Pull Request Guidelines

History uses concise imperative subjects, sometimes with `feat:` prefixes. Describe the behavior changed. In pull requests, explain the problem, link relevant issues, and report checks run plus any unavailable Oracle-dependent validation.
