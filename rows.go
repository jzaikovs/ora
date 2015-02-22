package ora

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"time"
)

type t_rows struct {
	stmt    *t_stmt
	columns []string
	descr   []t_descriptor
	count   int
}

func (self *t_rows) bind_result() (err error) {
	for i, d := range self.descr {
		switch d.typ() {
		case OCI_TYP_ROWID:
			buf := d.val_ptr.([]byte)
			err = self.stmt.conn.cerr(oci_OCIDefineByPos.Call(self.stmt.ptr, ref(&d.ptr), self.stmt.conn.err.ptr, uintptr(i+1), buf_addr(buf), 18, SQLT_AFC, int_ref(&self.descr[i].ind), 0, 0, 0))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			buf := d.val_ptr.([]byte)
			err = self.stmt.conn.cerr(oci_OCIDefineByPos.Call(self.stmt.ptr, ref(&d.ptr), self.stmt.conn.err.ptr, uintptr(i+1), buf_addr(buf), uintptr(len(buf)), SQLT_STR, int_ref(&self.descr[i].ind), 0, 0, 0))
		case OCI_TYP_NUMBER:
			err = self.stmt.conn.cerr(oci_OCIDefineByPos.Call(self.stmt.ptr, ref(&d.ptr), self.stmt.conn.err.ptr, uintptr(i+1), float64_ref(d.val_ptr.(*float64)), 8, SQLT_FLT, int_ref(&self.descr[i].ind), 0, 0, 0))
		case OCI_TYP_DATE:
			buf := d.val_ptr.([]byte)
			err = self.stmt.conn.cerr(oci_OCIDefineByPos.Call(self.stmt.ptr, ref(&d.ptr), self.stmt.conn.err.ptr, uintptr(i+1), buf_addr(buf), uintptr(len(buf)), SQLT_DAT, int_ref(&self.descr[i].ind), 0, 0, 0))
		}
		if err != nil {
			logLine(err)
			return
		}
	}
	return
}

func (self *t_rows) Next(dest []driver.Value) (err error) {
	// 1. we set destination types and pointers to them
	if err = self.bind_result(); err != nil {
		return
	}

	// 2. fetch result in result binds, TODO: manipulate fetch_size
	ret, _, _ := oci_OCIStmtFetch2.Call(self.stmt.ptr, self.stmt.conn.err.ptr, 1, OCI_DEFAULT, 0, OCI_DEFAULT)
	if ret == 100 {
		self.count++
		return sql.ErrNoRows
	} else if ret != 0 {
		if err := self.stmt.conn.handle_error(); err != nil {
			return err
		}
	}

	// 3. store result from binds to destination values
	for i, d := range self.descr {
		switch d.typ() {
		case OCI_TYP_ROWID:
			dest[i] = string(d.val_ptr.([]byte))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			if d.ind != 0 {
				dest[i] = sql.NullString{}
				break
			}
			buf := d.val_ptr.([]byte)
			n := bytes.IndexByte(buf, 0) // find null byte
			dest[i] = string(buf[:n])
		case OCI_TYP_NUMBER:
			if d.ind != 0 {
				dest[i] = sql.NullFloat64{}
				break
			}
			x := d.val_ptr.(*float64)
			dest[i] = *x
		case OCI_TYP_DATE:
			if d.ind != 0 {
				dest[i] = nil
				break
			}

			// convert from oracle date to time.Time
			//docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16288
			p := d.val_ptr.([]byte)
			year := int(p[0]-100)*100 + int(p[1]-100)
			month := time.Month(int(p[2]))
			dest[i] = time.Date(year, month, int(p[3]), int(p[4]-1), int(p[5]-1), int(p[6]-1), 0, time.Local)
		}
	}

	return nil
}

func (self *t_rows) Close() error {
	return self.stmt.Close()
}

func (self *t_rows) Columns() []string {
	return self.columns
}
