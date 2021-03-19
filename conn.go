package ora

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// Conn represents single connection with database
type Conn struct {
	env        *ociHandle
	serv       *ociHandle
	err        *ociHandle
	tx         *Transaction
	opened     bool
	statements map[int]*Statement
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci16rel001.htm#LNOCI7016
func newConnection() (*Conn, error) {
	conn := new(Conn)
	conn.env = &ociHandle{typ: OCI_HTYPE_ENV}
	conn.statements = make(map[int]*Statement)

	// TODO: OCI_THREADED
	err := conn.envErr(oci_OCIEnvCreate.Call(conn.env.ref(), OCI_DEFAULT, 0, 0, 0, 0, 0, 0))
	if err != nil {
		return nil, err
	}

	if conn.serv, err = conn.alloc(OCI_HTYPE_SVCCTX); err != nil {
		conn.env.free()
		return nil, err
	}

	if conn.err, err = conn.alloc(OCI_HTYPE_ERROR); err != nil {
		conn.serv.free()
		conn.env.free()
		return nil, err
	}

	return conn, nil
}

// Begin begins transaction
func (conn *Conn) Begin() (driver.Tx, error) {
	err := conn.cerr(oci_OCITransStart.Call(conn.serv.ptr, conn.err.ptr, 600, 1))
	conn.tx = &Transaction{conn}
	return conn.tx, err
}

// Exec implements driver.Execer
func (conn *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	stmt, err := conn.newStatement(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			trace.Println(err)
		}
	}()

	return stmt.Exec(args)
}

// Prepare creates statement for query
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	return conn.newStatement(query)
}

// Close closes connection
func (conn *Conn) Close() error {
	// trace.Println("conn.Close")
	if !conn.opened {
		return errors.New("already closed")
	}

	conn.free()
	return nil
}

// Query implements driver.Queryer
func (conn *ConnStd) Query(query string, args []driver.Value) (driver.Rows, error) {
	// debug.PrintStack()
	stmt, err := conn.newStatement(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	return rows, err
}

// Query executes query statement using specified connection
func (conn *Conn) Query(query string, binds ...interface{}) (qr *QueryResult, err error) {
	stmt, err := conn.newStatement(query)
	if err != nil {
		return nil, err
	}

	values := make([]driver.Value, len(binds))
	for i, b := range binds {
		values[i] = b
	}

	rows, err := stmt.Query(values)
	if err != nil {
		trace.Println(err)
		return nil, err
	}

	return newQueryResult(rows.(*Rows), stmt), err
}

func (conn *Conn) free() {
	for _, stmt := range conn.statements {
		if err := stmt.close(); err != nil {
			trace.Println(err)
		}
	}

	if err := conn.cerr(oci_OCILogoff.Call(conn.serv.ptr, conn.err.ptr)); err != nil {
		trace.Println(err)
	}

	conn.statements = make(map[int]*Statement)
	conn.serv.free()
	conn.env.free()
	conn.opened = false
}

func (conn *Conn) logon(user, pass, host []byte) (err error) {
	userLen := uintptr(len(user))
	passLen := uintptr(len(pass))
	hostLen := uintptr(len(host))

	if err = conn.cerr(oci_OCILogon.Call(conn.env.ptr, conn.err.ptr, ref(&conn.serv.ptr), bufAddr(user), userLen, bufAddr(pass), passLen, bufAddr(host), hostLen)); err != nil {
		err = conn.getErr(OCI_HTYPE_ERROR)
		conn.free()
	} else {
		conn.opened = true
	}
	return
}

// function for handling errors from OCI calls
func (conn *Conn) cerr(r uintptr, r2 uintptr, err error) error {
	return conn.onOCIReturn(int16(r), OCI_HTYPE_ERROR)
}

// function for handling errors on env
func (conn *Conn) envErr(r uintptr, r2 uintptr, err error) error {
	return conn.onOCIReturn(int16(r), OCI_HTYPE_ENV)
}

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci17msc007.htm#LNOCI17287
func (conn *Conn) onOCIReturn(code int16, htyp int) error {
	switch code {
	case OCI_SUCCESS:
		return nil
	case OCI_SUCCESS_WITH_INFO:
		// trace.Println("Error: OCI_SUCCESS_WITH_INFO")
		return nil
	case OCI_NEED_DATA:
		// trace.Println("Error: OCI_NEED_DATA")
		return nil
	case OCI_ERROR:
		return conn.getErr(htyp)
	case OCI_INVALID_HANDLE:
		return errors.New("OCI call returned OCI_INVALID_HANDLE")
	case OCI_STILL_EXECUTING:
		//return fmt.Errorf("Error: OCI_STILL_EXECUTE")
		return nil
	case OCI_CONTINUE:
		//fmt.Errorf("Error: OCI_CONTINUE")
		return nil
	default:
		//fmt.Println("OCI:", conn.getErr(htyp))
	}

	return fmt.Errorf("OCI call returned - %d, %v", code, conn.getErr(htyp))
}

// https://docs.oracle.com/database/121/LNOCI/oci17msc007.htm#LNOCI17287
func (conn *Conn) getErr(htyp int) error {
	buf := make([]byte, 3072) // OCI_ERROR_MAXMSG_SIZE2 3072
	errcode := 0
	if htyp == OCI_HTYPE_ERROR {
		if err := conn.cerr(oci_OCIErrorGet.Call(conn.err.ptr, uintptr(1), uintptr(0), intRef(&errcode), bufAddr(buf), uintptr(len(buf)), OCI_HTYPE_ERROR)); err != nil {
			return err
		}
	} else {
		if err := conn.cerr(oci_OCIErrorGet.Call(conn.env.ptr, uintptr(1), uintptr(0), intRef(&errcode), bufAddr(buf), uintptr(len(buf)), OCI_HTYPE_ENV)); err != nil {
			return err
		}
	}

	buf = buf[:bytes.IndexByte(buf, 0)]

	return Error{
		Code:    errcode,
		Message: strings.Trim(string(buf), "\n"),
	}
}
