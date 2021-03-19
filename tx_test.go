package ora

import (
	"database/sql"
	"testing"
)

// type queriyer interface {
// 	Query(query string, args ...interface{}) (*sql.Rows, error)
// }

// func query2(t *testing.T, q queriyer, sql string, itr string) bool {
// 	result, err := q.Query(sql)
// 	if err != nil {
// 		t.Error(err)
// 		return false
// 	}

// 	for result.Next() {
// 		itr()
// 	}

// 	return true
// }

// func oneRow(q queriyer, sql string) {

// }

func TestTxStartEnd(t *testing.T) {
	// db.Exec("INSERT INTO go_test (int64, float64) VALUES (3.000, 3.14159)")
	// query("SELECT dbms_transaction.local_transaction_id FROM dual", func(row *sql.Rows) error {
	// 	var id string
	// 	row.Scan(&id)
	// 	t.Log(id)
	// 	return nil
	// })

	tx, err := db.Begin()
	if err != nil {
		t.Error(err)
		return
	}

	// tx.Exec("INSERT INTO go_test (int64, float64) VALUES (3.000, 3.14159)")

	row, err := tx.Query("SELECT dbms_transaction.local_transaction_id FROM dual")
	if err != nil {
		t.Error(err)
		return
	}
	defer row.Close()

	if !row.Next() {
		t.Error("No row when expected")
		return
	}

	var id string
	if err = row.Scan(&id); err != nil {
		t.Error(id)
		return
	}

	t.Log(id)

	if err = tx.Rollback(); err != nil {
		t.Error(err)
		return
	}
}

func TestTxCommit(t *testing.T) {
	db.Exec("TRUNCATE TABLE go_test")

	for k := 0; k < 1000; k++ {

		tx, err := db.Begin()
		if err != nil {
			t.Error(err)
			return
		}

		_, err = tx.Exec("INSERT INTO go_test (int64, float64) VALUES (3.000, 3.14159)")
		if err != nil {
			t.Log(err)
			return
		}

		other, err := sql.Open("ora", testDBConnectString)
		if err != nil {
			t.Error(err)
			return
		}
		defer other.Close()

		cnt := 0
		if err = other.QueryRow("SELECT count(1) FROM go_test").Scan(&cnt); err != nil {
			t.Error(err)
			return
		}

		if cnt > 0 {
			t.Error("tx not working", cnt)
			return
		}

		tx.Commit()

		if err = other.QueryRow("SELECT count(1) FROM go_test").Scan(&cnt); err != nil {
			t.Error(err)
			return
		}

		if cnt != 1 {
			t.Error("after commit rows", cnt)
			return
		}
	}
}
