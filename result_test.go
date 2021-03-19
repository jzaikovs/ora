package ora

import (
	"testing"
	"time"
)

func TestExecInsert(t *testing.T) {
	res, err := db.Exec("INSERT INTO go_test (id, name, date_bind) values(:1, :2, :3)", 1337, "leet", time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	if cnt, err := res.RowsAffected(); err != nil || cnt != 1 {
		t.Log(err)
		t.Error("Incorrect rows affected value after INSERT")
		return
	}
}
