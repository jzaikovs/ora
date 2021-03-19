package ora

import (
	"fmt"
	"io"
	"io/ioutil"
)

// Lob represents oracle lob ociHandle
type Lob struct {
	conn     *Conn
	ptr      uintptr
	offset   int
	needData bool
	isNull   bool
}

func (conn *Conn) NewLob() (lob *Lob, err error) {
	lob = &Lob{conn: conn}
	err = conn.cerr(
		oci_OCIDescriptorAlloc.Call(
			conn.env.ptr,
			ref(&lob.ptr),
			OCI_DTYPE_LOB,
			null, null,
		))
	return
}

func (lob *Lob) CreateTemp() (err error) {
	err = lob.conn.cerr(oci_OCILobCreateTemporary.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
	))
	return
}

func (lob *Lob) FreeTemp() (err error) {
	err = lob.conn.cerr(oci_OCILobFreeTemporary.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
	))
	return
}

// open opens oracle lob for reading
func (lob *Lob) open() (err error) {
	if err = lob.conn.cerr(oci_OCILobOpen.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
		OCI_LOB_READONLY,
	)); err != nil {
		lob.free()
		return
	}

	lob.offset = 1
	return
}

func (lob *Lob) reset(isNull bool) {
	lob.offset = 0
	lob.isNull = isNull
}

func (lob *Lob) free() {
	if err := lob.conn.cerr(oci_OCIDescriptorFree.Call(
		lob.ptr,
		OCI_DTYPE_LOB,
	)); err != nil {
		trace.Println(err)
	}
}

func (lob *Lob) Write(buf []byte) (n int, err error) {
	if err = lob.open(); err != nil {
		fmt.Println("bad open")
		return
	}

	amount := int32(len(buf))

	err = lob.conn.cerr(oci_OCILobWriteAppend.Call(
		lob.conn.serv.ptr, // service
		lob.conn.err.ptr,  //error
		lob.ptr,           // ptr
		int32Ref(&amount),
		bufAddr(buf),
		uintptr(int32(len(buf))), // buffer length
		0,
		null, // OCICallbackLobWrite
		null,
		null,
		null,
		null,
		null,
		SQLCS_IMPLICIT,
	))

	trace.Println(err)
	return
}

// Read reads from lob
func (lob *Lob) Read(buf []byte) (n int, err error) {
	if lob.offset < 1 {
		if !lob.isNull {
			return 0, io.EOF

		}

		if err = lob.open(); err != nil {
			return
		}
	}

	amount := int32(len(buf))
	r1, r2, err2 := oci_OCILobRead.Call(
		lob.conn.serv.ptr, // service
		lob.conn.err.ptr,  //error
		lob.ptr,           // ptr
		int32Ref(&amount),
		uintptr(int32(lob.offset)), // offset
		bufAddr(buf),               // buffer
		uintptr(int32(len(buf))),   // buffer length
		null,
		null,
		null,
		null,
		null,
		null,
		null,
		SQLCS_IMPLICIT)

	if err := lob.conn.cerr(r1, r2, err2); err != nil {
		return 0, err
	}

	if n = int(amount); n <= 0 {
		return 0, io.EOF
	}

	if !lob.needData {
		lob.offset += n
	}

	lob.needData = r1 == OCI_NEED_DATA
	return
}

// Close closes lob reader
func (lob *Lob) Close() (err error) {
	err = lob.conn.cerr(oci_OCILobClose.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
	))

	lob.free()
	return
}

func (lob *Lob) String() (string, error) {
	b, err := ioutil.ReadAll(lob)
	return string(b), err
}

func (lob *Lob) Scan(src interface{}) (err error) {
	switch v := src.(type) {
	case *string:
		*v, err = lob.String()
	case *Lob:
		*lob = *v
	default:
		return fmt.Errorf("unsupported type %T", v)
	}

	return
}
