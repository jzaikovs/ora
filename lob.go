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
			null, null))
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
	n = len(buf)
	err = lob.conn.cerr(oci_OCILobRead2.Call(
		lob.conn.serv.ptr, // service
		lob.conn.err.ptr,  //error
		lob.ptr,           // ptr
		null,
		intRef(&n),
		uintptr(lob.offset), // offset
		bufAddr(buf),        // buffer
		uintptr(len(buf)),   // buffer length
		OCI_ONE_PIECE,
		null,
		null, // callback
		null,
		SQLCS_IMPLICIT))
	lob.offset += n
	if n == 0 || len(buf) != n {
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
