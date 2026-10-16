//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrongUsername(t *testing.T) {
	_, err := Open(fmt.Sprintf("%s/test@%s", dbUser, dbHost))
	if err == nil {
		t.Error("No error on bad connect attempt")
		return
	}

	var oracleErr Error
	if !errors.As(err, &oracleErr) || oracleErr.Code != 1017 {
		t.Errorf("got %v, want ORA-01017", err)
		return
	}
}

func TestConnectionClosed(t *testing.T) {
	other, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	rows, err := other.Query("SELECT 1 FROM dual")
	if err != nil {
		t.Error(err)
		return
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}

	if err = other.Close(); err != nil {
		t.Error(err)
		return
	}

	if !checkSessionCount(t, 1) {
		t.Error("Session not closed")
	}
}

func checkSessionCount(t *testing.T, expected int) bool {
	row := db.QueryRow("SELECT count(1) FROM v$session WHERE username = :1", dbUser)
	if row == nil {
		t.Error("Can't query sessions")
		return false
	}

	cnt := 0
	if err := row.Scan(&cnt); err != nil {
		t.Error(err)
		return false
	}

	if cnt != expected {
		t.Error("Session count doesn't match expected")
		return false
	}

	return true
}
