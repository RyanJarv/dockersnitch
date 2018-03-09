package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/RyanJarv/dockersnitch/dockersnitch"
)

func main() {
	iptables := dockersnitch.IPTables{
		Chain:   "DOCKERSNITCH",
		NFQueue: 4031,
	}
	iptables.Setup()
	socket := "/var/run/dockersnitch.sock"
	i := dockersnitch.NewIntercepter(socket, iptables.NFQueue, iptables.Blacklist, iptables.Whitelist)

	wg := runOnCtrlC(func() {
		iptables.Teardown()
		i.Teardown()
	})

	go i.RunMainQueue()
	var s net.Conn
	for {
		var err error
		s, err = net.Dial("unix", socket)
		if err == nil {
			break
		}
		log.Printf("Waiting for %s", socket)
		time.Sleep(time.Millisecond * 500)
	}
	go io.Copy(s, os.Stdout)
	go io.Copy(os.Stdin, s)
	wg.Wait()
}

func runOnCtrlC(c func()) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	signal_channel := make(chan os.Signal)
	signal.Notify(signal_channel, os.Interrupt)
	wg.Add(1)
	go func() {
		<-signal_channel
		c()
		wg.Done()
	}()
	return wg
}
