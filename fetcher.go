package ora

import (
	"database/sql/driver"
)

// QueryResult handles query result, it adds more functions for result than standard database/sql
type QueryResult struct {
	stmt    driver.Stmt
	rows    *Rows
	lastRow []driver.Value
}

func newQueryResult(rows *Rows, stmt driver.Stmt) *QueryResult {
	qr := new(QueryResult)
	qr.rows = rows
	qr.stmt = stmt
	qr.lastRow = make([]driver.Value, len(qr.rows.descriptors))
	return qr
}

// Next fetchers next binds in query result
func (qr *QueryResult) Next() error {
	// trace.Println("qr.Next")
	qr.lastRow = make([]driver.Value, len(qr.rows.descriptors))
	return qr.rows.Next(qr.lastRow)
}

func (qr *QueryResult) Close() (err error) {
	// trace.Println("qr.Close")
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
