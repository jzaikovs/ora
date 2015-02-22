package ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
)

var (
	ezcon_pattern = regexp.MustCompile(`^((.*?)/(.*?))?(@|//)(.*/.*)$`)
)

func init() {
	sql.Register("ora", t_driver{})
}

type t_driver struct {
}

// Function implements driver.Open interface
func (self t_driver) Open(connect_string string) (driver.Conn, error) {
	if len(connect_string) == 0 {
		return nil, errors.New("empty connect string")
	}

	// for now support only ezconnect connect string
	matches := ezcon_pattern.FindSubmatch([]byte(connect_string))
	if len(matches) == 0 {
		return nil, errors.New("only ezconnect connect string is supported")
	}

	username := matches[2]
	password := matches[3]
	database := matches[5]

	//logLine("db=", string(database), " user=", string(username), " pass=", string(password))

	// create connection and logon
	conn, err := new_conn()
	if err != nil {
		return nil, err
	}

	if err = conn.logon(username, password, database); err != nil {
		return nil, err
	}

	//logLine("connected...")
	return conn, nil
}
