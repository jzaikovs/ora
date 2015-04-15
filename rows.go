package ora

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"time"
)

// Rows implements handling rowset result from database
type Rows struct {
	stmt    *Statement
	columns []string
	descr   []*descriptor
}

// Next fetches rows from database and stores in destionation slice
func (rows *Rows) Next(dest []driver.Value) (err error) {
	// 1. fetch result in result binds,
	// TODO: manipulate fetch_size - prefech
	ret, _, _ := oci_OCIStmtFetch2.Call(rows.stmt.ptr, rows.stmt.conn.err.ptr, 1, OCI_DEFAULT, 0, OCI_DEFAULT)
	switch int16(ret) {
	case OCI_SUCCESS:
		// skip
	case OCI_NO_DATA:
		return sql.ErrNoRows
	default:
		if err = rows.stmt.conn.cerr(ret, 0, nil); err != nil {
			return
		}
	}

	// 2. store result from binds to destination values
	for i, d := range rows.descr {
		switch d.typ {
		case OCI_TYP_ROWID:
			dest[i] = string(d.valPtr.([]byte))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			if d.ind != 0 {
				dest[i] = sql.NullString{}
				break
			}
			buf := d.valPtr.([]byte)
			n := bytes.IndexByte(buf, 0) // find null byte
			dest[i] = string(buf[:n])
		case OCI_TYP_NUMBER:
			if d.ind != 0 {
				dest[i] = sql.NullFloat64{}
				break
			}
			x := d.valPtr.(*float64)
			dest[i] = *x
		case OCI_TYP_DATE:
			if d.ind != 0 {
				dest[i] = nil
				break
			}
			// convert from oracle date to time.Time
			//docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16288
			p := d.valPtr.([]byte)
			year := int(p[0]-100)*100 + int(p[1]-100)
			month := time.Month(int(p[2]))
			dest[i] = time.Date(year, month, int(p[3]), int(p[4]-1), int(p[5]-1), int(p[6]-1), 0, time.Local)
		}
	}
	return nil
}

// Close closes rowset handle
func (rows *Rows) Close() error {
	return rows.stmt.Close()
}

// Columns returns returned rowset column names
func (rows *Rows) Columns() []string {
	return rows.columns
}
