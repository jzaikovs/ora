//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestExecCancel(t *testing.T) {
	conn, err := sql.Open(DriverName, testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() { _ = conn.Close() }()

	ctx, clean := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer clean()

	beforeStart := time.Now()

	_, err = conn.ExecContext(ctx, "begin dbms_lock.sleep(:1); end;", 1.5)
	if err == nil {
		t.Error("Error not raised", time.Since(beforeStart))
		return
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error(err)
		return
	}
}
