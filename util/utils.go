package util

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/genproto/googleapis/type/datetime"
)

func ReqId(chId string) string {
	reqId := chId + time.Now().Format("20060102150405")
	return reqId
}

func ParseIntToRupiah(angka int) string {
	rupiah := ""
	angkaRev := reverseInt(angka)
	for i := 0; i < len(angkaRev); i++ {
		if i%3 == 0 && i != 0 {
			rupiah += "."
		}
		rupiah += string(angkaRev[i])
	}
	return "Rp " + reverseString(rupiah)
}

func reverseInt(angka int) string {
	angkaStr := string(angka)
	angkaRev := ""
	for i := len(angkaStr) - 1; i >= 0; i-- {
		angkaRev += string(angkaStr[i])
	}
	return angkaRev
}

func reverseString(str string) string {
	strRev := ""
	for i := len(str) - 1; i >= 0; i-- {
		strRev += string(str[i])
	}
	return strRev
}

func FormatRupiah(amount float64) string {
	strAmount := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(strAmount, ".")
	intPart := parts[0]
	decimalPart := parts[1]

	// Sisipkan tanda titik setiap 3 digit dari belakang
	n := len(intPart)
	if n > 3 {
		for i := n - 3; i > 0; i -= 3 {
			intPart = intPart[:i] + "." + intPart[i:]
		}
	}

	// Gabungkan bagian integer dan desimal dengan tanda koma
	rupiah := "Rp " + intPart + "," + decimalPart
	return rupiah
}

func ToTime(dt *datetime.DateTime) (time.Time, error) {
	if dt == nil {
		now := time.Now()

		dt = &datetime.DateTime{
			Year:    int32(now.Year()),
			Month:   int32(now.Month()),
			Day:     int32(now.Day()),
			Hours:   int32(now.Hour()),
			Minutes: int32(now.Minute()),
			Seconds: int32(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		}
	}

	res := time.Date(int(dt.Year), time.Month(dt.Month), int(dt.Day),
		int(dt.Hours), int(dt.Minutes), int(dt.Seconds), int(dt.Nanos), time.UTC)

	return res, nil
}

func CurrentDatetime() *datetime.DateTime {
	now := time.Now()

	return &datetime.DateTime{
		Year:       int32(now.Year()),
		Month:      int32(now.Month()),
		Day:        int32(now.Day()),
		Hours:      int32(now.Hour()),
		Minutes:    int32(now.Minute()),
		Seconds:    int32(now.Second()),
		Nanos:      int32(now.Second()),
		TimeOffset: &datetime.DateTime_UtcOffset{},
	}
}
