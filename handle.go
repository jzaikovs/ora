package ora

import (
	"sync/atomic"
)

var handleRefCount int64 = 0

type ociHandle struct {
	ptr uintptr
	typ int
}

func (conn *Conn) alloc(typ int) (*ociHandle, error) {
	h := &ociHandle{typ: typ}
	if err := conn.envErr(oci_OCIHandleAlloc.Call(conn.env.ptr, h.ref(), uintptr(typ), 0, 0)); err != nil {
		return nil, err
	}
	atomic.AddInt64(&handleRefCount, 1)
	return h, nil
}

func (handle *ociHandle) free() {
	// OCIHandleFree reports failures through its status, not the OS last-error value.
	status, _, _ := oci_OCIHandleFree.Call(handle.ptr, uintptr(handle.typ))
	if code := int16(status); code != OCI_SUCCESS {
		trace.Printf("OCIHandleFree failed for type %d: OCI status %d", handle.typ, code)
		return
	}
	atomic.AddInt64(&handleRefCount, -1)
}

func (handle *ociHandle) ref() uintptr {
	return ref(&handle.ptr)
}
