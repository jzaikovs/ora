package ora

import (
	"reflect"
)

// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629

type descriptor struct {
	stmt   *Statement // statement, TODO: do we need this?
	ptr    uintptr    // pointer to descriptor allocation
	typ    int
	name   string
	length int
	ind    int // indicator, used to determin if result is null value
	valPtr interface{}
}

func (descr *descriptor) getTyp() (t int) {
	err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, intRef(&t), 0, OCI_ATTR_DATA_TYPE, descr.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return
}

func (descr *descriptor) getName() string {
	name := make([]byte, 32)
	nameLen := 0

	err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, bufRef(&name), intRef(&nameLen), OCI_ATTR_NAME, descr.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return string(name[:nameLen])
}

func (descr *descriptor) getLen() int {
	sem := 0
	// Retrieve the length semantics for the column
	if err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, intRef(&sem), 0, OCI_ATTR_CHAR_USED, descr.stmt.conn.err.ptr)); err != nil {
		panic(err)
	}

	w := 0

	if sem > 0 {
		// Retrieve the column width in characters
		if err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, intRef(&w), 0, OCI_ATTR_CHAR_SIZE, descr.stmt.conn.err.ptr)); err != nil {
			panic(err)
		}
	} else {
		if err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, intRef(&w), 0, OCI_ATTR_DATA_SIZE, descr.stmt.conn.err.ptr)); err != nil {
			panic(err)
		}
	}
	return w
}

func (descr *descriptor) define(pos int, addr interface{}, size int, typ int) error {
	descr.valPtr = addr
	ptr := reflect.ValueOf(descr.valPtr).Pointer()
	return descr.stmt.conn.cerr(oci_OCIDefineByPos.Call(descr.stmt.ptr, ref(&descr.ptr), descr.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), intRef(&descr.ind), 0, 0, 0))
}

/*
func (descr descriptor) rowid() *oci_handle {
	h := descr.stmt.conn.alloc_descr()
	if err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.stmt.ptr, OCI_HTYPE_STMT, h.ref(), 0, OCI_ATTR_ROWID, descr.stmt.conn.err.ptr)); err != nil {
		panic(err)
	}
	return h
}
*/
