package ora

import (
	"database/sql/driver"
	"strconv"
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

func TestRowsErrNilAfterIteration(t *testing.T) {
	rows, err := db.Query("select 1 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
	}

	if err := rows.Err(); err != nil {
		t.Error("rows.Err() after full iteration:", err)
	}
}

func TestPreparedStmtQueryAfterRowsClose(t *testing.T) {
	stmt, err := db.Prepare("select :1 from dual")
	if err != nil {
		t.Error(err)
		return
	}

	for i := 1; i <= 2; i++ {
		var x string
		if err := stmt.QueryRow(i).Scan(&x); err != nil {
			t.Errorf("query %d: %v", i, err)
			return
		}
		if x != strconv.Itoa(i) {
			t.Errorf("query %d returned %q", i, x)
		}
	}

	if err := stmt.Close(); err != nil {
		t.Error("stmt.Close():", err)
	}
}

func TestStatementIDsUnique(t *testing.T) {
	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	s1, err := conn.newStatement("select 1 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	s2, err := conn.newStatement("select 2 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer s2.Close()

	s1.Close()

	s3, err := conn.newStatement("select 3 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer s3.Close()

	if s3.id == s2.id || conn.statements[s2.id] != s2 {
		t.Error("new statement overwrote open statement id", s2.id)
	}
}
