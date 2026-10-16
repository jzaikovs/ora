package ora

import (
	"fmt"
	"io"
	"io/ioutil"
	"unicode/utf8"
	"unsafe"
)

// Lob represents oracle lob ociHandle
type Lob struct {
	conn      *Conn
	ptr       uintptr
	offset    uint64
	isNull    bool
	character bool
	pending   []byte
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
	lob.pending = nil
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
	if len(buf) == 0 {
		return 0, nil
	}
	if len(lob.pending) > 0 {
		n = copy(buf, lob.pending)
		lob.pending = lob.pending[n:]
		return n, nil
	}
	if lob.offset < 1 {
		if !lob.isNull {
			return 0, io.EOF

		}

		if err = lob.open(); err != nil {
			return
		}
	}

	// Reserve room for two character units: Oracle represents supplementary
	// characters as surrogate pairs internally. Tiny buffers can fail in OCI.
	// Preserve excess bytes when the caller requests a smaller chunk.
	target := buf
	if lob.character && len(target) < 2*utf8.UTFMax {
		target = make([]byte, 2*utf8.UTFMax)
	}
	byteAmount := uint64(len(target))
	var charAmount uint64
	if lob.character {
		// Request enough characters to fill the buffer; OCI_ONE_PIECE stops
		// at its byte capacity without splitting a multibyte character.
		charAmount = byteAmount
	}
	r1, r2, err2 := oci_OCILobRead2.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
		uintptr(unsafe.Pointer(&byteAmount)),
		uintptr(unsafe.Pointer(&charAmount)),
		uintptr(lob.offset),
		uintptr(unsafe.Pointer(&target[0])),
		uintptr(len(target)),
		OCI_ONE_PIECE,
		null, null,
		OCI_AL32UTF8,
		SQLCS_IMPLICIT)

	if r1 == OCI_NO_DATA {
		return 0, io.EOF
	}
	if err := lob.conn.cerr(r1, r2, err2); err != nil {
		return 0, err
	}

	if byteAmount == 0 {
		return 0, io.EOF
	}

	// CLOB offsets count characters; BLOB offsets count bytes.
	if lob.character {
		lob.offset += charAmount
	} else {
		lob.offset += byteAmount
	}
	n = copy(buf, target[:int(byteAmount)])
	if n < int(byteAmount) {
		lob.pending = target[n:int(byteAmount)]
	}
	return n, nil
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
	case nil:
		// NULL LOB column: an empty reader whose Read returns io.EOF.
		lob.reset(false)
	case *string:
		*v, err = lob.String()
	case *Lob:
		*lob = *v
	default:
		return fmt.Errorf("unsupported type %T", v)
	}

	return
}
