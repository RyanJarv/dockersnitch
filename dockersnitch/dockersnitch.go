package dockersnitch

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	//"os/signal"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
	"github.com/janeczku/go-ipset/ipset"
)

func NewIntercepter(p string, nfqueue uint16, bl *ipset.IPSet, wl *ipset.IPSet) *Intercepter {
	i := &Intercepter{
		socket:    p,
		connList:  map[string]*Connection{},
		nfQueue:   nfqueue,
		blacklist: bl,
		whitelist: wl,
	}
	i.ListenSocket()
	return i
}

type Intercepter struct {
	socket    string
	stream    net.Conn
	nfQueue   uint16
	connList  map[string]*Connection
	blacklist *ipset.IPSet
	whitelist *ipset.IPSet
}

func (i *Intercepter) RunMainQueue() {
	log.Printf("Running main queue")
	nfq, err := netfilter.NewNFQueue(i.nfQueue, 100, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nfq.Close()

	signal_channel := make(chan os.Signal)
	signal.Notify(signal_channel, os.Interrupt)
	packets := nfq.GetPackets()
	for {
		select {
		case p := <-packets:
			go i.handlePacket(&p)
		case <-signal_channel:
			log.Printf("Cleaning up Dockersnitch.Intercepter")
			i.Teardown()
			log.Printf("Done tearing down Dockersnitch.Intercepter")
			return
		}
	}
}

func (i *Intercepter) Teardown() {
	os.Remove(i.socket)
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

func (i *Intercepter) ListenSocket() {
	l, err := net.Listen("unix", i.socket)
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
