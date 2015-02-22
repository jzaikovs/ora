package ora

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

type t_conn struct {
	env    *oci_handle
	serv   *oci_handle
	err    *oci_handle
	opened bool
}

// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci16rel001.htm#LNOCI7016
func new_conn() (*t_conn, error) {
	conn := new(t_conn)
	conn.env = &oci_handle{typ: OCI_HTYPE_ENV}

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
	return nil, nil // TODO: implement
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
	stmt = &t_stmt{conn: self}
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

// function for hanling errors from oci calls
func (self *t_conn) cerr(r uintptr, r2 uintptr, err error) error {
	if r > 0 {
		return self.handle_error()
	}
	return nil
}

// functin for handling erros on env create and alloc
func (self *t_conn) env_err(r uintptr, r2 uintptr, err error) error {
	if r > 0 {
		return self.handle_error()
	}
	return nil
}

func (self *t_conn) handle_error() error {
	buf := make([]byte, 4000)
	errcode := 0
	if oci_OCIErrorGet.Call(self.err.ptr, uintptr(1), uintptr(0), int_ref(&errcode), buf_addr(buf), uintptr(len(buf)), OCI_HTYPE_ERROR); errcode > 0 {
		return errors.New(string(buf[:bytes.IndexByte(buf, 0)]))
	}
	return nil
}

func (self *t_conn) handle_env_error() error {
	buf := make([]byte, 4000)
	errcode := 0
	if oci_OCIErrorGet.Call(self.env.ptr, uintptr(1), uintptr(0), int_ref(&errcode), buf_addr(buf), uintptr(len(buf)), OCI_HTYPE_ENV); errcode > 0 {
		return errors.New(string(buf[:bytes.IndexByte(buf, 0)]))
	}
	return nil
}
