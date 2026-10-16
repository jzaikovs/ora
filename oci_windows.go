package ora

import (
	"syscall"
)

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci17msc001.htm#LNOCI161
var (
	ociLibrary = syscall.NewLazyDLL("oci.dll")
)
