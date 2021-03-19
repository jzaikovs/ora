package ora

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// convert from oracle date to time.Time
//docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16288
func convertDateToGoTime(p []byte) time.Time {
	year := int(p[0]-100)*100 + int(p[1]-100)
	month := time.Month(int(p[2]))
	return time.Date(year, month, int(p[3]), int(p[4]-1), int(p[5]-1), int(p[6]-1), 0, time.Local)
}

// https://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16276
func convertNumberToString(num []byte) string {
	if len(num) == 1 && num[0] == 128 {
		return "0"
	}

	buf := bytes.NewBuffer(nil)
	buf.Grow(len(num)*2 + 1)

	var exp int8

	isPositive := num[0]&0x80 != 0
	if isPositive {
		exp = int8(num[0])&0x7f - 65
		for i := 1; i < len(num); i++ {
			num[i] = num[i] - 1
		}
	} else {
		exp = int8(^num[0] - 65 - 128)
		if len(num) < 21 {
			num = num[:len(num)-1]
		}
		for i := 1; i < len(num); i++ {
			num[i] = 101 - num[i]
		}
		buf.WriteByte('-')
	}

	num = num[1:]
	decimal := false

	if exp == 0 {
		for i, n := range num {
			if i == 0 {
				fmt.Fprintf(buf, "%v.", n)
				decimal = true
			} else {
				fmt.Fprintf(buf, "%02v", n)
			}
		}
	} else if exp < 0 {
		buf.WriteString("0.")
		decimal = true
		exp++
		for exp < 0 {
			buf.WriteString("00")
			exp++
		}

		for _, n := range num {
			fmt.Fprintf(buf, "%02v", n)
		}
	} else if exp > 0 {
		for i, n := range num {
			if i == 0 {
				fmt.Fprintf(buf, "%v", n)
			} else {
				fmt.Fprintf(buf, "%02v", n)
			}
			if exp == 0 {
				buf.WriteByte('.')
				decimal = true
			}
			exp--
		}

		for exp >= 0 {
			fmt.Fprint(buf, "00")
			exp--
		}
	}

	if decimal {
		return strings.TrimRight(strings.TrimRight(buf.String(), "0"), ".")
	}

	return buf.String()
}

func timeToOraBytes(val time.Time) []byte {
	return []byte{
		byte(val.Year()/100) + 100,
		byte(val.Year()%100) + 100,
		byte(val.Month()),
		byte(val.Day()),
		byte(val.Hour()) + 1,
		byte(val.Minute()) + 1,
		byte(val.Second()) + 1,
	}
}
