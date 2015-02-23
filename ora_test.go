package ora

import (
	"database/sql"
	"fmt"
	"github.com/jzaikovs/clitable"
	"testing"
)

var (
	DB_ACCESS = "ora_go_test/ora_go_test_password@//oracle:1521/XE"
)

/*
Setup for testing:
```
create user ora_go_test identified by ora_go_test_password;
grant connect, resource to ora_go_test;
```

*/

func TestExec(t *testing.T) {
	db, err := sql.Open("ora", DB_ACCESS)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	if _, err = db.Exec("create table go_test(id number, name varchar2(32))"); err != nil {
		t.Error(err)
		//return
	}

	if _, err = db.Exec("insert into go_test values(:1, :2)", 1337, "leet"); err != nil {
		t.Error(err)
	}

	stmt, err := db.Prepare("insert into go_test values(:1, :2)")
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 5; i++ {
			if _, err = stmt.Exec(i, "#"+fmt.Sprint(i)); err != nil {
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

	/*
		db2, err := sql.Open("ora", DB_ACCESS)
		if err != nil {
			t.Error(err)
		}

		r, err = db.Query("select * from v$session where username = user")
		if err != nil {
			t.Error(err)
		} else {
			if err = clitable.Print(r); err != nil {
				t.Error(err)
			}
		}

		if err = db2.Close(); err != nil {
			t.Error(err)
		}
	*/

	if _, err = db.Exec("drop table go_test"); err != nil {
		t.Error(err)
		return
	}
}

func TestDatabase(t *testing.T) {
	db, err := sql.Open("ora", DB_ACCESS)
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
	db, err := sql.Open("ora", DB_ACCESS)
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
	db, err := sql.Open("ora", DB_ACCESS)
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
