package ora

import (
	"database/sql/driver"
)

// Result implements driver.Result interface
type Result struct {
	rowsAffected int64
	err          error
}

// LastInsertId implements driver.Result interface
func (result Result) LastInsertId() (int64, error) {
	return driver.RowsAffected(0).LastInsertId()
}

// RowsAffected returns affected binds count
func (result Result) RowsAffected() (int64, error) {
	return result.rowsAffected, result.err
}

