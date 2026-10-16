package ora

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"
)

var (
	_ driver.QueryerContext     = (*ConnStd)(nil)
	_ driver.ConnPrepareContext = (*Conn)(nil)
	_ driver.ConnBeginTx        = (*Conn)(nil)
	_ driver.Validator          = (*Conn)(nil)
	_ driver.StmtQueryContext   = (*Statement)(nil)
)

func positionalValues(args []driver.NamedValue) ([]driver.Value, error) {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, errors.New("ora: named parameters are unsupported; use positional binds")
		}
		values[i] = arg.Value
	}
	return values, nil
}

func (stmt *Statement) ExecContext(ctx context.Context, args []driver.NamedValue) (result driver.Result, err error) {
	values, err := positionalValues(args)
	if err != nil {
		return nil, err
	}
	err = stmt.conn.runContext(ctx, func() error { result, err = stmt.Exec(values); return err })
	return result, err
}

func (conn *Conn) ExecContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue,
) (result driver.Result, err error) {
	values, err := positionalValues(args)
	if err != nil {
		return nil, err
	}
	err = conn.runContext(ctx, func() error { result, err = conn.Exec(query, values); return err })
	return result, err
}

func (conn *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var stmt *Statement
	err := conn.runContext(ctx, func() (err error) { stmt, err = conn.newStatement(query); return err })
	if err != nil {
		if stmt != nil {
			stmt.closeQuietly()
		}
		return nil, err
	}
	return stmt, nil
}

func (conn *ConnStd) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	prepared, err := conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	stmt := prepared.(*Statement)
	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		stmt.closeQuietly()
		return nil, err
	}
	rows.(*Rows).ownsStmt = true
	return rows, nil
}

func (stmt *Statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	values, err := positionalValues(args)
	if err != nil {
		return nil, err
	}
	var result driver.Rows
	err = stmt.conn.runContext(ctx, func() error { result, err = stmt.Query(values); return err })
	if err != nil {
		return nil, err
	}
	// Rows is scoped to this query; Next uses this context for each blocking fetch.
	result.(*Rows).ctx = ctx
	return result, nil
}

func (conn *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if opts.Isolation != 0 {
		return nil, errors.New("ora: non-default isolation is unsupported")
	}
	mode := OCI_TRANS_NEW
	if opts.ReadOnly {
		mode |= OCI_TRANS_READONLY
	}
	var tx driver.Tx
	err := conn.runContext(ctx, func() (err error) { tx, err = conn.begin(mode); return err })
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (conn *Conn) Ping(ctx context.Context) error {
	return conn.runContext(ctx, func() error {
		if err := conn.cerr(oci_OCIPing.Call(conn.serv.ptr, conn.err.ptr, OCI_DEFAULT)); err != nil {
			conn.bad = true
			return driver.ErrBadConn
		}
		return nil
	})
}

func (conn *Conn) ResetSession(ctx context.Context) error {
	if !conn.IsValid() {
		return driver.ErrBadConn
	}
	return ctx.Err()
}

// IsValid prevents database/sql from reusing a connection interrupted by OCIBreak.
func (conn *Conn) IsValid() bool { return conn.opened && !conn.bad }

func (conn *Conn) runContext(ctx context.Context, work func() error) error {
	if conn.bad {
		return driver.ErrBadConn
	}
	return interruptible(ctx, work, func() {
		// A separate error handle prevents concurrent access to the active call's handle.
		// OCI reports failures through its status, not the OS last-error value.
		status, _, _ := oci_OCIBreak.Call(conn.serv.ptr, conn.breakErr.ptr)
		if code := int16(status); code != OCI_SUCCESS && code != OCI_SUCCESS_WITH_INFO {
			trace.Printf("OCIBreak failed: OCI status %d", code)
		}
	}, func() { conn.bad = true })
}

// interruptible joins the breaker before returning, even when completion races
// cancellation. Invalidation occurs on the caller goroutine after that join.
func interruptible(ctx context.Context, work func() error, interrupt, invalidate func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ctx.Done() == nil {
		return work()
	}
	done, joined := make(chan struct{}), make(chan struct{})
	interrupted := false
	go func() {
		defer close(joined)
		select {
		case <-done:
		case <-ctx.Done():
			select {
			case <-done:
				return
			default:
			}
			interrupted = true
			// Repeat until work exits: a cancellation can arrive between native
			// calls or just before one, when a single break has nothing to stop.
			for {
				interrupt()
				timer := time.NewTimer(10 * time.Millisecond)
				select {
				case <-done:
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		}
	}()
	var err error
	func() {
		defer func() { close(done); <-joined }()
		err = work()
	}()
	if interrupted || ctx.Err() != nil {
		invalidate()
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
