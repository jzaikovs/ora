//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	rows, err := db.Query("select * from dual")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()

	count := 0

	for rows.Next() {
		var dummy string
		if err := rows.Scan(&dummy); err != nil {
			t.Error(err)
			return
		}

		if dummy != "X" {
			t.Error("dual not returning X")
			return
		}

		count++
	}

	if count == 0 {
		t.Error("Conn.Query() return no rows on dual query")
	}
}

func TestDatabase(t *testing.T) {
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()

	rows, err := db.Query("SELECT t.rowid, 'hello' as greet, dummy, 1337 as leet, sysdate as today from dual t connect by level <= :1", 1)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()

	var (
		rowid    string
		greet    string
		dummy    string
		leet     int
		dateTime time.Time
	)

	if !rows.Next() {
		t.Error("No rows fetched")
		return
	}

	if err := rows.Scan(&rowid, &greet, &dummy, &leet, &dateTime); err != nil {
		t.Error(err)
		return
	}

	t.Log([]interface{}{
		rowid,
		greet,
		dummy,
		leet,
		dateTime,
	})
}

func TestErrors(t *testing.T) {
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()

	_, err = db.Query("select x from dual_non_existing_table")
	var oracleErr Error
	if !errors.As(err, &oracleErr) || oracleErr.Code != 942 {
		t.Errorf("got %v, want ORA-00942", err)
		return
	}
}

func BenchmarkFetchSpeed(b *testing.B) {
	rows, err := db.Query("select -0.124235 from dual connect by level <= :1", b.N)
	if err != nil {
		b.Error(err)
		return
	}

	var val string
	count := 0

	b.StartTimer()
	for rows.Next() {
		if err := rows.Scan(&val); err != nil {
			b.Error(err)
			_ = rows.Close()
			return
		} else {
			count++
		}
	}
	b.StopTimer()
	log.Println(count)
}

func TestSimpleDualQueriesFromRaw(t *testing.T) {
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

	rows, err := conn.Query("SELECT 66 as x, :1 FROM dual", 99)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()

	if err := rows.Next(); err != nil {
		t.Error("Should have returned row", err)
		return
	}

	row, err := rows.Values()
	if err != nil {
		t.Error(err)
		return
	}

	if fmt.Sprint(row[0]) != "66" {
		t.Error("number not fetched", row[0])
		return
	}

	if fmt.Sprint(row[1]) != "99" {
		t.Error("bind passed into select not returned", row[1])
		return
	}
}
