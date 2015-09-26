package main

import (
	"fmt"
	"os"
)

// Formats according to a format specifier and writes to standard error and
// exits the application with a non-zero exit code.
func Fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

// Writes the string representation of the specified interface to standard
// error and exits the application with a non-zero exit code.
func Fatal(a interface{}) {
	Fatalf("%s\n", a)
}
