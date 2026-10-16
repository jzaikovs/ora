//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestReadOnlyTransaction(t *testing.T) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()
	for _, finish := range []string{"rollback", "commit"} {
		t.Run(finish, func(t *testing.T) {
			tx, err := conn.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
			if err != nil {
				t.Fatalf("begin read-only transaction: %v", err)
			}
			defer func() { _ = tx.Rollback() }()
			var value int
			if err := tx.QueryRowContext(context.Background(), "SELECT :1 FROM dual", 42).Scan(&value); err != nil {
				t.Fatal(err)
			}
			if value != 42 {
				t.Fatalf("value = %d, want 42", value)
			}
			if finish == "commit" {
				err = tx.Commit()
			} else {
				err = tx.Rollback()
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReadOnlyQueries(t *testing.T) {
	for _, test := range []struct {
		name      string
		query     string
		args      []interface{}
		maxRows   int
		columns   []string
		rows      [][]interface{}
		truncated bool
	}{
		{
			name: "bind_precision_and_types",
			query: "select :1 as N, :2 as TXT, cast(null as varchar2(10)) as EMPTY, " +
				"date '2000-01-02' as D from dual",
			args:    []interface{}{int64(9007199254740993), "āčē"},
			columns: []string{"N", "TXT", "EMPTY", "D"},
			rows: [][]interface{}{{
				"9007199254740993", "āčē", nil,
				time.Date(2000, time.January, 2, 0, 0, 0, 0, time.Local),
			}},
		},
		{
			name:      "row_limit",
			query:     "select level as N from dual connect by level <= 3 order by N",
			maxRows:   2,
			columns:   []string{"N"},
			rows:      [][]interface{}{{"1"}, {"2"}},
			truncated: true,
		},
		{
			name:    "empty_result",
			query:   "select 1 as N from dual where 1=0",
			columns: []string{"N"},
			rows:    [][]interface{}{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			rows, err := tx.QueryContext(ctx, test.query, test.args...)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = rows.Close() }()
			columns, err := rows.Columns()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(columns, test.columns) {
				t.Fatalf("columns = %v, want %v", columns, test.columns)
			}
			got := [][]interface{}{}
			var truncated bool
			for rows.Next() {
				// The consumer owns the limit; one extra row distinguishes truncation from EOF.
				if test.maxRows > 0 && len(got) == test.maxRows {
					truncated = true
					break
				}
				values := make([]interface{}, len(columns))
				dest := make([]interface{}, len(columns))
				for i := range dest {
					dest[i] = &values[i]
				}
				if err := rows.Scan(dest...); err != nil {
					t.Fatal(err)
				}
				got = append(got, values)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.rows) || truncated != test.truncated {
				t.Fatalf("rows = %#v, truncated = %t; want %#v, %t", got, truncated, test.rows, test.truncated)
			}
			if err := rows.Close(); err != nil {
				t.Fatal(err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
