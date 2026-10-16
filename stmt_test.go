//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"errors"
	"strconv"
	"testing"
)

// https://vrogier.github.io/ocilib/doc/html/group___ocilib_c_api_abort.html

func TestBadExecute(t *testing.T) {
	rows, err := db.Query("select * from dual", 1, 2, 3)
	if rows != nil {
		if closeErr := rows.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	}
	var oracleErr Error
	if !errors.As(err, &oracleErr) || oracleErr.Code != 1036 {
		t.Errorf("got %v, want ORA-01036", err)
	}
}

func TestRowsErrNilAfterIteration(t *testing.T) {
	rows, err := db.Query("select 1 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()

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
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()

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
	defer func() {
		if err := s2.Close(); err != nil {
			t.Error(err)
		}
	}()

	if err := s1.Close(); err != nil {
		t.Fatal(err)
	}

	s3, err := conn.newStatement("select 3 from dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := s3.Close(); err != nil {
			t.Error(err)
		}
	}()

	if s3.id == s2.id || conn.statements[s2.id] != s2 {
		t.Error("new statement overwrote open statement id", s2.id)
	}
}
