package common

import (
	"fmt"
	"strconv"
)

func FormatFloat64(val float64, precision int) string {
	formatter := "%." + strconv.Itoa(precision) + "f"
	return fmt.Sprintf(formatter, val)
}

func BytesToMB(b int64) string {
	val := float64(b) / float64(MBtoBytes(1))
	return FormatFloat64(val, 2) + "MB"
}
