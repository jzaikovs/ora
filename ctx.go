// +build go1.8

package ora

import (
	"context"
	"database/sql/driver"
)

func (stmt *Statement) ExecContext(ctx context.Context, args []driver.NamedValue) (result driver.Result, err error) {
	argCopy := make([]driver.Value, len(args))
	for i := range args {
		argCopy[i] = args[i].Value
	}

	stmt.conn.handleContext(ctx, func() {
		result, err = stmt.Exec(argCopy)
	})

	return
}

func (conn *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	stmt, err := conn.newStatement(query)
	if err != nil {
		return nil, err
	}

	result, err = stmt.ExecContext(ctx, args)
	return
}

func (conn *Conn) Ping(ctx context.Context) (err error) {
	conn.handleContext(ctx, func() {
		// any error from this is bad connection?
		if err = conn.cerr(oci_OCIPing.Call(conn.serv.ptr, conn.err.ptr, OCI_DEFAULT)); err != nil {
			err = driver.ErrBadConn
		}
	})
	return
}

func (conn *Conn) ResetSession(ctx context.Context) (err error) {
	conn.handleContext(ctx, func() {
		// any error from this is bad connection?
		err = conn.cerr(oci_OCIReset.Call(conn.serv.ptr, conn.err.ptr, OCI_DEFAULT))
	})
	return
}

// handleContext handles context lifetime interrupt calling OCIBreak or gracefully handles successful work done
func (conn *Conn) handleContext(ctx context.Context, work func()) {
	done := make(chan bool)
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			err := conn.cerr(oci_OCIBreak.Call(conn.serv.ptr, conn.err.ptr))
			trace.Println(err)
		}
	}()

	work()
}
