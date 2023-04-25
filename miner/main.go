// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	//"fmt"
	//"os"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/cfg"
	"github.com/pegnet/LXRPow/mine"
	"github.com/pegnet/LXRPow/sim"
	"github.com/pegnet/LXRPow/validator"
)

func main() {
	c := cfg.NewConfig()

	var validatorList []*validator.Validator
	for i := 0; i < 1; i++ { // Just running one validator for now
		v := validator.NewValidator(sim.GetURL(),c.LX)
		accumulate.MiningADI.RegisterMiner(v.URL)
		validatorList = append(validatorList, v)
	}
	for _, v := range validatorList {
		go v.Start()
	}

	minerList := make(map[string]*mine.Miner)
	for j := 0; j < c.MinerCnt; j++ {
		m := new(mine.Miner)
		m.Init(c)
		c.TokenURL = sim.GetURL() + "/tokens"
		if _, exists := minerList[m.Cfg.TokenURL]; exists {
			panic("duplicate miner url " + m.Cfg.TokenURL)
		}
		minerList[m.Cfg.TokenURL] = m
		go m.Run()
		c = c.Clone()
	}

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the mining simulation...\n")

		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	for {
		time.Sleep(time.Second)
	}

}

// interruptChannel is used to receive SIGINT (Ctrl+C) signals.
var interruptChannel chan os.Signal

// addHandlerChannel is used to add an interrupt handler to the list of handlers
// to be invoked on SIGINT (Ctrl+C) signals.
var addHandlerChannel = make(chan func())

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the
// interruptChannel and invokes the registered interruptCallbacks accordingly.
// It also listens for callback registration.  It must be run as a goroutine.
func mainInterruptHandler() {
	// interruptCallbacks is a list of callbacks to invoke when a
	// SIGINT (Ctrl+C) is received.
	var interruptCallbacks []func()

	// isShutdown is a flag which is used to indicate whether or not
	// the shutdown signal has already been received and hence any future
	// attempts to add a new interrupt handler should invoke them
	// immediately.
	var isShutdown bool

	for {
		select {
		case <-interruptChannel:
			// Ignore more than one shutdown signal.
			if isShutdown {
				fmt.Println("Ctrl+C Already being processed!")
				continue
			}
			isShutdown = true
			fmt.Println("Received SIGINT (Ctrl+C).  Shutting down...")

			// Run handlers in LIFO order.
			for i := range interruptCallbacks {
				idx := len(interruptCallbacks) - 1 - i
				callback := interruptCallbacks[idx]
				callback()
			}

		case handler := <-addHandlerChannel:
			// The shutdown signal has already been received, so
			// just invoke and new handlers immediately.
			if isShutdown {
				handler()
			}

			interruptCallbacks = append(interruptCallbacks, handler)
		}
	}
}

// AddInterruptHandler adds a handler to call when a SIGINT (Ctrl+C) is
// received.
func AddInterruptHandler(handler func()) {
	// Create the channel and start the main interrupt handler which invokes
	// all other callbacks and exits if not already done.
	if interruptChannel == nil {
		interruptChannel = make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)
		go mainInterruptHandler()
	}

	addHandlerChannel <- handler
}
