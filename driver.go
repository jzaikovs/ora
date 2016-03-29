package ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"os"
	"regexp"
)

var (
	trace            = log.New(os.Stdout, "TRACE:", log.Lshortfile)
	patternEzconnect = regexp.MustCompile(`^((.*?)/(.*?))?(@|//)(.*/.*)$`)
)

func init() {
	sql.Register("ora", driverStruct{})
}

type driverStruct struct {
}

// Function implements driver.Open interface
func (driverStruct) Open(connectionString string) (driver.Conn, error) {
	return Open(connectionString)
}

// Open creates new connection
func Open(connectionString string) (*Conn, error) {
	if len(connectionString) == 0 {
		return nil, errors.New("empty connect string")
	}

	// for now support only ezconnect connect string
	matches := patternEzconnect.FindSubmatch([]byte(connectionString))
	if len(matches) == 0 {
		return nil, errors.New("only ezconnect connect string is supported")
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

	//logLine("connected...")
	return conn, nil
}
