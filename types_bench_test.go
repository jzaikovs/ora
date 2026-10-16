package ora

import (
	"testing"
)

func BenchmarkNumToString(b *testing.B) {
	buf := []byte{62, 98, 87, 86, 9, 36, 66, 12, 22, 69, 63, 55, 75, 58, 63, 69, 22, 51, 73, 17, 81}
	for i := 0; i < b.N; i++ {
		convertNumberToString(buf)
	}
}
