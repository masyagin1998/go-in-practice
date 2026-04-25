package parse

import (
	"fmt"
	"strconv"
)

func FormatFmt(n int) string {
	return fmt.Sprintf("%d", n)
}

func FormatStrconv(n int) string {
	return strconv.Itoa(n)
}
