package ora

import (
	"reflect"
)

// Descriptor describes variable where to put query result
// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629
type Descriptor struct {
	*ociHandle
	stmt        *Statement
	typ         int
	name        string
	length      int
	rlen        int // offset if fetching large objects
	ind         int // indicator, used to determin if result is null value
	prec        int
	scale       int
	displaySize int
	valPtr      interface{}
}

func newDescriptor(stmt *Statement, pos int) (d *Descriptor, err error) {
	d = &Descriptor{
		ociHandle: &ociHandle{},
		stmt:      stmt,
	}

	if err = stmt.conn.cerr(oci_OCIParamGet.Call(stmt.ptr, OCI_HTYPE_STMT, stmt.conn.err.ptr, ref(&d.ptr), uintptr(pos))); err != nil {
		//trace.Printf("OCIParamGet(..., %d) -> %s", pos, err)
		return
	}

	if d.name, err = d.getName(); err != nil {
		return
	}

	if d.typ, err = d.getIntAttr(OCI_ATTR_DATA_TYPE); err != nil {
		return
	}

	if d.length, err = d.getIntAttr(OCI_ATTR_DATA_SIZE); err != nil {
		return
	}

	if d.displaySize, err = d.getIntAttr(OCI_ATTR_DISP_SIZE); err != nil {
		return
	}

	if d.prec, err = d.getIntAttr(OCI_ATTR_PRECISION); err != nil {
		return
	}

	if d.scale, err = d.getIntAttr(OCI_ATTR_SCALE); err != nil {
		return
	}

	return
}

func (descriptor *Descriptor) getIntAttr(attr int) (t int, err error) {
	err = descriptor.stmt.conn.cerr(oci_OCIAttrGet.Call(descriptor.ptr, OCI_DTYPE_PARAM, intRef(&t), 0, uintptr(attr), descriptor.stmt.conn.err.ptr))
	return
}

func (descriptor *Descriptor) getName() (name string, err error) {
	buf := make([]byte, 512) // TODO: what is max length of result column name?
	bufLen := 0

	if err = descriptor.stmt.conn.cerr(oci_OCIAttrGet.Call(descriptor.ptr, OCI_DTYPE_PARAM, bufRef(&buf), intRef(&bufLen), OCI_ATTR_NAME, descriptor.stmt.conn.err.ptr)); err != nil {
		return
	}
	name = string(buf[:bufLen])
	return
}

func (descriptor *Descriptor) define(pos int, addr interface{}, size int, typ int) error {
	if ptr, ok := addr.(uintptr); ok {
		return descriptor.stmt.conn.cerr(oci_OCIDefineByPos.Call(descriptor.stmt.ptr, descriptor.ref(), descriptor.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), intRef(&descriptor.ind), intRef(&descriptor.rlen), 0, 0))
	}

	descriptor.valPtr = addr
	ptr := reflect.ValueOf(descriptor.valPtr).Pointer()
	return descriptor.stmt.conn.cerr(oci_OCIDefineByPos.Call(descriptor.stmt.ptr, descriptor.ref(), descriptor.stmt.conn.err.ptr, uintptr(pos), ptr, uintptr(size), uintptr(typ), intRef(&descriptor.ind), intRef(&descriptor.rlen), 0, 0))
}
