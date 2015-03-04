package ora

import (
	"gopkgs.com/dl.v1"
)

type t_dll struct {
	dll *dl.DL
}

func NewLazyDLL(name string) (dll *t_dll) {
	dll = new(t_dll)
	var err error
	if dll.dll, err = dl.Open("libclntsh.so", dl.RTLD_LAZY); err != nil {
		panic(err)
	}
	return
}

func (self *t_dll) NewProc(name string) *t_callee {
	callee := new(t_callee)
	callee.err = self.dll.Sym(name, &callee.fn)
	return callee
}

type t_callee struct {
	err error
	fn  func(...uintptr) uintptr
}

func (self t_callee) Call(args ...uintptr) (r1 uintptr, r2 uintptr, err error) {
	err = self.err
	r1 = self.fn(args...)
	return
}

var (
	oci                    = NewLazyDLL("oci.dll")
	oci_OCIAttrGet         = oci.NewProc("OCIAttrGet")
	oci_OCIAttrSet         = oci.NewProc("OCIAttrSet")
	oci_OCIBindByName      = oci.NewProc("OCIBindByName")
	oci_OCIDefineByPos     = oci.NewProc("OCIDefineByPos")
	oci_OCIDescriptorAlloc = oci.NewProc("OCIDescriptorAlloc")
	oci_OCIDescriptorFree  = oci.NewProc("OCIDescriptorFree")
	oci_OCIEnvCreate       = oci.NewProc("OCIEnvCreate")
	oci_OCIErrorGet        = oci.NewProc("OCIErrorGet")
	oci_OCIHandleAlloc     = oci.NewProc("OCIHandleAlloc")
	oci_OCIHandleFree      = oci.NewProc("OCIHandleFree")
	oci_OCIInitialize      = oci.NewProc("OCIInitialize")
	oci_OCILogoff          = oci.NewProc("OCILogoff")
	oci_OCILogon           = oci.NewProc("OCILogon")
	oci_OCIParamGet        = oci.NewProc("OCIParamGet")
	oci_OCIRowidToChar     = oci.NewProc("OCIRowidToChar")
	oci_OCIStmtExecute     = oci.NewProc("OCIStmtExecute")
	oci_OCIStmtFetch2      = oci.NewProc("OCIStmtFetch2")
	oci_OCIStmtPrepare2    = oci.NewProc("OCIStmtPrepare2") // this allows statement caching
	oci_OCIStmtRelease     = oci.NewProc("OCIStmtRelease")
)
