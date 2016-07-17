package main

import (
	"errors"
	"os"
	"path"
)

type Flags struct {
	OutputDir       string
	AccountingFiles []string
}

// Parse command line flags.
// Returns the parsed flags or an error.
func ParseFlags() (*Flags, error) {
	const (
		stateNil byte = '-'
		stateOut      = 'o'
	)
	f := new(Flags)
	state := stateNil
	progName := path.Base(os.Args[0])

	usage := func() (*Flags, error) {
		return nil, errors.New("usage: " + progName + ` -o output-dir accounting-files

-h, --help:
    Display this message.
-o, --output:
    Directory where reports will be written.
`)
	}

	for _, arg := range os.Args[1:] {
		switch state {
		case stateOut:
			if f.OutputDir != "" {
				return usage()
			}
			f.OutputDir = arg
			state = stateNil
		default:
			switch arg {
			case "-h", "--help":
				return usage()
			case "-o", "--output":
				state = stateOut
			default:
				f.AccountingFiles = append(f.AccountingFiles, arg)
			}
		}
	}

	if f.OutputDir == "" || f.AccountingFiles == nil {
		return usage()
	}

	return f, nil
}
