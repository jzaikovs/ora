package ora

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jzaikovs/clitable"
)

/*
Setup for testing:
```
create user ora_go_test identified by ora_go_test_password;
grant connect, resource to ora_go_test;
```
*/

var (
	testDBConnectString = "ora_go_test/ora_go_test_password@//oracle:1521/XE"
	db                  *sql.DB
)

func setup() (err error) {
	fmt.Println("setup...")
	_, err = db.Exec("create table go_test(id number, name varchar2(32), date_bind date, lobcol clob)")
	return
}

func cleanup() (err error) {
	fmt.Println("cleanup...")
	_, err = db.Exec("drop table go_test")
	db.Close()
	return

}

func TestMain(m *testing.M) {
	// your func
	var err error
	db, err = sql.Open("ora", testDBConnectString)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = setup(); err != nil {
		fmt.Println(err)
	}

	retCode := m.Run()

	cleanup()

	// call with result of m.Run()
	os.Exit(retCode)
}

func TestQuery(t *testing.T) {
	conn, err := Open(testDBConnectString)
	if err != nil {
		t.Error(err)
		return
	}

	r, err := conn.Query("select * from dual")
	if err != nil {
		t.Error(err)
		return
	}

	defer r.Close()

	count := 0

	for r.Next() == nil {
		v, err := r.Values()
		if err != nil {
			t.Error(err)
			return
		}
		if fmt.Sprint(v[0]) != "X" {
			t.Error("dual not returning X")
		}
		count++
	}

	if count == 0 {
		t.Error("Conn.Query() return no rows on dual query")
	}
}

func TestExecInsert(t *testing.T) {
	var err error

	if _, err = db.Exec("insert into go_test (id, name, date_bind) values(:1, :2, :3)", 1337, "leet", time.Now()); err != nil {
		t.Error(err)
	}
}

func TestPrepareInsert(t *testing.T) {
	db.Exec("truncate table go_test")

	stmt, err := db.Prepare("insert into go_test (id, name, date_bind) values(:1, :2, :3)")
	if err != nil {
		t.Error(err)
		return
	}

	n := 5

	for i := 0; i < n; i++ {
		if _, err = stmt.Exec(i, "#"+fmt.Sprint(i), time.Now()); err != nil {
			t.Error(err)
			break
		}
	}

	if err = stmt.Close(); err != nil {
		t.Error(err)
	}

	var cnt float64
	row := db.QueryRow("select count(1) from go_test")
	row.Scan(&cnt)

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

	now := time.Now()
	for i := 0; i < 10; i++ { // executing insert
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

func TestQuery2(t *testing.T) {
	TestPrepareInsert(t)

	r, err := db.Query("select t.rowid, t.id, name, date_bind from go_test t")
	if err != nil {
		t.Error(err)
		return
	}

	if err = clitable.Print(r); err != nil {
		t.Error(err)
	}
}

func TestQuery3(t *testing.T) {
	r, err := db.Query("SELECT column_name as name, nullable, concat(concat(concat(data_type,'('),data_length),')') as type FROM user_tab_columns WHERE table_name= upper(:1)", "go_test")
	if err != nil {
		t.Error(err)
		return
	}

	if err = clitable.Print(r); err != nil {
		t.Error(err)
	}
}

func TestTransactions(t *testing.T) {
	tx, err := db.Begin()
	if err != nil {
		t.Error(err)
		return
	}
	db.Exec("TRUNCATE TABLE go_test")
	db.Exec("INSERT INTO go_test (id) VALUES(:1)", 123)
	tx.Rollback()

	row := tx.QueryRow("SELECT count(1) FROM go_test")
	var cnt int64
	row.Scan(&cnt)
	if cnt != 0 {
		t.Error("transaction rollback not working!")
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

func query(sql string, fn func(*sql.Rows) error, binds ...interface{}) (err error) {
	result, err := db.Query(sql, binds...)
	if err != nil {
		return err
	}

	defer result.Close()

	for result.Next() {
		if err = fn(result); err != nil {
			return err
		}
	}

	return
}
