package netns

import (
	"log"
	"runtime"

	"github.com/vishvananda/netns"
)

func NewNetNS() *NetNS {
	n := &NetNS{}
	n.Setup()
	return n
}

type NetNS struct {
	OrigNS netns.NsHandle
	RootNS netns.NsHandle
}

func (n *NetNS) Setup() {
	runtime.LockOSThread()

	var err error
	n.OrigNS, err = netns.Get()
	if err != nil {
		log.Fatal(err)
	}

	n.RootNS, err = netns.GetFromPid(1)
	if err != nil {
		log.Fatal(err)
	}
}

func (n *NetNS) Cleanup() {
	n.Restore()
	runtime.LockOSThread()
	n.OrigNS.Close()
	n.RootNS.Close()
}

func (n *NetNS) Restore() {
	if err := netns.Set(n.OrigNS); err != nil {
		log.Fatal(err)
	}
}

func (n *NetNS) SwitchToRoot() {
	// Root net namespace needed to access host iptables, works because we are running with `--pid host` and `--privileged`
	if err := netns.Set(n.RootNS); err != nil {
		log.Fatal(err)
	}
}
