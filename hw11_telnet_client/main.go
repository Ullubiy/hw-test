package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

var (
	ErrParseArgs = errors.New("parse args error")
	// соединение не установлено.
	ErrConnNotEstablish = errors.New("connection is not established")
	// завершение клиента.
	ErrClientExit = errors.New("connection was closed by client")
	// завершение сервера.
	ErrServerExit = errors.New("connection was closed by peer")
)

func main() {
	var (
		timeout time.Duration
		host    string
		port    int
	)

	if err := parseArgs(os.Args[1:], &timeout, &host, &port); err != nil {
		log.Println(err)
		log.Println("Usage: go-telnet [--timeout=<timeout>] <host> <port>")
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client := NewTelnetClient(addr, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		log.Fatalln(err)
	}

	log.Printf("...Connected to %s", addr)

	c := make(chan os.Signal, 1)
	clientErr := make(chan error)

	defer func() {
		client.Close()
		close(c)
		close(clientErr)
	}()

	signal.Notify(c, syscall.SIGINT)

	go func() {
		err := client.Send()
		if err != nil {
			clientErr <- err
			return
		}

		clientErr <- ErrClientExit
	}()

	go func() {
		err := client.Receive()
		if err != nil {
			clientErr <- err
			return
		}
		clientErr <- ErrServerExit
	}()

	select {
	case <-c:
		log.Println("...Connection was closed by", path.Base(os.Args[0]))
		return
	case err := <-clientErr:
		if errors.Is(err, ErrClientExit) {
			log.Println("...EOF")
			return
		}
		if errors.Is(err, ErrServerExit) {
			log.Println("...Connection was closed by peer")
			return
		}
		log.Println(err)
		return
	}
}
