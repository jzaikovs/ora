package ora

import (
	"log"
	"unsafe"
)

func logLine(args ...interface{}) {
	log.Println(args...)
}

func int_ref(p *int) uintptr {
	return uintptr(unsafe.Pointer(p))
}

func float64_ref(p *float64) uintptr {
	return uintptr(unsafe.Pointer(p))
}

func int64_ref(p *int64) uintptr {
	return uintptr(unsafe.Pointer(p))
}

func buf_addr(p []byte) uintptr {
	if len(p) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&p[0]))
}

func buf_ref(p *[]byte) uintptr {
	return uintptr(unsafe.Pointer(p))
}

func ref(p *uintptr) uintptr {
	return uintptr(unsafe.Pointer(p))
}
