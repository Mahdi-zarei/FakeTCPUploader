package common

import (
	"fmt"
	"strconv"
)

func FormatFloat64(val float64, precision int) string {
	formatter := "%." + strconv.Itoa(precision) + "f"
	return fmt.Sprintf(formatter, val)
}
