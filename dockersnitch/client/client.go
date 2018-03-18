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
	c.server.Close()

	log.Printf("Closing")
	//Hack so we are not reading a line from Stdin
	//c.ask.Write([]byte("\n"))
	//c.ask.Read(make([]byte, 0))
}

func (c *Client) Start(network, address string) {
	c.onCtrlC()
	//c.ask = &Ask{output: os.Stdout, input: os.Stdin}
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

	go func() (err error) {
		serverR := bufio.NewReader(bufio.NewReader(c.server))
		for {
			log.Printf("client readline")
			var line []byte
			if line, _, err = serverR.ReadLine(); err != nil {
				log.Fatal(err)
			}
			resp := c.Ask(fmt.Sprintf("Allow connection from %s? [w/b] ", line))
			log.Printf("client writeline")
			fmt.Fprintln(c.server, resp)
			log.Printf("client writeline done")
		}
	}()

	log.Printf("Done setting up")
}
