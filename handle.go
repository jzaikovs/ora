package ora

type ociHandle struct {
	ptr uintptr
	typ int
}

func (handle ociHandle) free() {
	oci_OCIHandleFree.Call(handle.ptr, uintptr(handle.typ))
}

func (handle *ociHandle) ref() uintptr {
	return ref(&handle.ptr)
}
