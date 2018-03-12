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

	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()

	// Create a new network namespace
	rootns, _ := netns.GetFromPid(1)
	netns.Set(rootns)
	defer rootns.Close()

	iptables := dockersnitch.IPTables{
		Chain:   "DOCKERSNITCH",
		NFQueue: 4031,
	}
	iptables.Setup()
	network, address := "tcp", "0.0.0.0:33504"

	server, client := net.Pipe()
	//go dclient.Client(network, address)
	netns.Set(origns)
	Server(client, network, address)
	netns.Set(rootns)
	i := dockersnitch.NewIntercepter(server, iptables.NFQueue, iptables.Blacklist, iptables.Whitelist)

	wg := runOnCtrlC(func() {
		netns.Set(rootns)
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
