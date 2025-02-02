package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		addr:    address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

type telnetClient struct {
	timeout time.Duration
	addr    string
	conn    net.Conn
	in      io.ReadCloser
	out     io.Writer
}

func (tc *telnetClient) Connect() (err error) {
	tc.conn, err = net.DialTimeout("tcp", tc.addr, tc.timeout)
	return
}

func (tc *telnetClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}

func (tc *telnetClient) Send() error {
	if tc.conn == nil {
		return ErrConnNotEstablish
	}

	scanner := bufio.NewScanner(tc.in)
	for scanner.Scan() {
		mess := append(scanner.Bytes(), '\n')
		if _, err := tc.conn.Write(mess); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (tc *telnetClient) Receive() error {
	if tc.conn == nil {
		return ErrConnNotEstablish
	}

	reader := bufio.NewReader(tc.conn)
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		_, err = tc.out.Write(message)
		if err != nil {
			return err
		}
	}
}
