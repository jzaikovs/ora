//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestExecDbmsSleep(t *testing.T) {
	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() { _ = conn.Close() }()

	then := time.Now()

	stmt, err := conn.Prepare("begin dbms_lock.sleep(2); end;")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() { _ = stmt.Close() }()

	if _, err := stmt.Exec([]driver.Value{}); err != nil { //nolint:staticcheck // Exercise the supported legacy driver.Stmt API.
		t.Error(err)
		return
	}

	if time.Since(then) < time.Second*2 {
		t.Error("Not slept enough")
		return
	}
}

func TestExecReturnsOutParams(t *testing.T) {
	if _, err := db.Exec("truncate table go_test"); err != nil {
		t.Fatal(err)
	}

	out := 0
	db, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	defer func() { _ = db.Close() }()

	_, err = db.Exec("begin :1 := 123; insert into go_test (id) values(:1); end;", []driver.Value{&out})
	if err != nil {
		t.Error(err)
		return
	}

	if out != 123 {
		t.Error("out binds not returned value", out)
	}
}
