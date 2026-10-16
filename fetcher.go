package ora

import (
	"database/sql"
	"database/sql/driver"
	"io"
)

// QueryResult handles query result, it adds more functions for result than standard database/sql
type QueryResult struct {
	stmt    driver.Stmt
	rows    *Rows
	lastRow []driver.Value
}

func newQueryResult(rows *Rows, stmt driver.Stmt) *QueryResult {
	return &QueryResult{
		rows:    rows,
		stmt:    stmt,
		lastRow: make([]driver.Value, len(rows.descriptors)),
	}
}

// Next fetchers next binds in query result, returns sql.ErrNoRows when there are no more rows
func (qr *QueryResult) Next() error {
	qr.lastRow = make([]driver.Value, len(qr.rows.descriptors))
	if err := qr.rows.Next(qr.lastRow); err != io.EOF {
		return err
	}
	return sql.ErrNoRows
}

func (qr *QueryResult) Close() (err error) {
	return qr.rows.Close()
}

func (qr *QueryResult) Values() (row []interface{}, err error) {
	row = make([]interface{}, len(qr.lastRow))
	for i, v := range qr.lastRow {
		row[i] = v
	}
	return
}

func (qr *QueryResult) FieldDescriptions() (fields []*Descriptor) {
	return qr.rows.descriptors
}
