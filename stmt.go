package ora

import (
	"database/sql/driver"
	"time"
)

// Statement handles single SQL statement
type Statement struct {
	*ociHandle
	conn   *connStruct
	tx     *Transaction
	binds  []interface{} // will hold pointers to every bind variable
	closed bool
}

// Close closes statement
func (stmt *Statement) Close() error {
	if stmt.closed {
		return nil
	}
	stmt.closed = true
	return stmt.conn.cerr(oci_OCIStmtRelease.Call(stmt.ptr, stmt.conn.err.ptr, 0, 0, OCI_DEFAULT))
}

// NumInput returns number of imput parameters in statement
func (stmt *Statement) NumInput() int {
	return -1 // TODO: count input parameters
}

// Exec executes statement with passed binds
func (stmt *Statement) Exec(args []driver.Value) (driver.Result, error) {
	if err := stmt.bind(args); err != nil {
		return nil, err
	}

	if err := stmt.exec(1); err != nil {
		return nil, err
	}

	// closing this is failing with invalid handle error,
	// OCIStmtExecute on non SELECT will close handle, can't find info about this case
	//if err := stmt.Close(); err != nil {
	//	return nil, err
	//}

	return Result{}, nil
}

func (stmt *Statement) bind(args []driver.Value) (err error) {
	// create link to variables, so that GC will not discard them
	stmt.binds = make([]interface{}, len(args))

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

			stmt.binds[i] = p
			err = stmt.conn.cerr(oci_OCIBindByPos.Call(stmt.ptr, ref(&bnd), stmt.conn.err.ptr, uintptr(i+1), bufAddr(p), uintptr(7), SQLT_DAT, 0, 0, 0, 0, 0, OCI_DEFAULT))
		case int64:
			x := int(val)
			stmt.binds[i] = &x
			err = stmt.conn.cerr(oci_OCIBindByPos.Call(stmt.ptr, ref(&bnd), stmt.conn.err.ptr, uintptr(i+1), intRef(&x), uintptr(sizeOfInt), SQLT_INT, 0, 0, 0, 0, 0, OCI_DEFAULT))
		case string:
			buf := append([]byte(val), 0)
			stmt.binds[i] = buf
			err = stmt.conn.cerr(oci_OCIBindByPos.Call(stmt.ptr, ref(&bnd), stmt.conn.err.ptr, uintptr(i+1), bufAddr(buf), uintptr(len(buf)), SQLT_STR, 0, 0, 0, 0, 0, OCI_DEFAULT))
		}

		if err != nil {
			stmt.Close()
			return
		}
	}
	return
}

func (stmt *Statement) exec(n int) (err error) {
	mode := OCI_DEFAULT // default will fetch n rows return describe, we do it only before first fetch when call define

	if stmt.tx == nil {
		mode = OCI_COMMIT_ON_SUCCESS
	}

	if err = stmt.conn.cerr(oci_OCIStmtExecute.Call(stmt.conn.serv.ptr, stmt.ptr, stmt.conn.err.ptr, uintptr(n), 0, 0, 0, uintptr(mode))); err != nil {
		stmt.Close()
	}
	return
}

// Query executes query statement
func (stmt *Statement) Query(args []driver.Value) (driver.Rows, error) {
	if err := stmt.bind(args); err != nil {
		return nil, err
	}

	if err := stmt.exec(0); err != nil {
		return nil, err
	}

	var (
		columns []string      // collect all columns names we will need them for database/sql
		descrs  []*descriptor // collect all description handles we will need them to fetch row
	)

	// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

	d, err := stmt.newDescriptor(1)
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
			err = d.define(pos, &tmp, sizeOfInt, SQLT_FLT)
		case OCI_TYP_DATE:
			buf := make([]byte, d.length)
			err = d.define(pos, buf, len(buf), SQLT_DAT)
		}

		if err != nil {
			logLine("define result failed with err:", err)
			return nil, err
		}

		descrs = append(descrs, d)
		d, err = stmt.newDescriptor(len(descrs) + 1)
	}

	return &Rows{stmt: stmt, columns: columns, descr: descrs}, nil
}

func (stmt *Statement) newDescriptor(pos int) (d *descriptor, err error) {
	d = &descriptor{stmt: stmt}
	if err = stmt.conn.cerr(oci_OCIParamGet.Call(stmt.ptr, OCI_HTYPE_STMT, stmt.conn.err.ptr, ref(&d.ptr), uintptr(pos))); err != nil {
		return
	}
	d.name = d.getName()
	d.typ = d.getTyp()
	d.length = d.getLen()

	return
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci17msc001.htm#i575144
func (stmt *Statement) prepare(query string) (err error) {
	buf := append([]byte(query), 0)
	if err = stmt.conn.cerr(oci_OCIStmtPrepare2.Call(
		stmt.conn.serv.ptr,
		ref(&stmt.ptr),
		stmt.conn.err.ptr,
		bufAddr(buf),
		uintptr(len(buf)),
		0,
		0,
		OCI_NTV_SYNTAX,
		0)); err != nil {
		//stmt.Close() // free alloc
	}
	return
}
