package dockersnitch

func NewConnectionList() *ConnectionList {
	return &ConnectionList{}
}

type ConnectionStatus int

const (
	Whitelisted ConnectionStatus := iota
	Blacklisted
	Unknown
)

type ConnectionList struct {
	destIP map[[]byte]bool
}

func (*c ConnectionList) Whitelist(ip []byte) {
	c.destIP[ip] = Whitelisted
}

func (*c ConnectionList) Blacklist(ip []byte) {
	c.destIP[ip] = Blacklisted
}

func (*c ConnectionList) Check(ip []byte) ConnectionStatus {
	return c.destIP[ip]
}