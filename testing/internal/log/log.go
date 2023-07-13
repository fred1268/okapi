package log

import "fmt"

func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func Fatalf(format string, args ...any) {
	format = fmt.Sprintf("FAIL\t%s", format)
	fmt.Printf(format, args...)
}
