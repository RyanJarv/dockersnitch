package dockersnitch

import	"github.com/google/gopacket"

type Connection struct {
	queue chan gopacket.Packet
	packet gopacket.Packet
}

func (*c Connection) Create(p gpacket.Packet) {
	c.packet = p
}
