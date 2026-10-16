//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"database/sql"
	"errors"
	"sync/atomic"
	"testing"
)

func TestWrongHost(t *testing.T) {
	_, err := Open("test/test@127.0.0.1:1/test")
	if err == nil {
		t.Error("No error on bad connect attempt")
		return
	}

	var oracleErr Error
	if !errors.As(err, &oracleErr) || oracleErr.Code != 12541 {
		t.Errorf("got %v, want ORA-12541", err)
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
