package ora

import (
//"database/sql/driver"
)

// transaction handler
type t_tx struct {
	conn *t_conn
}

func (self *t_tx) Commit() error {
	return self.conn.cerr(oci_OCITransCommit.Call(self.conn.serv.ptr, self.conn.err.ptr, OCI_DEFAULT))
}

func (self *t_tx) Rollback() error {
	return self.conn.cerr(oci_OCITransRollback.Call(self.conn.serv.ptr, self.conn.err.ptr, OCI_DEFAULT))
}
