package ora

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestLobHandleNulls(t *testing.T) {
	if _, err := db.Exec("delete go_test"); err != nil {
		t.Error(err)
		return
	}

	expected := []string{
		"Hello world!",
		"",
	}

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
			lob  Lob
		)

		if err = row.Scan(&id, &name, &lob); err != nil {
			t.Error(err)
			return
		}

		val, err := lob.String()
		if err != nil {
			t.Error(err)
		}

		t.Log(val)

		if val != expected[0] {
			t.Error("clob fetch not working", val, expected[0])
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
	if _, err := db.Exec(`declare c clob; begin c := :1; c := c || :1;
							 insert into go_test (id, lobcol) values(600, c);
							 insert into go_test (id, lobcol) values(600, c); end;`, val); err != nil {
		t.Error(err)
		return
	}

	var lob Lob

	query("select lobcol from go_test where id = 600", func(rows *sql.Rows) error {
		return rows.Scan(&lob)
	})

	s, _ := lob.String()
	if val+val != s {
		t.Errorf("lob.Read not returned expected")
	}
}

func TestLargerBlob(t *testing.T) {
	if _, err := db.Exec("DELETE go_test"); err != nil {
		t.Error(err)
		return
	}

	hashes := make([]string, 10)

	for i := range hashes {
		buf := make([]byte, 1024*1024*1) // 10MB
		rand.New(rand.NewSource(time.Now().Unix())).Read(buf)
		hash := sha1.Sum(buf)
		hashes[i] = hex.EncodeToString(hash[:])

		if _, err := db.Exec("INSERT INTO go_test(data) VALUES(:1)", buf); err != nil {
			t.Error(err)
			return
		}
	}

	row, err := db.Query("SELECT rawtohex(dbms_crypto.hash(data, 3)) FROM go_test")
	if err != nil {
		t.Error(err)
		return
	}
	defer row.Close()

	for _, h := range hashes {
		if !row.Next() {
			t.Error("Expected row")
			return
		}

		var col1 string
		if err := row.Scan(&col1); err != nil {
			t.Error(err)
		}

		if strings.ToLower(col1) != h {
			t.Error("data lost at save")
		}
	}
}
