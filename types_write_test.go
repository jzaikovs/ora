//go:build oracle_write_integration
// +build oracle_write_integration

package ora

import (
	"math"
	"testing"
)

func TestNumberFormatting(t *testing.T) {
	if _, err := db.Exec("DELETE go_test"); err != nil {
		t.Fatal(err)
	}

	res, err := db.Exec("INSERT INTO go_test (int64, float64) VALUES (3.000, 3.14159)")
	if err != nil {
		t.Log(err)
		return
	}

	if cnt, err := res.RowsAffected(); err != nil || cnt != 1 {
		t.Log(err)
		t.Error("Incorrect rows affected value after INSERT")
		return
	}

	var (
		i64         int64
		f64, f64int float64
	)

	if err := db.QueryRow("SELECT int64, float64, int64 as int_as_float FROM go_test").Scan(&i64, &f64, &f64int); err != nil {
		t.Error(err)
		return
	}

	expectInt := int64(3)
	if i64 != expectInt {
		t.Errorf("Incorrect INTEGER formating (%v) != (%v)", i64, expectInt)
		return
	}

	const TOLERANCE = 0.00000001

	expect := 3.142
	if diff := math.Abs(f64 - expect); diff > TOLERANCE {
		t.Errorf("Incorrect FLOAT formating (%v) - (%v) = (%f)", f64, expect, diff)
		return
	}

	expect = 3.0
	if diff := math.Abs(f64int - expect); diff > TOLERANCE {
		t.Errorf("Incorrect FLOAT casting (%v) - (%v) = (%f)", f64int, expect, diff)
		return
	}
}
