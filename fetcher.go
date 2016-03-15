package ora

import "database/sql/driver"

// QueryResult handles query result, it adds more functions for result than standart database/sql
type QueryResult struct {
	stmt    driver.Stmt
	rows    *Rows
	dest    []driver.Value
	fetched int
}

func newQueryResult(rows *Rows, stmt driver.Stmt) *QueryResult {
	qr := new(QueryResult)
	qr.rows = rows
	qr.stmt = stmt
	qr.dest = make([]driver.Value, len(qr.rows.descr))
	return qr
}

// Next fetchers next row in query result
func (qr *QueryResult) Next() error {
	qr.dest = make([]driver.Value, len(qr.rows.descr))
	return qr.rows.Next(qr.dest)
}

func (qr *QueryResult) Close() (err error) {
	qr.stmt.Close()
	return qr.rows.Close()
}

func (qr *QueryResult) Scan(x ...interface{}) (err error) {
	for i, v := range qr.dest {
		x[i] = v
	}
	return
}

func (qr *QueryResult) Values() (row []interface{}, err error) {
	row = make([]interface{}, len(qr.dest))
	for i, v := range qr.dest {
		row[i] = v
	}
	return
}

func (qr *QueryResult) FieldDescriptions() (fields []*Descriptor) {
	return qr.rows.descr
}
