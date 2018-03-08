package dockersnitch

import (
	"fmt"
	"log"
	"os"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
)

func NewIntercepter() *Intercepter {
	return &Intercepter{connList: map[string]*Connection{}}
}

type Intercepter struct {
	connList map[string]*Connection
}

func (i *Intercepter) RunMainQueue() {
	log.Printf("Running main queue")
	var err error

	nfq, err := netfilter.NewNFQueue(0, 100, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nfq.Close()
	packets := nfq.GetPackets()

	for true {
		select {
		case p := <-packets:
			i.handlePacket(&p)
		}
	}
}

func (i *Intercepter) handlePacket(p *netfilter.NFPacket) {
	dst := (*p).Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String()

	c, ok := i.connList[dst]
	if ok == false {
		c = &Connection{Status: Unitialized, NFPacket: p, Queue: make(chan *netfilter.NFPacket, 100)}
		i.connList[dst] = c
	}

	switch c.Status {
	case Whitelisted:
		log.Printf("Whitelisted dst: %s", dst)
		p.SetVerdict(netfilter.NF_ACCEPT)
	case Blacklisted:
		log.Printf("Blacklisted dst: %s", dst)
		p.SetVerdict(netfilter.NF_DROP)
	case Prompting:
		c.QueueNFPacket(p)
	case Unitialized:
		log.Printf("Unitialized dst: %s", dst)
		c.Prompt(p)
	default:
		log.Print("This shouldn't happen: %v", (*p))
	}
}
