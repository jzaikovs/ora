package ora

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
)

func TestWrongHost(t *testing.T) {
	_, err := Open("test/test@127.0.0.1:1/test")
	if err == nil {
		t.Error("No error on bad connect attempt")
		return
	}

	if err.Error() != "ORA-12541: TNS:no listener" {
		t.Errorf("Wrong error %v", []byte(err.Error()))
		return
	}
}

func TestWrongUsername(t *testing.T) {
	_, err := Open(fmt.Sprintf("%s/test@%s", dbUser, dbHost))
	if err == nil {
		t.Error("No error on bad connect attempt")
		return
	}

	if err.Error() != "ORA-01017: invalid username/password; logon denied" {
		t.Errorf("Wrong error %v", err)
		return
	}
}

func TestSimpleDualQueriesFromStd(t *testing.T) {
	refCount := atomic.LoadInt64(&handleRefCount)

	conn, err := sql.Open(DriverName, testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	rows, err := conn.Query("SELECT 66 as x, :1 FROM dual", 99)
	if err != nil {
		t.Error(err)
		return
	}

	if !rows.Next() {
		t.Error("Should have returned row")
		return
	}

	var col, col2 int
	if err := rows.Scan(&col, &col2); err != nil {
		t.Error(err)
		return
	}

	if col != 66 {
		t.Error("number not fetched", col)
		return
	}

	if col2 != 99 {
		t.Error("bind passed into select not returned", col)
		return
	}

	if err := rows.Close(); err != nil {
		t.Error(err)
		return
	}

	if err = conn.Close(); err != nil {
		t.Error(err)
		return
	}

	if gc := atomic.LoadInt64(&handleRefCount); gc != refCount {
		t.Errorf("Handles not freed, refCount %v after test %v", refCount, gc)
	}
}

func TestHandlesFreed(t *testing.T) {
	allocCount := handleRefCount

	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	allocCountAfterOpen := handleRefCount
	result, err := conn.Query("SELECT 99 as x, :1 FROM dual", 99)
	if err != nil {
		t.Error(err)
		return
	}

	if err := result.Next(); err != nil {
		t.Error("Should have returned row", err)
		return
	}

	if err := result.Close(); err != nil {
		t.Error(err)
		return
	}

	if handleRefCount != allocCountAfterOpen {
		t.Errorf("Handles not freed after result close allocCount %v after test %v", allocCountAfterOpen, handleRefCount)
		return
	}

	if err = conn.Close(); err != nil {
		t.Error(err)
		return
	}

	if handleRefCount != allocCount {
		t.Errorf("Handles not freed, allocCount %v after test %v", allocCount, handleRefCount)
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
	rows.Close()

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
