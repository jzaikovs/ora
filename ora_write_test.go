//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func setupOracleIntegration() (err error) {
	// fmt.Println("setup...")
	_, err = db.Exec("CREATE TABLE go_test(id number, int64 number(9), float64 number(9, 3), name varchar2(32), date_bind date, lobcol clob, data blob)")
	return
}

func teardownOracleIntegration() (err error) {
	// fmt.Println("tearDown...")
	_, err = db.Exec("DROP TABLE go_test")
	return
}

func TestPrepareInsert(t *testing.T) {
	if _, err := db.Exec("truncate table go_test"); err != nil {
		t.Fatal(err)
	}

	stmt, err := db.Prepare("insert into go_test (id, name, date_bind) values(:1, :2, :3)")
	if err != nil {
		t.Error(err)
		return
	}

	n := 5

	for i := 0; i < n; i++ {
		if _, err = stmt.Exec(i, "#"+fmt.Sprint(i), time.Now()); err != nil {
			t.Error(err)
			return
		}
	}

	if err = stmt.Close(); err != nil {
		t.Error(err)
		return
	}

	var cnt float64
	row := db.QueryRow("select count(1) from go_test")
	if row == nil {
		t.Error("No row fetched")
		return
	}

	if err = row.Scan(&cnt); err != nil {
		t.Error(err)
		return
	}

	if int(cnt) != n {
		t.Log(cnt)
		t.Error("After insert count not what expected")
	}
}

func TestPrepareInsert2(t *testing.T) {
	stmt, err := db.Prepare("insert into go_test (id, name, date_bind) values(:1, :2, :3)")
	if err != nil {
		t.Error(err)
		return
	}

	// now := time.Now()
	for i := 0; i < 10; i++ { // executing insert
		if _, err = stmt.Exec(100000+i, "#"+fmt.Sprint(i), time.Now()); err != nil {
			t.Error(err)
			break
		}
	}
	// t.Log(time.Since(now).Seconds())

	if err = stmt.Close(); err != nil {
		t.Error(err)
	}
}

func TestDelete(t *testing.T) {
	var err error

	if _, err = db.Exec("insert into go_test (id, name, date_bind) values(:1, :2, :3)", 1337, "leet", time.Now()); err != nil {
		t.Error(err)
	}

	if _, err = db.Exec("delete go_test where id = :1", 2); err != nil {
		t.Error(err)
	}

	if _, err = db.Exec("delete go_test where name = :1", "leet"); err != nil {
		t.Error(err)
	}
}

func TestTransactions(t *testing.T) {
	if _, err := db.Exec("DELETE FROM go_test"); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec("INSERT INTO go_test (id) VALUES (:1)", 123); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := tx.QueryRow("SELECT count(*) FROM go_test WHERE id = 123").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rows inside transaction = %d, want 1", count)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT count(*) FROM go_test WHERE id = 123").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rows after rollback = %d, want 0", count)
	}
}

func query(sql string, fn func(*sql.Rows) error, binds ...interface{}) (err error) {
	result, err := db.Query(sql, binds...)
	if err != nil {
		return err
	}

	defer func() { _ = result.Close() }()

	for result.Next() {
		if err = fn(result); err != nil {
			return err
		}
	}

	return result.Err()
}
