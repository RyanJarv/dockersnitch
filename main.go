package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/RyanJarv/dockersnitch/dockersnitch"
	"github.com/vishvananda/netns"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, err := netns.Get()
	if err != nil {
		log.Fatal(err)
	}

	network, address := "tcp", "0.0.0.0:33504"

	server, client := net.Pipe()

	// Server needs to be started in the original net namespace for port forwarding to work
	//go dclient.Client(network, address)
	Server(client, network, address)

	// Root net namespace needed to access host iptables, works because we are running with `--pid host` and `--privileged`
	rootns, err := netns.GetFromPid(1)
	if err != nil {
		log.Fatal(err)
	}
	defer rootns.Close()

	if err = netns.Set(rootns); err != nil {
		log.Fatal(err)
	}
	iptables := dockersnitch.IPTables{
		Chain:   "DOCKERSNITCH",
		NFQueue: 4031,
	}
	iptables.Setup()

	i := dockersnitch.NewIntercepter(server, iptables.NFQueue, iptables.Blacklist, iptables.Whitelist)

	wg := runOnCtrlC(func() {
		if err := netns.Set(rootns); err != nil {
			log.Fatal(err)
		}
		iptables.Teardown()
		i.Teardown()
		netns.Set(origns)
	})

	i.RunMainQueue()
	log.Printf("Running dockersnitch")
	wg.Wait()
}

func Server(stream net.Conn, network string, address string) {
	log.Printf("Attempting to listen on %s %s", network, address)
	var client net.Conn
	if server, err := net.Listen(network, address); err != nil {
		log.Fatal(err)
	} else {
		client, err = server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		_, err = client.Write([]byte("ready\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		if _, err := io.Copy(client, stream); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		c := bufio.NewReader(client)
		if _, err := io.Copy(stream, c); err != nil {
			log.Fatal(err)
		}
	}()
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
