package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("bad_connect", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		client := NewTelnetClient("localhost:4242", 1*time.Second, io.NopCloser(in), out)

		require.EqualError(t, client.Connect(), "dial tcp [::1]:4242: connect: connection refused")
	})

	t.Run("without_connect", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		client := NewTelnetClient("localhost:4242", 5*time.Second, io.NopCloser(in), out)

		require.ErrorIs(t, client.Send(), ErrConnNotEstablish)
		require.ErrorIs(t, client.Receive(), ErrConnNotEstablish)
		require.NoError(t, client.Close())
	})

	t.Run("close_listen", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		timeout, err := time.ParseDuration("1s")
		require.NoError(t, err)

		client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		l.Close()

		conn := client.(*telnetClient).conn

		in.WriteString("hello\n")
		require.EqualError(t, client.Send(),
			"write tcp "+conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()+": write: connection reset by peer")

		// при получении возращает EOF, который перехватывается т.к. трактуется как корректное окончание
		require.NoError(t, client.Receive())
	})
}

var testParamData = []struct {
	name    string
	args    []string
	timeout time.Duration
	host    string
	port    int
	err     string
}{
	{
		name: "parseArgs_success",
		args: []string{
			"--timeout=5s",
			"localhost",
			"12345",
		},
		timeout: 5 * time.Second,
		host:    "localhost",
		port:    12345,
		err:     "",
	},
	{
		name: "parseArgs_success_without_timeout",
		args: []string{
			"localhost",
			"12345",
		},
		timeout: 10 * time.Second,
		host:    "localhost",
		port:    12345,
		err:     "",
	},
	{
		name: "parseArgs_bad_timeout",
		args: []string{
			"--timeout=5ass",
			"localhost",
			"12345",
		},
		timeout: 5 * time.Second,
		host:    "",
		port:    0,
		err:     `parse args error: invalid value "5ass" for flag -timeout: parse error`,
	},
	{
		name: "parseArgs_bad_args_count",
		args: []string{
			"--timeout=5s",
			"localhost",
		},
		timeout: 5 * time.Second,
		host:    "",
		port:    0,
		err:     "parse args error: not enough arguments",
	},
	{
		name: "parseArgs_invalid_host",
		args: []string{
			"--timeout=5s",
			"domain/localhost",
			"123",
		},
		timeout: 5 * time.Second,
		host:    "",
		port:    0,
		err:     "parse args error host: string \"domain/localhost\" cannot be a host name or address",
	},
	{
		name: "parseArgs_invalid_port",
		args: []string{
			"--timeout=5s",
			"localhost",
			"a123",
		},
		timeout: 5 * time.Second,
		host:    "localhost",
		port:    0,
		err:     "parse args error port: strconv.Atoi: parsing \"a123\": invalid syntax",
	},
}

func TestParseArgs(t *testing.T) {
	for _, data := range testParamData {
		t.Run(data.name, func(t *testing.T) {
			var (
				timeout time.Duration
				host    string
				port    int
			)

			err := parseArgs(data.args, &timeout, &host, &port)

			if data.err != "" {
				require.EqualError(t, err, data.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, data.host, host)
			require.Equal(t, data.port, port)
		})
	}
}
