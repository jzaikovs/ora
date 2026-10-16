package ora

import (
	"reflect"
	"unsafe"
)

// Descriptor describes variable where to put query result
// http://web.stanford.edu/dept/itss/docs/oracle/10gR2/appdev.102/b14250/oci04sql.htm#sthref629
type Descriptor struct {
	*ociHandle
	stmt        *Statement
	typ         int
	name        string
	length      int
	rlen        uint16 // OCI ub2 returned length
	ind         int16  // OCI sb2: -1 NULL, positive/-2 truncated
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

	if err = stmt.conn.cerr(oci_OCIParamGet.Call(
		stmt.ptr,
		OCI_HTYPE_STMT,
		stmt.conn.err.ptr,
		ref(&d.ptr),
		uintptr(pos),
	)); err != nil {
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
	err = descriptor.stmt.conn.cerr(oci_OCIAttrGet.Call(
		descriptor.ptr,
		OCI_DTYPE_PARAM,
		uintptr(unsafe.Pointer(&t)),
		0,
		uintptr(attr),
		descriptor.stmt.conn.err.ptr,
	))
	return
}

func (descriptor *Descriptor) getName() (name string, err error) {
	// OCI returns a borrowed pointer and a ub4 length, not a Go slice.
	// Convert output pointers in the call so Go keeps them stable during OCI.
	var namePtr *[1 << 30]byte
	var nameLen uint32
	if err = descriptor.stmt.conn.cerr(oci_OCIAttrGet.Call(
		descriptor.ptr,
		OCI_DTYPE_PARAM,
		uintptr(unsafe.Pointer(&namePtr)),
		uintptr(unsafe.Pointer(&nameLen)),
		OCI_ATTR_NAME,
		descriptor.stmt.conn.err.ptr,
	)); err != nil {
		return
	}
	if nameLen != 0 {
		name = string(namePtr[:nameLen:nameLen])
	}
	return
}

func (descriptor *Descriptor) define(pos int, addr interface{}, size int, typ int) error {
	// A uintptr address is used directly; any other value is retained in valPtr
	// so the buffer stays alive while OCI holds the pointer.
	ptr, ok := addr.(uintptr)
	if !ok {
		descriptor.valPtr = addr
		ptr = reflect.ValueOf(descriptor.valPtr).Pointer()
	}
	return descriptor.stmt.conn.cerr(oci_OCIDefineByPos.Call(
		descriptor.stmt.ptr,
		descriptor.ref(),
		descriptor.stmt.conn.err.ptr,
		uintptr(pos),
		ptr,
		uintptr(size),
		uintptr(typ),
		int16Ref(&descriptor.ind),
		uint16Ref(&descriptor.rlen),
		0,
		0,
	))
}
