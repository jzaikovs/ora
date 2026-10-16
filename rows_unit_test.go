package ora

import (
	"database/sql/driver"
	"testing"
)

func TestRowsPreserveNullAndDetectTruncation(t *testing.T) {
	for _, typ := range []int{OCI_TYP_VARCHAR, OCI_TYP_CHAR, OCI_TYP_LONG, OCI_TYP_NUMBER, OCI_TYP_DATE} {
		d := &Descriptor{typ: typ, ind: -1}
		rows := &Rows{descriptors: []*Descriptor{d}}
		dest := []driver.Value{"previous"}
		if err := rows.scan(dest); err != nil || dest[0] != nil {
			t.Fatalf("type %d: value=%v err=%v", typ, dest[0], err)
		}
		d.ind = 8
		if err := rows.scan(dest); err == nil {
			t.Fatalf("type %d: truncation accepted", typ)
		}
	}
}

func TestRowsCloseIdempotent(t *testing.T) {
	rows := &Rows{}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
}
