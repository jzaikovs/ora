package ora

import (
	"database/sql/driver"
	"testing"
	"time"
)

// https://vrogier.github.io/ocilib/doc/html/group___ocilib_c_api_abort.html
func TestExecDbmsSleep(t *testing.T) {
	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	then := time.Now()

	stmt, err := conn.Prepare("begin dbms_lock.sleep(2); end;")
	if err != nil {
		t.Error(err)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec([]driver.Value{}); err != nil {
		t.Error(err)
		return
	}

	if time.Since(then) < time.Second*2 {
		t.Error("Not slept enough")
		return
	}
}

func TestBadExecute(t *testing.T) {
	rows, err := db.Query("select * from dual", 1, 2, 3)
	if err == nil && err.Error() != "ORA-01036: illegal variable name/number" {
		rows.Close()
		t.Error("No error on bad statement")
		return
	}
}

func TestExecReturnsOutParams(t *testing.T) {
	db.Exec("truncate table go_test")

	out := 0
	db, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	defer db.Close()

	_, err = db.Exec("begin :1 := 123; insert into go_test (id) values(:1); end;", []driver.Value{&out})
	if err != nil {
		t.Error(err)
		return
	}

	if out != 123 {
		t.Error("out binds not returned value", out)
	}
}

func xTestExecReturnsOutParams2(t *testing.T) {
	out := 0
	_, err := db.Exec("begin :x := 123; end;", &out)
	if err != nil {
		t.Error(err)
		return
	}

	if out != 123 {
		t.Error("out binds not returned value", out)
	}
}
