package dockersnitch

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
)

type ConnectionStatus int

const (
	Whitelisted ConnectionStatus = iota
	Blacklisted
	Prompting
	Unitialized
)

// NewConnection ...
// We're still in the main loop here so return as soon as possible
func NewConnection(p *netfilter.NFPacket) *Connection {
	c := &Connection{}
	c.Status = Unitialized
	c.NFPacket = p
	c.Queue = make(chan *netfilter.NFPacket, 100)
	return c
}

type Connection struct {
	Queue    chan *netfilter.NFPacket
	NFPacket *netfilter.NFPacket
	Status   ConnectionStatus
}

func (c *Connection) Prompt(p *netfilter.NFPacket) ConnectionStatus {
	c.Status = Prompting
	dst := (*p).Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String()
	log.Printf("Prompting on dst %s", dst)
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("New connection to %s found, is this expected? [yes/no] ", dst)
	resp, _, err := r.ReadLine()
	if err != nil {
		log.Fatal(err)
	}
	if string(resp) == "yes" {
		c.Whitelist(p)
		return Whitelisted
	} else {
		c.Whitelist(p)
		//c.Blacklist(p)
		return Blacklisted
	}
}

func (c *Connection) Whitelist(p *netfilter.NFPacket) {
	log.Printf("Setting whitelist for dst %s", (*p).Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String())
	(*p).SetVerdict(netfilter.NF_ACCEPT)
	select {
	case p := <-c.Queue:
		(*p).SetVerdict(netfilter.NF_ACCEPT)
	default:
	}
	c.Status = Whitelisted //Do this last to prevent any out of order packets
	return
}

func (c *Connection) Blacklist(p *netfilter.NFPacket) {
	log.Printf("Setting blacklist for dst %s", (*p).Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String())
	(*p).SetRequeueVerdict(1)
	select {
	case p := <-c.Queue:
		(*p).SetRequeueVerdict(1)
	default:
	}
	c.Status = Blacklisted //Do this last to prevent any out of order packets
	return
}

func (c *Connection) QueueNFPacket(p *netfilter.NFPacket) error {
	log.Printf("Queueing packet with dst %s", (*p).Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String())
	select {
	case c.Queue <- p:
		return nil
	default:
		return errors.New("Queue is full")
	}
}
