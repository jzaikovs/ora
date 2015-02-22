package ora

import (
	"database/sql"
	"github.com/jzaikovs/clitable"
	"testing"
)

func TestExec(t *testing.T) {
	db, err := sql.Open("ora", "jp/parole@//192.168.1.120:1521/XE")
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	if _, err = db.Exec("create table go_test(id number)"); err != nil {
		t.Error(err)
		//return
	}

	if _, err = db.Exec("insert into go_test values(:1)", 1337); err != nil {
		t.Error(err)
	}

	stmt, err := db.Prepare("insert into go_test values(:1)")
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 5; i++ {
		stmt.Exec(i)
	}

	r, err := db.Query("select rowid, id from go_test")
	if err != nil {
		t.Error(err)
	}

	clitable.Print(r)

	if err = r.Close(); err != nil {
		t.Error(err)
	}

	if _, err = db.Exec("drop table go_test"); err != nil {
		t.Error(err)
		return
	}
}

func TestDatabase(t *testing.T) {
	db, err := sql.Open("ora", "jp/parole@//192.168.1.120:1521/XE")
	if err != nil {
		t.Error(err)
		return
	}

	rows, err := db.Query("select t.rowid, 'hello' as greet, dummy, 1337 as leet, sysdate as today from dual t connect by level <= :1", 1)
	if err != nil {
		t.Error(err)
		return
	}

	if err = clitable.Print(rows); err != nil {
		t.Error(err)
		return
	}

	if rows, err = db.Query("select id, login, full_name, block_status, rowid from users where  rownum < 10"); err != nil {
		t.Error(err)
		return
	}

	if err = clitable.Print(rows); err != nil {
		t.Error(err)
		return
	}

	if err = db.Close(); err != nil {
		t.Error(err)
		return
	}
}

func TestErrors(t *testing.T) {
	db, err := sql.Open("ora", "jp/parole@//192.168.1.120:1521/XE")
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Query("select x from dual_non_existing_table")
	if err == nil || err.Error() != "ORA-00942: table or view does not exist\n" {
		t.Error("should raise error: ORA-00942: table or view does not exist")
		return
	}
}

func BenchmarkQuery(b *testing.B) {
	b.StopTimer()
	db, err := sql.Open("ora", "jp/parole@//192.168.1.120:1521/XE")
	if err != nil {
		b.Error(err)
		return
	}
	defer db.Close()
	rows, err := db.Query("select dummy from dual connect by level < :1", b.N)
	if err != nil {
		b.Error(err)
		return
	}
	defer rows.Close()
	b.StartTimer()

	count := 0
	var val string
	for rows.Next() {
		if err := rows.Scan(&val); err != nil {
			b.Error(err)
		}
		count++
	}

	b.Log(count)
}
