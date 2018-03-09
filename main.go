package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/RyanJarv/dockersnitch/dockersnitch"
	"github.com/janeczku/go-ipset/ipset"
)

func main() {
	cmd := exec.Command("ipset", "create", "-exist", "dockersnitch_blacklistset", "list:set")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	blacklist, err := ipset.New("dockersnitch_blacklist", "hash:ip", ipset.Params{})
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "flush", "dockersnitch_blacklistset")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "add", "dockersnitch_blacklistset", blacklist.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("ipset", "create", "-exist", "dockersnitch_whitelistset", "list:set")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	whitelist, err := ipset.New("dockersnitch_whitelist", "hash:ip", ipset.Params{})
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "flush", "dockersnitch_whitelistset")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "add", "dockersnitch_whitelistset", blacklist.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	blacklist.Refresh([]string{})
	whitelist.Refresh([]string{})

	dockersnitch.SetupIPTables()
	i := dockersnitch.NewIntercepter("/var/run/dockersnitch.sock", blacklist, whitelist)

	go i.RunMainQueue()

	WaitForCtrlC()
}

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
	dockersnitch.TeardownIPTables()
	os.Remove("/var/run/dockersnitch.sock")
}
