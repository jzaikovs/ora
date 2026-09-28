package ora

import "testing"

func TestParseConnectString(t *testing.T) {
	cases := []struct {
		in, user, pass, db string
	}{
		{"user/pass@//host:1521/service", "user", "pass", "//host:1521/service"},
		{"user:pass@//host:1521/service", "user", "pass", "//host:1521/service"},
		{"user/pass@host:1521/service", "user", "pass", "host:1521/service"},
		{"user/pass@tnsname", "user", "pass", "tnsname"},
		{"user/p@ss@//host:1521/service", "user", "p@ss", "//host:1521/service"},
		{"user/p/a:ss@//host:1521/service", "user", "p/a:ss", "//host:1521/service"},
		{"user:p/a:ss@//host:1521/service", "user", "p/a:ss", "//host:1521/service"},
		{"user/@tnsname", "user", "", "tnsname"},
		{"@tnsname", "", "", "tnsname"},
		{"//host:1521/service", "", "", "//host:1521/service"},
	}

	for _, c := range cases {
		user, pass, db, err := ParseConnectString(c.in)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if user != c.user || pass != c.pass || db != c.db {
			t.Errorf("%q: got (%q, %q, %q), want (%q, %q, %q)", c.in, user, pass, db, c.user, c.pass, c.db)
		}
	}

	for _, in := range []string{"", "tnsname", "user@tnsname", "user/pass@"} {
		if _, _, _, err := ParseConnectString(in); err == nil {
			t.Errorf("%q: expected error", in)
		}
	}
}
