package dockersnitch

import	(
	"github.com/google/gopacket"
	"errors"
)

type ConnectionStatus int

const (
	Whitelisted ConnectionStatus := iota
	Blacklisted
	InProgress
	Unknown
)

type Connection struct {
	queue chan gopacket.Packet
	packet gopacket.Packet
	status ConnectionStatus
}

func (c *Connection) Create(p packet.Packet) {
	c.packet = p
	c.status = InProgress
	c.queue = make(chan gopacket.Packet, 100)
	c.Prompt(p)
}

func (c *Connection) Prompt(p gopacket.Packet) {
	return
}

func (c *Connection) Whitelist(ip []byte) {
	c.status = Whitelisted
}

func (c *Connection) Blacklist(ip []byte) {
	c.status = Blacklisted
}

func (c *Connection) SetInProgress(*Connection) {
	c.status = InProgress
}

func (c *Connection) Status() ConnectionStatus {
	return c.Status
}

func (*c ConnectionList) Queue(p gopacket.Packet) error {
	select {
	case c.queue <- p:
		return nil
	default:
		return errors.New("Queue is full")
	}
}