package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func parseArgs(args []string, timeout *time.Duration, host *string, port *int) error {
	fs := flag.NewFlagSet("telnet", flag.ContinueOnError)
	fs.SetOutput(bytes.NewBuffer(nil))

	fs.DurationVar(timeout, "timeout", 10*time.Second, "Connection timeout")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", ErrParseArgs, err)
	}

	if fs.NArg() != 2 {
		return fmt.Errorf("%w: %w", ErrParseArgs, errors.New("not enough arguments"))
	}

	h := fs.Arg(0)
	b, _ := regexp.MatchString(`^[A-Za-z0-9-.]+$`, h)
	if !b {
		return fmt.Errorf("%w %w", ErrParseArgs, errors.New("host: string \""+h+"\" cannot be a host name or address"))
	}
	*host = h

	p, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return fmt.Errorf("%w %w", ErrParseArgs, fmt.Errorf("port: %w", err))
	}
	*port = p

	return nil
}
