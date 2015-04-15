package ora

import (
	"database/sql/driver"
)

// Result implements driver.Result interface
type Result struct {
	affectedRows int64
}

// LastInsertId returns last inserted id
// TODO: this is not implemented
func (result Result) LastInsertId() (int64, error) {
	return 0, driver.ErrSkip
}

// RowsAffected returns affected row count
// TODO: this is not implemented
func (result Result) RowsAffected() (int64, error) {
	return 0, driver.ErrSkip
}
