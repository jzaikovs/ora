package ora

import (
	"reflect"
)

// Descriptor describes variable where to put query result
// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629
type Descriptor struct {
	*ociHandle
	stmt   *Statement // statement, TODO: do we need this?
	typ    int
	name   string
	length int
	rlen   int // offset if fetching large objects
	ind    int // indicator, used to determin if result is null value
	valPtr interface{}
}

func newDescriptor(stmt *Statement) *Descriptor {
	return &Descriptor{ociHandle: &ociHandle{}, stmt: stmt}
}

func (descr *Descriptor) Type() int {
	return descr.typ

}

func (descr *Descriptor) Name() string {
	return descr.name
}

func (descr *Descriptor) getTyp() (t int) {
	err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, intRef(&t), 0, OCI_ATTR_DATA_TYPE, descr.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return
}

func (descr *Descriptor) getName() string {
	name := make([]byte, 512) // TODO: what is max length of result column name?
	nameLen := 0

	err := descr.stmt.conn.cerr(oci_OCIAttrGet.Call(descr.ptr, OCI_DTYPE_PARAM, bufRef(&name), intRef(&nameLen), OCI_ATTR_NAME, descr.stmt.conn.err.ptr))
	if err != nil {
		panic(err)
	}
	return string(name[:nameLen])
}

func (descr *Descriptor) getLen() int {
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

func (descr *Descriptor) define(pos int, addr interface{}, size int, typ int) error {
	if ptr, ok := addr.(uintptr); ok {
		return descr.stmt.conn.cerr(oci_OCIDefineByPos.Call(descr.stmt.ptr, descr.ref(), descr.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), intRef(&descr.ind), intRef(&descr.rlen), 0, 0))
	}

	descr.valPtr = addr
	ptr := reflect.ValueOf(descr.valPtr).Pointer()
	return descr.stmt.conn.cerr(oci_OCIDefineByPos.Call(descr.stmt.ptr, descr.ref(), descr.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), intRef(&descr.ind), intRef(&descr.rlen), 0, 0))
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
