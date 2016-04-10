package ora

import (
	"database/sql"
	"testing"
)

func TestLob(t *testing.T) {
	if _, err := db.Exec("delete go_test"); err != nil {
		t.Error(err)
		return
	}

	expected := []string{"helloworld", ""}

	for _, e := range expected {
		if _, err := db.Exec("insert into go_test (lobcol) values(:1)", e); err != nil {
			t.Error(err)
			return
		}
	}

	db.Exec("commit")

	err := query("SELECT id, name, lobcol FROM go_test", func(row *sql.Rows) (err error) {
		var (
			id   sql.NullFloat64
			name sql.NullString
			lob  sql.NullString
		)

		if err = row.Scan(&id, &name, &lob); err != nil {
			t.Error(err)
			return
		}

		if lob.String != expected[0] {
			t.Error("clob fetch not working")
		}
		expected = expected[1:]
		return
	})

	if err != nil {
		t.Error(err)
	}
}

func TestLobRead(t *testing.T) {
	val := ""
	for i := 0; i < 3200; i++ {
		val = val + "0123456789"
	}

	db.Exec("delete go_test")
	if _, err := db.Exec("declare c clob; begin c := :1; c := c || :1; insert into go_test (id, lobcol) values(600, c); end;", val); err != nil {
		t.Error(err)
		return
	}

	var lob sql.NullString

	query("select lobcol from go_test where id = 600", func(rows *sql.Rows) error {
		return rows.Scan(&lob)
	})

	if val+val != lob.String {
		t.Error("lob.Read not working")
	}
}
