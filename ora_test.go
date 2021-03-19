package ora

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

/*
Setup for testing:
```
CREATE USER ora_go_test IDENTIFIED BY ora_go_test_password;
GRANT CONNECT, RESOURCES TO ora_go_test;
ALTER USER ora_go_test QUOTA 100M ON users;
GRANT SELECT ON v_$session TO ora_go_test;
GRANT EXECUTE ON dbms_lock TO ora_go_test;
```
*/

var (
	dbUser              = "ORA_GO_TEST"
	dbPassword          = "ora_go_test_password"
	dbHost              = "//localhost:1521/XE"
	testDBConnectString = fmt.Sprintf("%s/%s@%s", dbUser, dbPassword, dbHost)
	db                  *sql.DB
)

func setup() (err error) {
	// fmt.Println("setup...")
	_, err = db.Exec("CREATE TABLE go_test(id number, int64 number(9), float64 number(9, 3), name varchar2(32), date_bind date, lobcol clob, data blob)")
	return
}

func tearDown() (err error) {
	// fmt.Println("tearDown...")
	_, err = db.Exec("DROP TABLE go_test")
	db.Close()
	return
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

	tearDown()

	if handleRefCount > 0 {
		panic("OCI handles not freed")
	}

	// call with result of m.Run()
	os.Exit(retCode)
}

func TestQuery(t *testing.T) {
	rows, err := db.Query("select * from dual")
	if err != nil {
		t.Error(err)
		return
	}

	defer rows.Close()

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

// func TestQuery2(t *testing.T) {
// 	TestPrepareInsert(t)

// 	r, err := db.Query("select t.rowid, t.id, name, date_bind from go_test t")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if err = clitable.Print(r); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestQuery3(t *testing.T) {
// 	r, err := db.Query("SELECT column_name as name, nullable, concat(concat(concat(data_type,'('),data_length),')') as type FROM user_tab_columns WHERE table_name= upper(:1)", "go_test")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if err = clitable.Print(r); err != nil {
// 		t.Error(err)
// 	}
// }

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
	defer db.Close()

	rows, err := db.Query("SELECT t.rowid, 'hello' as greet, dummy, 1337 as leet, sysdate as today from dual t connect by level <= :1", 1)
	if err != nil {
		t.Error(err)
		return
	}
	defer rows.Close()

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
	defer db.Close()

	_, err = db.Query("select x from dual_non_existing_table")
	if err == nil || err.Error() != "ORA-00942: table or view does not exist" {
		t.Error("should raise error: ORA-00942: table or view does not exist")
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
			rows.Close()
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
	defer conn.Close()

	rows, err := conn.Query("SELECT 66 as x, :1 FROM dual", 99)
	if err != nil {
		t.Error(err)
		return
	}
	defer rows.Close()

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
