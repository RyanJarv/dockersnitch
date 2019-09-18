package dockersnitch

import (
	"log"
	"net"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
	"github.com/janeczku/go-ipset/ipset"
)

func NewIntercepter(s net.Conn, nfqueue uint16, bl *ipset.IPSet, wl *ipset.IPSet) *Intercepter {
	i := &Intercepter{
		Stream:    s,
		connList:  map[string]*Connection{},
		nfQueue:   nfqueue,
		blacklist: bl,
		whitelist: wl,
	}

	return i
}

type Intercepter struct {
	Stream    net.Conn // read by network or other objects
	nfQueue   uint16
	nfq       *netfilter.NFQueue
	connList  map[string]*Connection
	blacklist *ipset.IPSet
	whitelist *ipset.IPSet
}

func (i *Intercepter) RunMainQueue() {
	log.Printf("Running main queue")
	var err error
	i.nfq, err = netfilter.NewNFQueue(i.nfQueue, 100, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		log.Fatal(err)
	}

	packets := i.nfq.GetPackets()
	go func() {
		for p := range packets {
			i.handlePacket(&p)
		}
	}()
}

func (i *Intercepter) Teardown() {
	i.nfq.Close()
}

func (i *Intercepter) handlePacket(p *netfilter.NFPacket) {
	dst := p.Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String()
	log.Printf("Handling packet dst %s", dst)

	c, ok := i.connList[dst]
	if ok == false {
		c = &Connection{Status: Unitialized, NFPacket: p, Queue: make(chan *netfilter.NFPacket, 100), Dst: dst}
		i.connList[dst] = c
	}

	switch c.Status {
	case Whitelisted, Blacklisted:
		log.Printf("Connection for %s is already in state %s", dst, string(c.Status))
		c.ProcessPacket(p)
	case Prompting:
		log.Printf("Connection is Prompting, adding packet for %s to queue", dst)
		c.QueuePacket(p)
	case Unitialized:
		log.Printf("Prompting for connection with dst: %s", dst)
		c.QueuePacket(p)
		if c.Prompt(i.Stream) == Whitelisted {
			log.Printf("Whitelisting connection with dst %s", dst)
			i.whitelist.Add(dst, 0)
		} else {
			log.Printf("Blacklisting connection with dst %s", dst)
			i.blacklist.Add(dst, 0)
		}
		c.ProcessQueue()
	default:
		log.Print("This shouldn't happen: %v", (*p))
	}
}
