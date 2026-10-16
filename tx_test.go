//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestReadOnlyTransactionRejectsWrite(t *testing.T) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec("DELETE FROM go_test WHERE 1=0")
	var oracleErr Error
	if !errors.As(err, &oracleErr) || oracleErr.Code != 1456 {
		t.Fatalf("write in read-only transaction: got %v, want ORA-01456", err)
	}
}

func TestTxStartEnd(t *testing.T) {
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	var value int
	if err := tx.QueryRow("SELECT 1 FROM dual").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != 1 {
		t.Fatalf("value = %d, want 1", value)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow("SELECT 1 FROM dual").Scan(&value); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("query after rollback: got %v, want sql.ErrTxDone", err)
	}
}

func TestTxCommit(t *testing.T) {
	if _, err := db.Exec("DELETE FROM go_test"); err != nil {
		t.Fatal(err)
	}
	other, err := sql.Open(DriverName, testDBConnectString)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := other.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Two consecutive transactions exercise reuse without counting previously committed rows.
	for _, id := range []int{701, 702} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			if _, err := tx.Exec("INSERT INTO go_test (id) VALUES (:1)", id); err != nil {
				t.Fatal(err)
			}
			var count int
			query := "SELECT count(*) FROM go_test WHERE id = :1"
			if err := other.QueryRow(query, id).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatalf("uncommitted rows visible: %d", count)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			if err := other.QueryRow(query, id).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("committed rows = %d, want 1", count)
			}
		})
	}
}
