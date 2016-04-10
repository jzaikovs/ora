package ora

import "io"

// Lob represents oracle lob handle
type Lob struct {
	conn *Conn
	ptr  uintptr
}

func (conn *Conn) newLob() (lob *Lob, err error) {
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

// OpenReader creates reder for reading from oracle lob
func (lob *Lob) OpenReader() (lobr *LobReader, err error) {
	err = lob.conn.cerr(oci_OCILobOpen.Call(
		lob.conn.serv.ptr,
		lob.conn.err.ptr,
		lob.ptr,
		OCI_LOB_READONLY))

	return &LobReader{Lob: lob, offset: 1}, err
}

// LobReader implements reader for oracle lob reading
type LobReader struct {
	*Lob
	offset int
}

// Read reads from lob
func (lob *LobReader) Read(buf []byte) (n int, err error) {
	amount := int32(len(buf))

	err = lob.conn.cerr(oci_OCILobRead.Call(
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
		SQLCS_IMPLICIT))

	n = int(amount)
	lob.offset += n

	if n == 0 {
		err = io.EOF
	}

	return
}

// Close closes lob reader
func (lob *LobReader) Close() (err error) {
	err = lob.conn.cerr(
		oci_OCILobClose.Call(
			lob.conn.serv.ptr,
			lob.conn.err.ptr,
			lob.ptr))
	return
}
