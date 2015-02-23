package ora

import (
	"database/sql/driver"
	"fmt"
)

type t_stmt struct {
	*oci_handle
	conn   *t_conn
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
		name := []byte(fmt.Sprintf(":%d", i+1))
		n := uintptr(len(name))

		// store pointer to val in binds because carbage collector will discard val
		// and oci when we will execute, pased bind value will holde some random data from memory
		switch val := arg.(type) { // GC will discard val if not referenced somewhere
		case int64:
			self.binds[i] = &val
			err = self.conn.cerr(oci_OCIBindByName.Call(self.ptr, ref(&bnd), self.conn.err.ptr, buf_addr(name), n, int64_ref(&val), 8, SQLT_INT, 0, 0, 0, 0, 0, OCI_DEFAULT))
		case string:
			buf := append([]byte(val), 0)
			self.binds[i] = buf
			err = self.conn.cerr(oci_OCIBindByName.Call(self.ptr, ref(&bnd), self.conn.err.ptr, buf_addr(name), n, buf_addr(buf), uintptr(len(buf)), SQLT_STR, 0, 0, 0, 0, 0, OCI_DEFAULT))
		}

		if err != nil {
			self.Close()
			return
		}
	}
	return
}

func (self *t_stmt) exec(itr int) (err error) {
	if err = self.conn.cerr(oci_OCIStmtExecute.Call(self.conn.serv.ptr, self.ptr, self.conn.err.ptr, uintptr(itr), 0, 0, 0, 0)); err != nil {
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
	descrs := make([]*t_descriptor, 0) // collect all desciption handles we will need them to fetch row

	// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629
	d := new_desc(self)
	parm_status, _, _ := oci_OCIParamGet.Call(self.ptr, OCI_HTYPE_STMT, self.conn.err.ptr, ref(&d.ptr), uintptr(len(descrs)+1))

	for parm_status == OCI_SUCCESS { // oci_success
		columns = append(columns, d.name())

		var err error
		pos := uintptr(len(descrs) + 1)

		switch d.typ() {
		case OCI_TYP_ROWID:
			buf := make([]byte, 18)
			d.val_ptr = buf
			err = self.conn.cerr(oci_OCIDefineByPos.Call(self.ptr, ref(&d.ptr), self.conn.err.ptr, pos, buf_addr(buf), 18, SQLT_AFC, int_ref(&d.ind), 0, 0, 0))
		case OCI_TYP_VARCHAR, OCI_TYP_CHAR:
			buf := make([]byte, d.len()+1)
			d.val_ptr = buf
			err = self.conn.cerr(oci_OCIDefineByPos.Call(self.ptr, ref(&d.ptr), self.conn.err.ptr, pos, buf_addr(buf), uintptr(len(buf)), SQLT_STR, int_ref(&d.ind), 0, 0, 0))
		case OCI_TYP_NUMBER:
			x := float64(0)
			d.val_ptr = &x
			err = self.conn.cerr(oci_OCIDefineByPos.Call(self.ptr, ref(&d.ptr), self.conn.err.ptr, pos, float64_ref(d.val_ptr.(*float64)), 8, SQLT_FLT, int_ref(&d.ind), 0, 0, 0))
		case OCI_TYP_DATE:
			buf := make([]byte, d.len())
			d.val_ptr = buf
			err = self.conn.cerr(oci_OCIDefineByPos.Call(self.ptr, ref(&d.ptr), self.conn.err.ptr, pos, buf_addr(buf), uintptr(len(buf)), SQLT_DAT, int_ref(&d.ind), 0, 0, 0))
		}

		if err != nil {
			return nil, err
		}

		descrs = append(descrs, d)
		d = new_desc(self)
		parm_status, _, _ = oci_OCIParamGet.Call(self.ptr, OCI_HTYPE_STMT, self.conn.err.ptr, ref(&d.ptr), uintptr(len(descrs)+1))
	}

	return &t_rows{stmt: self, columns: columns, descr: descrs}, nil
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
