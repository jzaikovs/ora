package ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// MaxLongSize is size of buffer allocated for long type, (TODO: can this be improved to dynamic allocation?)
var MaxLongSize = 100000

// PrefetchRows is binds count to prefetch
var PrefetchRows = 1000

// Statement handles single SQL statement
type Statement struct {
	*ociHandle
	id     int
	conn   *Conn
	tx     *Transaction
	binds  []interface{} // will hold pointers to every bind variable
	closed bool
	rows   *Rows
}

// newStatement allocates new statement
func (conn *Conn) newStatement(query string) (stmt *Statement, err error) {
	var h *ociHandle
	// allocate prepare statement, later we will need to free it
	if h, err = conn.alloc(OCI_HTYPE_STMT); err != nil {
		return nil, err
	}

	stmt = &Statement{
		conn:      conn,
		tx:        conn.tx,
		ociHandle: h,
		id:        len(conn.statements) + 1,
	}

	conn.statements[stmt.id] = stmt

	if err = stmt.prepare(query); err != nil {
		if err2 := stmt.Close(); err2 != nil {
			return nil, err2
		}
	}

	return
}

// Close closes statement
func (stmt *Statement) Close() error {
	// trace.Println("stmt.Close")
	if stmt.closed {
		return errors.New("already closed")
	}

	delete(stmt.conn.statements, stmt.id)
	return stmt.close()
}

func (stmt *Statement) close() error {
	// trace.Println("stmt.close")
	if stmt.closed {
		return errors.New("already closed")
	}

	stmt.closed = true
	err := stmt.conn.cerr(oci_OCIStmtRelease.Call(stmt.ptr, stmt.conn.err.ptr, 0, 0, OCI_DEFAULT))
	stmt.free()
	return err
}

// NumInput returns number of input parameters in statement
func (stmt *Statement) NumInput() int {
	return -1
}

// Exec executes statement with passed binds
func (stmt *Statement) Exec(args []driver.Value) (driver.Result, error) {
	// fmt.Println("stmt.Exec")
	if err := stmt.bind(args); err != nil {
		return nil, err
	}

	// for _, b := range stmt.binds {
	// 	switch v := b.(type) {
	// 	case *Lob:
	// 		defer v.free()
	// 		defer v.freeTemp()
	// 	}
	// }

	if err := stmt.exec(1); err != nil {
		return nil, err
	}

	for i, vp := range args {
		switch v := vp.(type) {
		case *int:
			x := stmt.binds[i].(*int)
			*v = *x
		case *int64:
			x := stmt.binds[i].(*int64)
			*v = *x
		}
	}

	return stmt.result(), nil
}

func (stmt *Statement) result() (r Result) {
	r.err = stmt.conn.cerr(oci_OCIAttrGet.Call(stmt.ptr, OCI_HTYPE_STMT, int64Ref(&r.rowsAffected), 0, OCI_ATTR_ROW_COUNT, stmt.conn.err.ptr))
	return
}

