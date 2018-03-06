package main

import (
	"os"
	"os/signal"
	"sync"

	"github.com/RyanJarv/dockersnitch/dockersnitch"
)

func main() {
	i := dockersnitch.NewIntercepter()

	go i.RunMainQueue()
	go i.RunDropQueue()

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
}
