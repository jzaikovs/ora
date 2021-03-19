// +build go1.8

package ora

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestExecCancel(t *testing.T) {
	conn, err := sql.Open(DriverName, testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	ctx, clean := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer clean()

	beforeStart := time.Now()

	_, err = conn.ExecContext(ctx, "begin dbms_lock.sleep(:1); end;", 1.5)
	if err == nil {
		t.Error("Error not raised", time.Since(beforeStart))
		return
	}

	if err.Error() != "ORA-01013: user requested cancel of current operation" {
		t.Error(err)
		return
	}
}

func TestPingerContext(t *testing.T) {
	if err := db.Ping(); err != nil {
		t.Error(err)
	}
}