func (stmt *Statement) bind(args []driver.Value) (err error) {
	// create link to variables, so that GC will not discard them
	stmt.binds = make([]interface{}, len(args))

	for i, arg := range args {
		var bnd uintptr
		// store pointer to val in binds because garbage collector will discard val
		// and OCI will pass some random data from memory
		switch val := arg.(type) { // GC will discard val if not referenced somewhere
		case time.Time:
			buf := timeToOraBytes(val)
			stmt.binds[i] = buf
			err = stmt.bindPyPos(i+1, &bnd, bufAddr(buf), len(buf), SQLT_DAT)
		case int:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, intRef(&x), sizeOfInt, SQLT_INT)
		case *int:
			stmt.binds[i] = val
			err = stmt.bindPyPos(i+1, &bnd, intRef(val), sizeOfInt, SQLT_INT)
		case uint:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, uintRef(&x), sizeOfInt, SQLT_INT)
		case int64:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, int64Ref(&x), 8, SQLT_INT)
		case uint64:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, uint64Ref(&x), 8, SQLT_INT)
		case float32:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, float32Ref(&x), 4, SQLT_FLT)
		case float64:
			x := val
			stmt.binds[i] = &x
			err = stmt.bindPyPos(i+1, &bnd, float64Ref(&x), 8, SQLT_FLT)
		case string:
			buf := append([]byte(val), 0) // null terminated string
			stmt.binds[i] = buf
			err = stmt.bindPyPos(i+1, &bnd, bufAddr(buf), len(buf), SQLT_STR)
		case []byte:
			buf := val
			stmt.binds[i] = buf
			err = stmt.bindPyPos(i+1, &bnd, bufAddr(buf), len(buf), SQLT_BIN)
		// case []byte:
		// 	var lob *Lob
		// 	if lob, err = stmt.conn.newLob(); err != nil {
		// 		return err
		// 	}

		// 	if err = lob.createTemp(); err != nil {
		// 		lob.free()
		// 		return err
		// 	}

		// 	if _, err = lob.Write(val); err != nil {
		// 		trace.Println(err)
		// 		lob.freeTemp()
		// 		lob.free()
		// 		return err
		// 	}

		// 	stmt.binds[i] = lob
		// 	err = stmt.bindPyPos(i+1, &bnd, ref(&lob.ptr), sizeOfInt, SQLT_BLOB)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 	}
		case nil:
			buf := []byte{0} // null terminated string
			stmt.binds[i] = buf
			err = stmt.bindPyPos(i+1, &bnd, bufAddr(buf), len(buf), SQLT_STR)
		default:
			return fmt.Errorf("unsupported bind type %T", val)
		}

		if err != nil {
			if err2 := stmt.Close(); err2 != nil {
				trace.Println(err)
			}
			return
		}
	}

	return
}

func (stmt *Statement) bindPyPos(i int, bnd *uintptr, addr uintptr, size int, typ int) error {
	return stmt.conn.cerr(oci_OCIBindByPos.Call(stmt.ptr, ref(bnd), stmt.conn.err.ptr, uintptr(i), addr, uintptr(size), uintptr(typ), 0, 0, 0, 0, 0, OCI_DEFAULT))
}

func (stmt *Statement) exec(n int) (err error) {
	mode := OCI_DEFAULT // default will fetch n rows return describe, we do it only before first fetch when call define

	if stmt.tx == nil {
		mode = OCI_COMMIT_ON_SUCCESS
	}

	if err = stmt.conn.cerr(oci_OCIStmtExecute.Call(stmt.conn.serv.ptr, stmt.ptr, stmt.conn.err.ptr, uintptr(n), 0, 0, 0, uintptr(mode))); err != nil {
		if err2 := stmt.Close(); err2 != nil {
			trace.Println(err)
		}
	}

	return
}

// Query executes query statement
func (stmt *Statement) Query(args []driver.Value) (driver.Rows, error) {
	// fmt.Println("stmt.Query")
	if err := stmt.SetPrefetch(PrefetchRows); err != nil {
		return nil, err
	}

	if err := stmt.bind(args); err != nil {
		return nil, err
	}

	if err := stmt.exec(0); err != nil {
		return nil, err
	}
	var err error
	stmt.rows, err = newRows(stmt)
	return stmt.rows, err
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci17msc001.htm#i575144
func (stmt *Statement) prepare(query string) (err error) {
	buf := append([]byte(query), 0)

	return stmt.conn.cerr(oci_OCIStmtPrepare2.Call(
		stmt.conn.serv.ptr,
		ref(&stmt.ptr),
		stmt.conn.err.ptr,
		bufAddr(buf),
		uintptr(len(buf)),
		0,
		0,
		OCI_NTV_SYNTAX,
		0,
	))
}

// SetPrefetch sets actual prefetch
func (stmt *Statement) SetPrefetch(n int) (err error) {
	val := int32(n)
	return stmt.conn.cerr(oci_OCIAttrSet.Call(
		stmt.ptr,
		uintptr(OCI_HTYPE_STMT),
		int32Ref(&val),
		uintptr(sizeOfInt),
		uintptr(OCI_ATTR_PREFETCH_ROWS),
		stmt.conn.err.ptr,
	))
}
