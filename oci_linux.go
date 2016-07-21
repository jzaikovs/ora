package ora

import "github.com/rainycape/dl"

var (
	ociLibrary = NewLazyDLL("libclntsh.so")
)

// Library handler
type Library struct {
	dll *dl.DL
}

// NewLazyDLL loads static library
func NewLazyDLL(name string) (dll *Library) {
	dll = new(Library)
	var err error
	if dll.dll, err = dl.Open(name, dl.RTLD_LAZY); err != nil {
		panic(err)
	}
	return
}

// NewProc creates system call proc to passed functio from Library
func (lib *Library) NewProc(name string) *LibraryProc {
	callee := new(LibraryProc)
	callee.err = lib.dll.Sym(name, &callee.fn)
	return callee
}

// LibraryProc handler for single procedure from library
type LibraryProc struct {
	err error
	fn  func(...uintptr) uintptr
}

// Call calls function from library
func (lib LibraryProc) Call(args ...uintptr) (r1 uintptr, r2 uintptr, err error) {
	err = lib.err
	r1 = lib.fn(args...)
	return
}
