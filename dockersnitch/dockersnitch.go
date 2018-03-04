package dockersnitch

import (
	"fmt"
	"os"

	"github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
)

func Run() {
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
			HandlePacket(p)
		}
	}
}

func HandlePacket(p netfilter.NFPacket) {
	dst := p.Packet.(gopacket.Packet).TransportLayer().TransportFlow().Dst().Raw()
	connList := NewConnectionList()
	select connList.Check(dst) {
	case: Whitelisted
		fmt.Println("Whitelisted: %s", p.Packet)
		p.SetVerdict(netfilter.NF_ACCEPT)
	case: Blaclisted
		fmt.Println("Blacklisted: %s", p.Packet)
		p.SetVerdict(netfilter.NF_DROP)
	case: Unknown
		fmt.Println("Unknown: %s", p.Packet)
		dockersnitch.Connection.Create(p)
		//p.SetVerdict(netfilter.NF_ACCEPT)
	}
}

func HandlePacketRoutine(p gopacket.Packet) {

}