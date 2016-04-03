package ora

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
)

type Conn struct {
	env    *ociHandle
	serv   *ociHandle
	err    *ociHandle
	tx     *Transaction
	opened bool
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci16rel001.htm#LNOCI7016
func newConnection() (*Conn, error) {
	conn := new(Conn)
	conn.env = &ociHandle{typ: OCI_HTYPE_ENV}

	// TODO: OCI_THREADED
	err := conn.envErr(oci_OCIEnvCreate.Call(conn.env.ref(), OCI_DEFAULT, 0, 0, 0, 0, 0, 0))
	if err != nil {
		return nil, err
	}

	if conn.serv, err = conn.alloc(OCI_HTYPE_SVCCTX); err != nil {
		return nil, err
	}

	if conn.err, err = conn.alloc(OCI_HTYPE_ERROR); err != nil {
		return nil, err
	}

	return conn, nil
}

// Begin begins transaction
func (conn *Conn) Begin() (driver.Tx, error) {
	conn.tx = &Transaction{conn}
	return conn.tx, nil
}

// Prepare creates statement for query
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.newStatement()
	if err != nil {
		return nil, err
	}
	if err = stmt.prepare(query); err != nil {
		return nil, err
	}
	return stmt, nil
}

// Close closes connection TODO: test that connection is actually closed!
func (conn *Conn) Close() error {
	if conn.opened {
		oci_OCILogoff.Call(conn.serv.ptr, conn.err.ptr)
		conn.opened = false
	}
	oci_OCIHandleFree.Call(conn.serv.ptr, uintptr(OCI_HTYPE_SVCCTX))
	oci_OCIHandleFree.Call(conn.env.ptr, uintptr(OCI_HTYPE_ENV))
	return nil
}

// function for creating statement
func (conn *Conn) newStatement() (stmt *Statement, err error) {
	stmt = &Statement{conn: conn, tx: conn.tx}
	stmt.ociHandle, err = conn.alloc(OCI_HTYPE_STMT) // allocate prepare statement, later we will need to free it
	return
}

func (conn *Conn) logon(user, pass, host []byte) (err error) {
	userLen := uintptr(len(user))
	passLen := uintptr(len(pass))
	hostLen := uintptr(len(host))

	if err = conn.cerr(oci_OCILogon.Call(conn.env.ptr, conn.err.ptr, ref(&conn.serv.ptr), bufAddr(user), userLen, bufAddr(pass), passLen, bufAddr(host), hostLen)); err != nil {
		err = conn.getErr(OCI_HTYPE_ERROR)
		conn.Close()
	} else {
		conn.opened = true
	}
	return
}

func (conn *Conn) alloc(typ int) (*ociHandle, error) {
	h := &ociHandle{typ: typ}
	if err := conn.envErr(oci_OCIHandleAlloc.Call(conn.env.ptr, h.ref(), uintptr(typ), 0, 0)); err != nil {
		return nil, err
	}
	return h, nil
}

// function for handling errors from OCI calls
func (conn *Conn) cerr(r uintptr, r2 uintptr, err error) error {
	return conn.onOCIReturn(int16(r), OCI_HTYPE_ERROR)
}

// function for handling errors on env create and alloc
func (conn *Conn) envErr(r uintptr, r2 uintptr, err error) error {
	return conn.onOCIReturn(int16(r), OCI_HTYPE_ENV)
}

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci17msc007.htm#LNOCI17287
func (conn *Conn) onOCIReturn(code int16, htyp int) error {
	switch code {
	case OCI_SUCCESS:
		return nil
	case OCI_SUCCESS_WITH_INFO:
		//trace.Println("Error: OCI_SUCCESS_WITH_INFO")
		return nil
	case OCI_NEED_DATA:
		//trace.Println("Error: OCI_NEED_DATA")
		return nil
	case OCI_ERROR:
		return conn.getErr(htyp)
	case OCI_INVALID_HANDLE:
		return errors.New("Error: OCI call returned OCI_INVALID_HANDLE")
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

	return errors.New(string(buf[:bytes.IndexByte(buf, 0)]))
}

// Query executes query statement using specified connenction
func (conn *Conn) Query(stmt string, binds ...interface{}) (qr *QueryResult, err error) {
	prep, err := conn.Prepare(stmt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//defer prep.Close()

	x := make([]driver.Value, len(binds))
	for i, b := range binds {
		x[i] = b
	}

	result, err := prep.Query(x)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	qr = newQueryResult(result.(*Rows), prep)

	return
}
