package ora

import (
	"database/sql"
	"database/sql/driver"
	"io/ioutil"
	"log"
)

// DriverName is name used to register driver
const DriverName = "ora"

var (
	// trace = log.New(os.Stdout, "DEBUG:", log.Lshortfile)
	trace = log.New(ioutil.Discard, "DEBUG:", log.Lshortfile)
)

func init() {
	sql.Register(DriverName, Driver{})
}

type ConnStd struct {
	*Conn
}

type Driver struct {
}

// Open implements driver.Open interface
func (Driver) Open(connectionString string) (driver.Conn, error) {
	conn, err := Open(connectionString)
	if err != nil {
		return nil, err
	}

	return &ConnStd{conn}, nil
}

// Open creates new connection
func Open(connectionString string) (*Conn, error) {
	username, password, database, err := ParseConnectString(connectionString)
	if err != nil {
		return nil, err
	}

	// create connection and logon
	conn, err := newConnection()
	if err != nil {
		return nil, err
	}

	if err = conn.logon([]byte(username), []byte(password), []byte(database)); err != nil {
		return nil, err
	}

	return conn, nil
}
