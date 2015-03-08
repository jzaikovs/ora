package ora

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
)

type t_conn struct {
	env    *oci_handle
	serv   *oci_handle
	err    *oci_handle
	tx     *t_tx
	opened bool
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci16rel001.htm#LNOCI7016
func new_conn() (*t_conn, error) {
	conn := new(t_conn)
	conn.env = &oci_handle{typ: OCI_HTYPE_ENV}

	// TODO: OCI_THREADED
	err := conn.env_err(oci_OCIEnvCreate.Call(conn.env.ref(), OCI_DEFAULT, 0, 0, 0, 0, 0, 0))
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

func (self *t_conn) Begin() (driver.Tx, error) {
	self.tx = &t_tx{self}
	return self.tx, nil
}

func (self *t_conn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := self.new_stmt()
	if err != nil {
		return nil, err
	}
	if err = stmt.prepare(query); err != nil {
		return nil, err
	}
	return stmt, nil
}

// TODO: test that connection is actually closed!
func (self *t_conn) Close() error {
	if self.opened {
		oci_OCILogoff.Call(self.serv.ptr, self.err.ptr)
		self.opened = false
	}
	oci_OCIHandleFree.Call(self.serv.ptr, uintptr(OCI_HTYPE_SVCCTX))
	oci_OCIHandleFree.Call(self.env.ptr, uintptr(OCI_HTYPE_ENV))
	return nil
}

// function for creating statement
func (self *t_conn) new_stmt() (stmt *t_stmt, err error) {
	stmt = &t_stmt{conn: self, tx: self.tx}
	stmt.oci_handle, err = self.alloc(OCI_HTYPE_STMT) // allocate prepare statement, later we will need to free it
	return
}

func (self *t_conn) logon(user, pass, host []byte) (err error) {
	user_n := uintptr(len(user))
	pass_n := uintptr(len(pass))
	host_n := uintptr(len(host))

	if err = self.cerr(oci_OCILogon.Call(self.env.ptr, self.err.ptr, ref(&self.serv.ptr), buf_addr(user), user_n, buf_addr(pass), pass_n, buf_addr(host), host_n)); err != nil {
		self.Close()
	} else {
		self.opened = true
	}
	return
}

func (self *t_conn) alloc(typ int) (*oci_handle, error) {
	h := &oci_handle{typ: typ}
	if err := self.env_err(oci_OCIHandleAlloc.Call(self.env.ptr, h.ref(), uintptr(typ), 0, 0)); err != nil {
		return nil, err
	}
	return h, nil
}

/* for later use
func (self *t_conn) alloc_descr() *oci_handle {
	h := new(oci_handle)
	err := self.env_err(oci_OCIDescriptorAlloc.Call(self.env.ptr, h.ref(), OCI_DTYPE_ROWID, 0, 0))
	if err != nil {
		panic(err)
	}
	return h
}
*/

// function for handling errors from OCI calls
func (self *t_conn) cerr(r uintptr, r2 uintptr, err error) error {
	return self.on_oci_return(int16(r), OCI_HTYPE_ERROR)
}

// function for handling errors on env create and alloc
func (self *t_conn) env_err(r uintptr, r2 uintptr, err error) error {
	return self.on_oci_return(int16(r), OCI_HTYPE_ENV)
}

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci17msc007.htm#LNOCI17287
func (self *t_conn) on_oci_return(code int16, htyp int) error {
	switch code {
	case OCI_SUCCESS:
		return nil
	case OCI_ERROR:
		return self.error_get(htyp)
	case OCI_INVALID_HANDLE:
		return errors.New("OCI call returned OCI_INVALID_HANDLE")
	}

	return errors.New(fmt.Sprintf("OCI call returned - %d", code))
}

// https://docs.oracle.com/database/121/LNOCI/oci17msc007.htm#LNOCI17287
func (self *t_conn) error_get(htyp int) error {
	buf := make([]byte, 3072) // OCI_ERROR_MAXMSG_SIZE2 3072
	errcode := 0
	if htyp == OCI_HTYPE_ERROR {
		if err := self.cerr(oci_OCIErrorGet.Call(self.err.ptr, uintptr(1), uintptr(0), int_ref(&errcode), buf_addr(buf), uintptr(len(buf)), OCI_HTYPE_ERROR)); err != nil {
			return err
		}
	} else {
		if err := self.cerr(oci_OCIErrorGet.Call(self.env.ptr, uintptr(1), uintptr(0), int_ref(&errcode), buf_addr(buf), uintptr(len(buf)), OCI_HTYPE_ENV)); err != nil {
			return err
		}
	}

	return errors.New(string(buf[:bytes.IndexByte(buf, 0)]))
}
