package ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io/ioutil"
	"log"
	"regexp"
)

// DriverName is name used to register driver
const DriverName = "ora"

var (
	// trace = log.New(os.Stdout, "DEBUG:", log.Lshortfile)
	trace            = log.New(ioutil.Discard, "DEBUG:", log.Lshortfile)
	patternEZConnect = regexp.MustCompile(`^((.*?)/(.*?))?(@|//)(.*(/.*)?)$`)
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
	if len(connectionString) == 0 {
		return nil, errors.New("empty connect string")
	}

	// for now support only ezconnect connect string
	matches := patternEZConnect.FindSubmatch([]byte(connectionString))
	if len(matches) == 0 {
		return nil, errors.New("unsupported connect string")
	}

	username := matches[2]
	password := matches[3]
	database := matches[5]

	// create connection and logon
	conn, err := newConnection()
	if err != nil {
		return nil, err
	}

	if err = conn.logon(username, password, database); err != nil {
		return nil, err
	}

	return conn, nil
}
