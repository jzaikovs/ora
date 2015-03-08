package ora

import (
	"database/sql/driver"
	"time"
)

type t_stmt struct {
	*oci_handle
	conn   *t_conn
	tx     *t_tx
	binds  []interface{} // will hold pointers to every bind variable
	closed bool
}

func (self *t_stmt) Close() error {
	if self.closed {
		return nil
	}
	self.closed = true
	return self.conn.cerr(oci_OCIStmtRelease.Call(self.ptr, self.conn.err.ptr, 0, 0, OCI_DEFAULT))
}

func (self *t_stmt) NumInput() int {
	return -1 // TODO: count input parameters
}

func (self *t_stmt) Exec(args []driver.Value) (driver.Result, error) {
	if err := self.bind(args); err != nil {
		return nil, err
	}

	if err := self.exec(1); err != nil {
		return nil, err
	}

	// closing this is failing with invalid handle error,
	// OCIStmtExecute on non SELECT will close handle, can't find info about this case
	//if err := self.Close(); err != nil {
	//	return nil, err
	//}

	return t_result{}, nil
}

func (self *t_stmt) bind(args []driver.Value) (err error) {
	// create link to variables, so that GC will not discard them
	self.binds = make([]interface{}, len(args))

	for i, arg := range args {
		var bnd uintptr
		//name := []byte(fmt.Sprintf(":%d", i+1))
		//n := uintptr(len(name))

		// store pointer to val in binds because garbage collector will discard val
		// and OCI will pass some random data from memory
		switch val := arg.(type) { // GC will discard val if not referenced somewhere
		case time.Time:
			p := make([]byte, 7)
			p[0] = byte(int(val.Year()/100)) + 100
			p[1] = byte(val.Year()%100) + 100
			p[2] = byte(val.Month())
			p[3] = byte(val.Day())
			p[4] = byte(val.Hour()) + 1
			p[5] = byte(val.Minute()) + 1
			p[6] = byte(val.Second()) + 1

			self.binds[i] = p
			err = self.conn.cerr(oci_OCIBindByPos.Call(self.ptr, ref(&bnd), self.conn.err.ptr, uintptr(i+1), buf_addr(p), uintptr(7), SQLT_DAT, 0, 0, 0, 0, 0, OCI_DEFAULT))
		case int64:
			x := int(val)
			self.binds[i] = &x
			err = self.conn.cerr(oci_OCIBindByPos.Call(self.ptr, ref(&bnd), self.conn.err.ptr, uintptr(i+1), int_ref(&x), uintptr(sizeof_int), SQLT_INT, 0, 0, 0, 0, 0, OCI_DEFAULT))
		case string:
			buf := append([]byte(val), 0)
			self.binds[i] = buf
			err = self.conn.cerr(oci_OCIBindByPos.Call(self.ptr, ref(&bnd), self.conn.err.ptr, uintptr(i+1), buf_addr(buf), uintptr(len(buf)), SQLT_STR, 0, 0, 0, 0, 0, OCI_DEFAULT))
		}

		if err != nil {
			self.Close()
			return
		}
	}
	return
}

func (self *t_stmt) exec(n int) (err error) {
	mode := OCI_DEFAULT // default will fetch n rows return describe, we do it only before first fetch when call define

	if self.tx == nil {
		mode = OCI_COMMIT_ON_SUCCESS
	}

	if err = self.conn.cerr(oci_OCIStmtExecute.Call(self.conn.serv.ptr, self.ptr, self.conn.err.ptr, uintptr(n), 0, 0, 0, uintptr(mode))); err != nil {
		self.Close()
	}
	return
}

func (self *t_stmt) Query(args []driver.Value) (driver.Rows, error) {
	if err := self.bind(args); err != nil {
		return nil, err
	}

	if err := self.exec(0); err != nil {
		return nil, err
	}

	columns := make([]string, 0)       // collect all columns names we will need them for database/sql
	descrs := make([]*t_descriptor, 0) // collect all description handles we will need them to fetch row

	// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

	d, err := self.newDescriptor(1)
	for err == nil {
		columns = append(columns, d.name)
		pos := len(descrs) + 1

		switch d.typ {
		case OCI_TYP_ROWID:
			buf := make([]byte, 18) // rowid at most is 18 bytes long
			err = d.define(pos, buf, len(buf), SQLT_AFC)
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			buf := make([]byte, d.length+1) // make buffer where result is stored + 1 null byte
			err = d.define(pos, buf, len(buf), SQLT_STR)
		case OCI_TYP_NUMBER:
			tmp := float64(0) // note: oracle numbers can be bigger than float
			err = d.define(pos, &tmp, sizeof_int, SQLT_FLT)
		case OCI_TYP_DATE:
			buf := make([]byte, d.length)
			err = d.define(pos, buf, len(buf), SQLT_DAT)
		}

		if err != nil {
			logLine("define result failed with err:", err)
			return nil, err
		}

		descrs = append(descrs, d)
		d, err = self.newDescriptor(len(descrs) + 1)
	}

	return &t_rows{stmt: self, columns: columns, descr: descrs}, nil
}

func (self *t_stmt) newDescriptor(pos int) (d *t_descriptor, err error) {
	d = &t_descriptor{stmt: self}
	if err = self.conn.cerr(oci_OCIParamGet.Call(self.ptr, OCI_HTYPE_STMT, self.conn.err.ptr, ref(&d.ptr), uintptr(pos))); err != nil {
		return
	}
	d.name = d.getName()
	d.typ = d.getTyp()
	d.length = d.getLen()

	return
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci17msc001.htm#i575144
func (self *t_stmt) prepare(query string) (err error) {
	buf := append([]byte(query), 0)
	if err = self.conn.cerr(oci_OCIStmtPrepare2.Call(
		self.conn.serv.ptr,
		ref(&self.ptr),
		self.conn.err.ptr,
		buf_addr(buf),
		uintptr(len(buf)),
		0,
		0,
		OCI_NTV_SYNTAX,
		0)); err != nil {
		//self.Close() // free alloc
	}
	return
}
