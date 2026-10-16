//go:build oracle_integration
// +build oracle_integration

package ora

import (
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testNumberTo(t *testing.T, v interface{}) bool {
	//fmt.Println(v)

	row := db.QueryRow("SELECT to_number(:1), to_char(to_number(:1)) FROM dual", v)
	if row == nil {
		t.Error("Should have fetched row")
		return false
	}

	var (
		col1 string
		col2 string
	)

	if err := row.Scan(&col1, &col2); err != nil {
		t.Error(err)
		return false
	}

	//if col1 != col2 {
	//	t.Error("data lost", col1, col2, v)
	//	return false
	//}

	if col1 != fmt.Sprint(v) {
		t.Error("data lost", col1, col2, v)
		return false
	}
	return true
}

func TestNumberCastingRandomPositiveIntegers(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	for digits := 1; digits < 19; digits++ {
		for k := 1; k < 10; k++ {
			v := rng.Intn(int(math.Pow10(digits)))
			if !testNumberTo(t, v) {
				return
			}
		}
	}
}

func TestNumberCastingRandomNegativeIntegers(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	for digits := 1; digits < 19; digits++ {
		for k := 1; k < 10; k++ {
			v := -rng.Intn(int(math.Pow10(digits)))
			if !testNumberTo(t, v) {
				return
			}
		}
	}
}

func TestNumberCastFloat(t *testing.T) {
	data := []interface{}{
		12341234,
		1234.1234,
		67.1234,
		7.1234,
		0.1234,
		0.01234,
		0.001234,
		-0.0001234,
		-0.1234,
		-0.01234,
		-0.001234,
		-0.0001234,
		-12341234,
		-1234.1234,
		-67.1234,
		0,
		0.0,
		1.1,
		0.111111,
		1.11111,
		11.1111,
		111111.1,
	}

	for _, v := range data {
		testNumberTo(t, v)
	}
}

// func TestOracleNumberBind(t *testing.T) {
// 	row := db.QueryRow("SELECT 3.1415926535897 as pi FROM dual")
// 	if row == nil {
// 		t.Error("Should have fetched row")
// 		return
// 	}

// 	var col1 Number
// 	if err := row.Scan(&col1); err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if col1.String() != "3.1415926535897" {
// 		t.Error("data lost")
// 	}
// }

func testNumberCastStrict(t *testing.T, v interface{}) {
	t.Helper()
	number := "to_number(:1)"
	if text, ok := v.(string); ok {
		parts := strings.Split(strings.TrimPrefix(text, "-"), ".")
		format := strings.Repeat("9", len(parts[0]))
		if len(parts) == 2 {
			format += "D" + strings.Repeat("9", len(parts[1]))
		}
		number = "to_number(:1, '" + format + "', 'NLS_NUMERIC_CHARACTERS=''.,''')"
	}
	// Fix the SQL conversion locale without changing the database session's NLS settings.
	query := "SELECT " + number + ", to_char(" + number + ", 'TM9', 'NLS_NUMERIC_CHARACTERS=''.,''') FROM dual"
	row := db.QueryRow(query, v)
	if row == nil {
		t.Error("Should have fetched row")
		return
	}

	var col1 string
	var col2 string
	if err := row.Scan(&col1, &col2); err != nil {
		t.Errorf("convert %v: %v", v, err)
		return
	}

	if strings.HasPrefix(col2, "-.") {
		col2 = "-0." + col2[2:]
	}

	if strings.HasPrefix(col2, ".") {
		col2 = "0" + col2
	}

	f1, err := strconv.ParseFloat(col1, 64)
	if err != nil {
		t.Error(err, col1)
		return
	}

	f2, err := strconv.ParseFloat(col2, 64)
	if err != nil {
		t.Error(err, col2)
		return
	}

	if math.Abs(f2-f1) > 0.00000001 {
		t.Error("float parsing failed")
		return
	}

	expect, _, err := (&big.Float{}).Parse(col1, 10)
	if err != nil {
		t.Error(err, col1)
		return
	}
	got, _, err := (&big.Float{}).Parse(col2, 10)
	if err != nil {
		t.Error(err, col2)
		return
	}

	if expect.Cmp(got) != 0 {
		t.Error("Data loss detected in type casting", expect, got)
	}
}

func TestTypeCastingNumbers(t *testing.T) {
	values := []interface{}{
		0, 1, 2, -5, 6, 9, 10, -11, 19, 20, 21, 22, -23, 98, 99,
		100, -101,
		"0.1", "0.01", "0.0001",
		"-54.12345",
		"-0.12345",
		102, -111, 127, 199, -200, -201, 234, -299, 300, 301, 1.1,
		-1.1123, 0, 5, -5, 2767, -2676, 100000, 1234567, "0.00000001", "-3.141592653589793238462643383279502884197169399375",
	}

	for _, v := range values {
		testNumberCastStrict(t, v)
	}
}

func TestTypeCastingRandomNumbers(t *testing.T) {
	got := &big.Float{}
	expect := &big.Float{}

	tries := 100
	for tries > 0 {
		sign := ""
		if rand.Intn(2) > 0 {
			sign = "-"
		}

		part1 := rand.Int63()
		if tries > 90 {
			part1 = 0
		}

		v := fmt.Sprintf("%s%d.%d", sign, part1, rand.Uint32())

		if _, _, err := expect.Parse(v, 10); err != nil {
			t.Fatal(err)
		}

		row := db.QueryRow("SELECT to_number(:1, '99999999999999999999.99999999999999999') FROM dual", v)
		if row == nil {
			t.Error("Should have fetched row")
			return
		}

		var col1 string
		if err := row.Scan(&col1); err != nil {
			t.Error(err)
			return
		}

		if _, _, err := got.Parse(col1, 10); err != nil {
			t.Fatal(err)
		}

		if expect.Cmp(got) != 0 {
			t.Error("Data loss detected in type casting", col1, fmt.Sprint(v))
		}
		tries--
	}
}

func TestEmptyStringIsNullInOracle(t *testing.T) {
	row := db.QueryRow("SELECT '', cast(null as varchar2(10)) FROM dual")

	var col1, col2 sql.NullString

	if err := row.Scan(&col1, &col2); err != nil {
		t.Error(err)
		return
	}

	if col1.Valid {
		t.Error("Empty string should invalid int NullString")
		return
	}

	if col2.Valid {
		t.Error("Null string should invalid int NullString")
		return
	}
}

func TestNullFetch(t *testing.T) {
	rows, err := db.Query("SELECT level, case when level = 2 then level else cast(null as number) end FROM dual connect by level <= 2")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() { _ = rows.Close() }()

	var col1, col2 *sql.NullInt64

	rows.Next()
	if err := rows.Scan(&col1, &col2); err != nil {
		t.Error(err)
		return
	}

	if col2 != nil {
		t.Error("1st fetch should be null")
		return
	}

	rows.Next()
	if err := rows.Scan(&col1, &col2); err != nil {
		t.Error(err)
		return
	}

	if col2 == nil {
		t.Error("2nd fetch should not null")
		return
	}
}

func BenchmarkFetchString(b *testing.B) {
	b.StopTimer()
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		b.Error(err)
		return
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT dummy FROM dual CONNECT BY level <= :1", b.N)
	if err != nil {
		b.Error(err)
		return
	}
	b.StartTimer()
	for rows.Next() {
		var col1 string
		if err := rows.Scan(&col1); err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
	if err := rows.Close(); err != nil {
		b.Error(err)
	}
}

func BenchmarkFetchNumber(b *testing.B) {
	b.StopTimer()
	db, err := sql.Open("ora", testDBConnectString)
	if err != nil {
		b.Error(err)
		return
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT -3.141592653589793238462643383279502884197169399375 FROM dual CONNECT BY level <= :1", b.N)
	if err != nil {
		b.Error(err)
		return
	}
	defer func() { _ = rows.Close() }()

	b.StartTimer()
	for rows.Next() {
		var col1 string
		if err := rows.Scan(&col1); err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
}
