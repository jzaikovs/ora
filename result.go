package ora

import (
	"database/sql/driver"
)

type t_result struct {
	affted_rows int64
}

func (self t_result) LastInsertId() (int64, error) {
	return 0, driver.ErrSkip
}

func (self t_result) RowsAffected() (int64, error) {
	return 0, driver.ErrSkip
}
