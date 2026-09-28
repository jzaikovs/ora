package ora

import (
	"errors"
	"strings"
)

// ParseConnectString splits connect string into username, password and database.
//
// Supported forms:
//
//	user/password@//host:port/service  (EZConnect)
//	user:password@//host:port/service
//	user/password@tnsname
//	@database or //host:port/service   (no credentials)
//
// Database part is split on the last '@', so passwords may contain '@'.
// Credentials are split on the first '/' or ':', unquoted Oracle user names
// can't contain those, so passwords may contain '/' and ':'.
func ParseConnectString(connectString string) (username, password, database string, err error) {
	if len(connectString) == 0 {
		return "", "", "", errors.New("empty connect string")
	}

	at := strings.LastIndexByte(connectString, '@')
	if at < 0 {
		if strings.HasPrefix(connectString, "//") {
			return "", "", connectString, nil
		}
		return "", "", "", errors.New("unsupported connect string, expected user/password@database")
	}

	credentials, database := connectString[:at], connectString[at+1:]
	if len(database) == 0 {
		return "", "", "", errors.New("unsupported connect string, missing database after '@'")
	}

	if len(credentials) == 0 {
		return "", "", database, nil
	}

	sep := strings.IndexAny(credentials, "/:")
	if sep < 0 {
		return "", "", "", errors.New("unsupported connect string, missing password")
	}

	return credentials[:sep], credentials[sep+1:], database, nil
}
