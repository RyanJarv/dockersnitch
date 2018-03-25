package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type Ask struct {
	output io.WriteCloser
	input  io.Reader
}

type Client struct {
	server net.Conn
	Ask    func(string) string
}

func (c *Client) onCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		c.Teardown()
	}()
}

func (c *Client) Teardown() {
	if c == nil {
		log.Printf("dockersnitch Client instance was nil")
		return
	}
	c.server.Close()
}

func (c *Client) Start(network, address string) {
	c.onCtrlC()
	for {
		log.Printf("Attempting to connect to %s %s", network, address)
		var err error
		c.server, err = net.Dial(network, address)
		if err == nil {
			s := bufio.NewReader(c.server)
			line, _, err := s.ReadLine()
			if err == nil && string(line) == "ready" {
				break
			}
		}
		time.Sleep(time.Millisecond * 250)
	}

	go func() {
		serverR := bufio.NewReader(bufio.NewReader(c.server))
		for {
			line, _, err := serverR.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			if len(line) == 0 {
				continue
			}
			resp := c.Ask(fmt.Sprintf("Allow connection from %s? [w/b] ", line))
			fmt.Fprintln(c.server, resp)
		}
	}()
}
