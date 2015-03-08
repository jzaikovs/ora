package ora

import (
	"syscall"
)

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci17msc001.htm#LNOCI161
var (
	oci                    = syscall.NewLazyDLL("oci.dll")
	oci_OCIAttrGet         = oci.NewProc("OCIAttrGet")
	oci_OCIAttrSet         = oci.NewProc("OCIAttrSet")
	oci_OCIBindByPos       = oci.NewProc("OCIBindByPos")
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
	oci_OCITransCommit     = oci.NewProc("OCITransCommit")
	oci_OCITransRollback   = oci.NewProc("OCITransRollback")

	//oci_OCIStmtFetch      = oci.NewProc("OCIStmtFetch") // old
	//oci_OCIStmtPrepare     = oci.NewProc("OCIStmtPrepare") // old
)
