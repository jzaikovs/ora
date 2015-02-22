package ora

type oci_handle struct {
	ptr uintptr
	typ int
}

func (self oci_handle) free() {
	oci_OCIHandleFree.Call(self.ptr, uintptr(self.typ))
}

func (self *oci_handle) ref() uintptr {
	return ref(&self.ptr)
}
