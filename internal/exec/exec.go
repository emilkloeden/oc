package exec

import (
	"bytes"
	"io"
	"os"
	osexec "os/exec"
)

type Options struct {
	Dir    string
	Env    []string // appended to os.Environ() when set
	Stdout io.Writer
	Stderr io.Writer
}

func Run(name string, args []string, opts Options) error {
	cmd := osexec.Command(name, args...)

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func Output(name string, args []string, opts Options) (string, error) {
	var buf bytes.Buffer
	opts.Stdout = &buf
	if err := Run(name, args, opts); err != nil {
		return "", err
	}
	return buf.String(), nil
}
