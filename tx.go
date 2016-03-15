package ora

// Transaction handler
type Transaction struct {
	conn *Conn
}

// Commit implements transaction commit
func (tx *Transaction) Commit() error {
	return tx.conn.cerr(oci_OCITransCommit.Call(tx.conn.serv.ptr, tx.conn.err.ptr, OCI_DEFAULT))
}

// Rollback implements transaction rollback
func (tx *Transaction) Rollback() error {
	return tx.conn.cerr(oci_OCITransRollback.Call(tx.conn.serv.ptr, tx.conn.err.ptr, OCI_DEFAULT))
}
