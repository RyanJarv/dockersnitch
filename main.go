package main

import (
	"github.com/RyanJarv/dockersnitch/dockersnitch"
)

func main() {
	iptables := dockersnitch.IPTables{
		Chain:   "DOCKERSNITCH",
		NFQueue: 4031,
	}
	iptables.Setup()
	i := dockersnitch.NewIntercepter("/var/run/dockersnitch.sock", iptables.NFQueue, iptables.Blacklist, iptables.Whitelist)

	i.RunMainQueue()
}
