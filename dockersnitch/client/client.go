package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Ask struct {
	output io.Writer
	input  io.Reader
}

func (a *Ask) Write(b []byte) (int, error) {
	addr := strings.TrimSuffix(string(b), "\n")
	q := fmt.Sprintf("\nAllow connection from %s? [w/b] ", addr)
	_, err := a.output.Write([]byte(q))
	return len(b), err
}

func (a *Ask) Read(b []byte) (int, error) {
	return a.input.Read(b)
}

func Client(network, address string) {
	var server net.Conn
	ask := &Ask{output: os.Stdout, input: os.Stdin}
	for {
		log.Printf("Attempting to connect to %s %s", network, address)
		var err error
		server, err = net.Dial(network, address)
		if err == nil {
			s := bufio.NewReader(server)
			line, _, err := s.ReadLine()
			if err == nil && string(line) == "ready" {
				break
			}
		}
		time.Sleep(time.Millisecond * 250)
	}

	go func() {
		if _, err := io.Copy(server, ask); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if _, err := io.Copy(ask, server); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("Done setting up")
}
