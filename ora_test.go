package ora

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jzaikovs/clitable"
)

var (
	testDBConnectString = "ora_go_test/ora_go_test_password@//oracle:1521/XE"
)

/*
Setup for testing:
```
create user ora_go_test identified by ora_go_test_password;
grant connect, resource to ora_go_test;
```

*/

func TestExec(t *testing.T) {
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	if _, err = db.Exec("create table go_test(id number, name varchar2(32), date_bind date)"); err != nil {
		t.Error(err)
		//return
	}

	if _, err = db.Exec("insert into go_test values(:1, :2, :3)", 1337, "leet", time.Now()); err != nil {
		t.Error(err)
	}

	stmt, err := db.Prepare("insert into go_test values(:1, :2, :3)")
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 5; i++ {
			if _, err = stmt.Exec(i, "#"+fmt.Sprint(i), time.Now()); err != nil {
				t.Error(err)
				break
			}
		}

		if err = stmt.Close(); err != nil {
			t.Error(err)
		}
	}

	if _, err = db.Exec("delete go_test where id = :1", 2); err != nil {
		t.Error(err)
	}

	if _, err = db.Exec("delete go_test where name = :1", "leet"); err != nil {
		t.Error(err)
	}

	r, err := db.Query("select t.rowid, t.* from go_test t")
	if err != nil {
		t.Error(err)
	} else {
		if err = clitable.Print(r); err != nil {
			t.Error(err)
		}
	}

	r, err = db.Query("SELECT column_name as name, nullable, concat(concat(concat(data_type,'('),data_length),')') as type FROM user_tab_columns WHERE table_name= upper(:1)", "go_test")
	if err != nil {
		t.Error(err)
	} else {
		if err = clitable.Print(r); err != nil {
			t.Error(err)
		}
	}

	stmt, err = db.Prepare("insert into go_test values(:1, :2, :3)")
	if err != nil {
		t.Error(err)
	} else {
		now := time.Now()
		for i := 0; i < 10; i++ {
			// executing insert

			if _, err = stmt.Exec(100000+i, "#"+fmt.Sprint(i), time.Now()); err != nil {
				t.Error(err)
				break
			}
		}
		t.Log(time.Since(now).Seconds())

		if err = stmt.Close(); err != nil {
			t.Error(err)
		}
	}

	tx, _ := db.Begin()
	db.Exec("TRUNCATE TABLE go_test")
	db.Exec("INSERT INTO go_test (id) VALUES(:1)", 123)
	tx.Rollback()
	row := tx.QueryRow("SELECT count(1) FROM go_test")
	var cnt int64
	row.Scan(&cnt)
	if cnt != 0 {
		t.Error("transaction rollback not working!")
	}

	if _, err = db.Exec("drop table go_test"); err != nil {
		t.Error(err)
		return
	}
}

func TestDatabase(t *testing.T) {
	db, err := sql.Open("ora", testDBConnectString)
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

	if err = db.Close(); err != nil {
		t.Error(err)
		return
	}
}

func TestErrors(t *testing.T) {
	db, err := sql.Open("ora", testDBConnectString)
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
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		b.Error(err)
		return
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {

		b.StartTimer()
		rows, err := db.Query("select dummy from dual connect by level <= :1", 0)
		if err != nil {
			b.Error(err)
			return
		}
		b.StopTimer()

		rows.Close()
	}
	//b.Log(count)
}

func BenchmarkFetch(b *testing.B) {
	b.StopTimer()
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		b.Error(err)
		return
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select dummy from dual connect by level <= :1", 0)
		if err != nil {
			b.Error(err)
			return
		}

		b.StartTimer()

		count := 0
		var val string
		for rows.Next() {
			if err := rows.Scan(&val); err != nil {
				b.Error(err)
				rows.Close()
				return
			}
			count++
		}

		b.StopTimer()
		rows.Close()
	}
	//b.Log(count)
}
