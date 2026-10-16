//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"context"
	"database/sql/driver"
	"testing"
)

func TestPingerContext(t *testing.T) {
	if err := db.Ping(); err != nil {
		t.Error(err)
	}
}

func TestExecContextFreesStatement(t *testing.T) {
	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()

	allocCountAfterOpen := handleRefCount

	if _, err := conn.ExecContext(context.Background(), "begin null; end;", []driver.NamedValue{}); err != nil {
		t.Error(err)
		return
	}

	if _, err := conn.ExecContext(context.Background(), "begin null; end; bad", []driver.NamedValue{}); err == nil {
		t.Error("No error on bad statement")
	}

	if handleRefCount != allocCountAfterOpen || len(conn.statements) != 0 {
		t.Errorf("Statement not freed after ExecContext allocCount %v after exec %v, open statements %v", allocCountAfterOpen, handleRefCount, len(conn.statements))
	}
}
