//go:build oracle_integration || oracle_write_integration
// +build oracle_integration oracle_write_integration

package ora

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestLobReadChunks(t *testing.T) {
	for _, tc := range []struct {
		name string
		sql  string
		bind interface{}
		want []byte
	}{
		{
			name: "unicode_clob",
			sql:  "select to_clob(:1) from dual",
			bind: strings.Repeat("āčē🙂abc", 100),
			want: []byte(strings.Repeat("āčē🙂abc", 100)),
		},
		{
			name: "binary_blob",
			sql:  "select to_blob(:1) from dual",
			bind: bytes.Repeat([]byte{0, 1, 0xfe, 0xff}, 100),
			want: bytes.Repeat([]byte{0, 1, 0xfe, 0xff}, 100),
		},
		{name: "null_clob", sql: "select to_clob(cast(null as varchar2(1))) from dual"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, size := range []int{1, 7, 4096} {
				var binds []interface{}
				if tc.bind != nil {
					binds = append(binds, tc.bind)
				}
				rows, err := db.Query(tc.sql, binds...)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = rows.Close() }()
				if !rows.Next() {
					t.Fatalf("missing LOB row: %v", rows.Err())
				}
				var lob Lob
				if err := rows.Scan(&lob); err != nil {
					t.Fatal(err)
				}
				if n, err := lob.Read(nil); n != 0 || err != nil {
					t.Fatalf("empty Read = (%d, %v), want (0, nil)", n, err)
				}
				var got []byte
				buf := make([]byte, size)
				for {
					n, err := lob.Read(buf)
					got = append(got, buf[:n]...)
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatalf("buffer %d, offset %d, read %d bytes: %v", size, lob.offset, len(got), err)
					}
					if n == 0 || len(got) > len(tc.want) {
						t.Fatalf("buffer %d: invalid progress: n=%d, total=%d", size, n, len(got))
					}
				}
				if !bytes.Equal(got, tc.want) {
					t.Fatalf("buffer %d: LOB content differs (got %d bytes, want %d)", size, len(got), len(tc.want))
				}
				if err := rows.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
