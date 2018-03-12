package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/RyanJarv/dockersnitch/dockersnitch"
	ns "github.com/RyanJarv/dockersnitch/dockersnitch/netns"
)

func main() {
	netns := ns.NewNetNS()
	// Server needs to be started in the original net namespace for port forwarding to work
	//go dclient.Client(network, address)
	network, address := "tcp", "0.0.0.0:33504"
	server, client := net.Pipe()
	Server(client, network, address)

	netns.SwitchToRoot()
	iptables := dockersnitch.IPTables{
		Chain:   "DOCKERSNITCH",
		NFQueue: 4031,
	}
	iptables.Setup()

	i := dockersnitch.NewIntercepter(server, iptables.NFQueue, iptables.Blacklist, iptables.Whitelist)

	wg := runOnCtrlC(func() {
		netns.SwitchToRoot()
		iptables.Teardown()
		i.Teardown()
		netns.Restore()
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
