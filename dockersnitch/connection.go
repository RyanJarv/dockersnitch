package dockersnitch

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/AkihiroSuda/go-netfilter-queue"
)

type ConnectionStatus string

const (
	Whitelisted = "whitelisted"
	Blacklisted = "blacklisted"
	Prompting   = "prompting"
	Unitialized = "unitialized"
)

type Connection struct {
	Queue    chan *netfilter.NFPacket
	NFPacket *netfilter.NFPacket
	Dst      string
	Status   ConnectionStatus
}

func (c *Connection) ProcessPacket(p *netfilter.NFPacket) {
	switch c.Status {
	case Whitelisted:
		c.Accept(p)
	case Blacklisted:
		c.Drop(p)
	default:
		log.Fatalf("Connection status is %s, needs to be Whitelisted or Blacklisted before running ProcessPacket.", string(c.Status))
	}
}

func (c *Connection) ProcessQueue() {
	select {
	case p := <-c.Queue:
		switch c.Status {
		case Whitelisted:
			c.Accept(p)
		case Blacklisted:
			c.Drop(p)
		default:
			log.Fatalf("Connection status is %s, needs to be Whitelisted or Blacklisted before running ProcessQueue.", string(c.Status))
		}
	default:
		log.Printf("Done processing queue")
	}
}

func (c *Connection) Prompt(f io.ReadWriteCloser) ConnectionStatus {
	c.Status = Prompting
	fmt.Printf("Prompting on dst %s", c.Dst)
	r := bufio.NewReader(f)
	for {
		f.Write([]byte(c.Dst))
		resp, _, err := r.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		switch string(resp) {
		case "w":
			c.Status = Whitelisted
		case "b":
			c.Status = Blacklisted
		default:
			log.Printf("Unexpected response %s", string(resp))
		}
		if c.Status != Prompting {
			break
		}
	}
	log.Printf("Allowing connection to %s", c.Dst)
	return c.Status
}

func (c *Connection) Accept(p *netfilter.NFPacket) {
	log.Printf("Accepting packet for dst %s", c.Dst)
	p.SetVerdict(netfilter.NF_ACCEPT)
}

func (c *Connection) Drop(p *netfilter.NFPacket) {
	log.Printf("Dropping packet for dst %s", c.Dst)
	p.SetVerdict(netfilter.NF_DROP)
}

func (c *Connection) QueuePacket(p *netfilter.NFPacket) error {
	log.Printf("Queueing packet with dst %s", c.Dst)
	select {
	case c.Queue <- p:
		return nil
	default:
		return errors.New("Queue is full")
	}
}
