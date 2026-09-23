# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`github.com/jzaikovs/ora` is an Oracle database driver for Go's `database/sql`, built on Oracle's OCI C library. It registers itself as driver name `"ora"` in `init()` (`driver.go`). It does **not** use cgo bindings to OCI headers: OCI functions are loaded dynamically at runtime and called through `uintptr` arguments.

## Commands

```sh
go build ./...
go vet ./...
go test ./...                       # needs a live Oracle DB (see below)
go test -run TestPrepareInsert ./...  # single test
```

Testing requirements:
- The Oracle Instant Client must be installed and loadable (`libclntsh.so` on Linux — check with `ldconfig -p | grep libclntsh`; `oci.dll` on PATH on Windows). A 64-bit Go toolchain needs the 64-bit client.
- Every test, including the pure conversion tests in `types_test.go`, runs under `TestMain` in `ora_test.go`, which connects to `ORA_GO_TEST/ora_go_test_password@//localhost:1521/XE` and creates/drops a `go_test` table. The SQL to create that test user is in the comment at the top of `ora_test.go`.
- `TestMain` panics if `handleRefCount > 0` after the run — any new code that allocates an OCI handle must free it.

## Architecture

**Dynamic library loading.** `oci_linux.go` uses `github.com/sergewu/dl` (dlopen, `RTLD_LAZY`) and `oci_windows.go` uses `syscall.NewLazyDLL`; both expose `ociLibrary.NewProc(name)`. `oci.go` holds all OCI constants and the `oci_OCI*` proc variables. The Linux loader **panics at package init** if `libclntsh.so` cannot be opened. To use a new OCI function, add a `NewProc` entry in `oci.go`.

**Calling convention.** Every OCI call is `oci_X.Call(uintptr...)`, and the return goes through `conn.cerr(...)` (error handle) or `conn.envErr(...)` (env handle) → `onOCIReturn`, which turns `OCI_ERROR` into an `ora.Error{Code, Message}` via `OCIErrorGet`. Pointer helpers such as `intRef`, `bufAddr` and `ref` are in `unsafe.go`.

**GC safety for binds/defines.** OCI keeps raw addresses of bind and define buffers, so Go values must stay referenced for as long as OCI may use them: `Statement.binds` keeps every bound value alive (`stmt.go` `bind`), and `Descriptor.valPtr` keeps define buffers alive. Keep this pattern whenever you add a bind or column type.

**Object model.**
- `Conn` (`conn.go`) owns the env, service-context and error handles and tracks open statements. `ConnStd` wraps it to satisfy `driver.Conn`/`driver.Queryer`. `Conn.Query` also offers a non-`database/sql` API that returns `*QueryResult` (`fetcher.go`).
- Connect strings are parsed by `patternEZConnect` in `driver.go` (`user/pass@//host:port/service` or `user/pass@tnsname`).
- `Statement` (`stmt.go`) is prepared with `OCIStmtPrepare2`, which gives statement caching. Outside a transaction (`stmt.tx == nil`), execution uses `OCI_COMMIT_ON_SUCCESS` (autocommit). `Conn.Begin` sets `conn.tx`, and statements capture it when they are created. `NumInput` returns -1, and binds are positional (`:1, :2…`).
- `Rows` (`rows.go`) creates a `Descriptor` (`descriptor.go`) per column and defines output buffers by Oracle type: NUMBER is fetched as VARNUM and returned as a **string** (`convertNumberToString` in `types.go`); DATE is decoded to `time.Time` in `time.Local`; VARCHAR/CHAR buffers are sized `length*2+2` to handle client/server encoding differences; LONG uses the fixed `MaxLongSize` buffer; CLOB/BLOB return a `*Lob` (`lob.go`) that implements `io.Reader`/`Writer` and `sql.Scanner`. Unsupported types return an error from `newRows`.
- Context support (`ctx.go`): `handleContext` runs the work synchronously and calls `OCIBreak` from a goroutine if the context is cancelled.
- Package-level tunables: `PrefetchRows` and `MaxLongSize` (`stmt.go`).

Debug logging uses the package-level `trace` logger in `driver.go`, which writes to `ioutil.Discard`. To debug, switch it to the commented-out `os.Stdout` line.
