package ora

import (
	"reflect"
)

// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

type t_descriptor struct {
	stmt    *t_stmt // statement, TODO: do we need this?
	ptr     uintptr // pointer to descriptor allocation
	typ     int
	name    string
	length  int
	ind     int // indicator, used to determin if result is null value
	val_ptr interface{}
}

func (self *t_descriptor) getTyp() (t int) {
	err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.ptr, OCI_DTYPE_PARAM, int_ref(&t), 0, OCI_ATTR_DATA_TYPE, self.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return
}

func (self *t_descriptor) getName() string {
	name := make([]byte, 32)
	name_len := 0

	err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.ptr, OCI_DTYPE_PARAM, buf_ref(&name), int_ref(&name_len), OCI_ATTR_NAME, self.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return string(name[:name_len])
}

func (self *t_descriptor) getLen() int {
	sem := 0
	// Retrieve the length semantics for the column
	if err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.ptr, OCI_DTYPE_PARAM, int_ref(&sem), 0, OCI_ATTR_CHAR_USED, self.stmt.conn.err.ptr)); err != nil {
		panic(err)
	}

	w := 0

	if sem > 0 {
		// Retrieve the column width in characters
		if err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.ptr, OCI_DTYPE_PARAM, int_ref(&w), 0, OCI_ATTR_CHAR_SIZE, self.stmt.conn.err.ptr)); err != nil {
			panic(err)
		}
	} else {
		if err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.ptr, OCI_DTYPE_PARAM, int_ref(&w), 0, OCI_ATTR_DATA_SIZE, self.stmt.conn.err.ptr)); err != nil {
			panic(err)
		}
	}
	return w
}

func (self *t_descriptor) define(pos int, addr interface{}, size int, typ int) error {
	self.val_ptr = addr
	ptr := reflect.ValueOf(self.val_ptr).Pointer()
	return self.stmt.conn.cerr(oci_OCIDefineByPos.Call(self.stmt.ptr, ref(&self.ptr), self.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), int_ref(&self.ind), 0, 0, 0))
}

/*
func (self t_descriptor) rowid() *oci_handle {
	h := self.stmt.conn.alloc_descr()
	if err := self.stmt.conn.cerr(oci_OCIAttrGet.Call(self.stmt.ptr, OCI_HTYPE_STMT, h.ref(), 0, OCI_ATTR_ROWID, self.stmt.conn.err.ptr)); err != nil {
		panic(err)
	}
	return h
}
*/
