package ora

import "database/sql/driver"

// Transaction handler
type Transaction struct {
	conn *Conn
}

// Commit implements transaction commit
func (tx *Transaction) Commit() error {
	if tx.conn.bad {
		tx.conn.tx = nil
		return driver.ErrBadConn // Not committed; free() rolls back before logoff.
	}
	err := tx.conn.cerr(oci_OCITransCommit.Call(tx.conn.serv.ptr, tx.conn.err.ptr, OCI_DEFAULT))
	if err != nil {
		tx.conn.bad = true
	}
	tx.conn.tx = nil
	return err
}

// Rollback implements transaction rollback
func (tx *Transaction) Rollback() error {
	if tx.conn.bad {
		tx.conn.tx = nil
		return nil // The interrupted connection is never pooled; free() rolls back before logoff.
	}
	err := tx.conn.cerr(oci_OCITransRollback.Call(tx.conn.serv.ptr, tx.conn.err.ptr, OCI_DEFAULT))
	if err != nil {
		tx.conn.bad = true
	}
	tx.conn.tx = nil
	return err
}
