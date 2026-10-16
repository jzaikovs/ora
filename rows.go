package ora

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"unsafe"
)

// Rows implements handling query result from database
type Rows struct {
	ctx         context.Context
	closed      bool
	conn        *Conn
	stmt        *Statement
	columns     []string
	descriptors []*Descriptor
	ownsStmt    bool // statement was created only for this query and is closed with rows
}

func newRows(stmt *Statement) (*Rows, error) {
	var (
		columns     []string      // collect all columns names we will need them for database/sql
		descriptors []*Descriptor // collect all description handles we will need them to fetch binds
	)

	// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

	var count uint32
	if err := stmt.conn.cerr(oci_OCIAttrGet.Call(
		stmt.ptr,
		OCI_HTYPE_STMT,
		uintptr(unsafe.Pointer(&count)),
		0,
		OCI_ATTR_PARAM_COUNT,
		stmt.conn.err.ptr,
	)); err != nil {
		return nil, err
	}
	for pos := 1; pos <= int(count); pos++ {
		d, err := newDescriptor(stmt, pos)
		if err != nil {
			return nil, err
		}
		columns = append(columns, d.name)
		switch d.typ {
		case OCI_TYP_ROWID:
			buf := make([]byte, 18) // rowid at most is 18 bytes long
			err = d.define(pos, buf, len(buf), SQLT_AFC)
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			// oracle doesn't return correct size
			// if client side encoding uses more bytes than server side to encode single character
			// up to 4 client bytes per server byte (UTF-8), + null byte; scan rejects truncation
			buf := make([]byte, d.length*4+2)
			err = d.define(pos, buf, len(buf), SQLT_STR)
		case OCI_TYP_LONG:
			buf := make([]byte, MaxLongSize)
			err = d.define(pos, buf, len(buf), SQLT_LNG)
		case OCI_TYP_CLOB, OCI_TYP_BLOB:
			var lob *Lob
			if lob, err = stmt.conn.NewLob(); err == nil {
				lob.character = d.typ == OCI_TYP_CLOB
				d.valPtr = lob
				sqlt := SQLT_CLOB
				if d.typ == OCI_TYP_BLOB {
					sqlt = SQLT_BLOB
				}
				err = d.define(pos, ref(&lob.ptr), -1, sqlt)
			}
		case OCI_TYP_NUMBER:
			// Oracle numbers can be bigger than int and float
			// best thing is to cast to string
			tmp := make([]byte, 22)
			err = d.define(pos, tmp, len(tmp), SQLT_VNU)
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
	}

	return &Rows{conn: stmt.conn, stmt: stmt, columns: columns, descriptors: descriptors}, nil
}

// Next fetches rows from database and stores in destination slice
func (rows *Rows) Next(dest []driver.Value) (err error) {
	if rows.closed {
		return io.EOF
	}
	if rows.ctx != nil {
		return rows.conn.runContext(rows.ctx, func() error { return rows.next(dest) })
	}
	return rows.next(dest)
}

func (rows *Rows) next(dest []driver.Value) (err error) {
	// trace.Println("rows.Next")
	// 1. fetch result in result binds,
	// TODO: manipulate fetch_size - prefetch
	ret, ret2, err := oci_OCIStmtFetch2.Call(rows.stmt.ptr, rows.conn.err.ptr, 1, OCI_DEFAULT, 0, OCI_DEFAULT)
	switch int16(ret) {
	case OCI_SUCCESS:
		// skip
	case OCI_NO_DATA:
		return io.EOF
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
		if d.ind == -1 {
			dest[i] = nil
			continue
		}
		if d.ind != 0 {
			return fmt.Errorf("ora: truncated column %d", i+1)
		}
		switch d.typ {
		case OCI_TYP_ROWID:
			dest[i] = string(d.valPtr.([]byte))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR, OCI_TYP_LONG:
			buf := d.valPtr.([]byte)
			dest[i] = string(buf[:d.rlen])
		case OCI_TYP_CLOB, OCI_TYP_BLOB:
			lob := d.valPtr.(*Lob)
			lob.reset(true) // non-NULL: NULL returned nil above
			dest[i] = lob
		case OCI_TYP_NUMBER:
			v := d.valPtr.([]byte)
			dest[i] = convertNumberToString(v[1 : int(v[0])+1])
		case OCI_TYP_DATE:
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

// Close closes statement if rows own it, prepared statements stay open for reuse
func (rows *Rows) Close() error {
	if rows.closed {
		return nil
	}
	rows.closed = true
	// trace.Println("rows.Close")
	if !rows.ownsStmt || rows.stmt.closed {
		return nil
	}
	return rows.stmt.Close()
}

// Columns returns returned rowset column names
func (rows *Rows) Columns() []string {
	return rows.columns
}
