package dockersnitch

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
	"github.com/janeczku/go-ipset/ipset"
)

const NFQueueNum = 3413

func NewIntercepter(p string, bl *ipset.IPSet, wl *ipset.IPSet) *Intercepter {
	i := &Intercepter{
		connList:  map[string]*Connection{},
		blacklist: bl,
		whitelist: wl,
	}
	i.ListenSocket(p)
	return i
}

type Intercepter struct {
	stream    net.Conn
	connList  map[string]*Connection
	blacklist *ipset.IPSet
	whitelist *ipset.IPSet
}

func (i *Intercepter) RunMainQueue() {
	log.Printf("Running main queue")
	var err error

	nfq, err := netfilter.NewNFQueue(NFQueueNum, 100, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nfq.Close()
	packets := nfq.GetPackets()
	for p := range packets {
		go i.handlePacket(&p)
	}
}

func (i *Intercepter) handlePacket(p *netfilter.NFPacket) {
	dst := p.Packet.(gopacket.Packet).NetworkLayer().NetworkFlow().Dst().String()

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
		if c.Prompt(i.stream) == Whitelisted {
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

func (i *Intercepter) ListenSocket(path string) {
	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go func() {
		defer l.Close()
		for {
			s, err := l.Accept()
			if err != nil {
				log.Fatal("accept error:", err)
			}
			if i.stream == nil {
				i.stream = s
			} else if _, err := i.stream.Read([]byte{}); err != nil {
				i.stream = s
			} else {
				s.Write([]byte("Connection already open"))
				s.Close()
			}
		}
	}()
}
