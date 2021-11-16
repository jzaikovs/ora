package ora

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

// Rows implements handling query result from database
type Rows struct {
	conn        *Conn
	stmt        *Statement
	columns     []string
	descriptors []*Descriptor
}

func newRows(stmt *Statement) (*Rows, error) {
	var (
		columns     []string      // collect all columns names we will need them for database/sql
		descriptors []*Descriptor // collect all description handles we will need them to fetch binds
	)

	// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

	d, err := newDescriptor(stmt, 1)
	for err == nil {
		columns = append(columns, d.name)
		pos := len(descriptors) + 1
		switch d.typ {
		case OCI_TYP_ROWID:
			buf := make([]byte, 18) // rowid at most is 18 bytes long
			err = d.define(pos, buf, len(buf), SQLT_AFC)
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			// oracle doesn't return correct size
			// if client side encoding uses more bytes than server side to encode single character
			buf := make([]byte, d.length*2+2) // make buffer where result is stored + null byte
			err = d.define(pos, buf, len(buf), SQLT_STR)
		case OCI_TYP_LONG:
			buf := make([]byte, MaxLongSize)
			err = d.define(pos, buf, len(buf), SQLT_LNG)
		case OCI_TYP_CLOB, OCI_TYP_BLOB:
			var lob *Lob
			if lob, err = stmt.conn.NewLob(); err == nil {
				d.valPtr = lob
				err = d.define(pos, ref(&lob.ptr), -1, SQLT_CLOB)
			}
		case OCI_TYP_NUMBER:
			// Oracle numbers can be bigger than int and float
			// best thing is to cast to string
			tmp := make([]byte, 22)
			d.define(pos, tmp, len(tmp), SQLT_VNU)
		case OCI_TYP_DATE:
			buf := make([]byte, d.length)
			err = d.define(pos, buf, len(buf), SQLT_DAT)
		default:
			err = fmt.Errorf("unsupported Oracle data type %v", d.typ)
		}

		if err != nil {
			trace.Printf("Define pos: %d, failed with err: %s", pos, err)
			return nil, err
		}

		descriptors = append(descriptors, d)
		d, err = newDescriptor(stmt, len(descriptors)+1)
	}

	return &Rows{conn: stmt.conn, stmt: stmt, columns: columns, descriptors: descriptors}, nil
}

// Next fetches rows from database and stores in destination slice
func (rows *Rows) Next(dest []driver.Value) (err error) {
	// trace.Println("rows.Next")
	// 1. fetch result in result binds,
	// TODO: manipulate fetch_size - prefetch
	ret, ret2, err := oci_OCIStmtFetch2.Call(rows.stmt.ptr, rows.conn.err.ptr, 1, OCI_DEFAULT, 0, OCI_DEFAULT)
	switch int16(ret) {
	case OCI_SUCCESS:
		// skip
	case OCI_NO_DATA:
		return sql.ErrNoRows
	default:
		if err = rows.conn.cerr(ret, ret2, err); err != nil {
			trace.Printf("OCIStmtFetch2(...) -> %s", err)
			return err
		}
	}

	// 2. store result from binds to destination values
	return rows.scan(dest)
}

func (rows *Rows) scan(dest []driver.Value) error {
	for i, d := range rows.descriptors {
		switch d.typ {
		case OCI_TYP_ROWID:
			dest[i] = string(d.valPtr.([]byte))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR, OCI_TYP_LONG:
			if d.ind != 0 {
				dest[i] = "" //  null string in oracle is empty string
				break
			}
			buf := d.valPtr.([]byte)
			dest[i] = string(buf[:d.rlen])
		case OCI_TYP_CLOB, OCI_TYP_BLOB:
			lob := d.valPtr.(*Lob)
			lob.reset(d.ind == 0)
			dest[i] = lob
		case OCI_TYP_NUMBER:
			if d.ind != 0 {
				dest[i] = nil
				break
			}

			v := d.valPtr.([]byte)
			dest[i] = convertNumberToString(v[1 : int(v[0])+1])
		case OCI_TYP_DATE:
			if d.ind != 0 {
				dest[i] = nil
				break
			}

			p := d.valPtr.([]byte)
			dest[i] = convertDateToGoTime(p)
		}
	}

	return nil
}

func (rows *Rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	if index < 0 || len(rows.descriptors) <= index {
		ok = false
		return
	}

	d := rows.descriptors[index]
	if d.typ == OCI_TYP_NUMBER {
		return int64(d.prec), int64(d.scale), true
	}

	return
}

// Close closes nothing to close
func (rows *Rows) Close() error {
	// trace.Println("rows.Close")
	return rows.stmt.Close()
}

// Columns returns returned rowset column names
func (rows *Rows) Columns() []string {
	return rows.columns
}
