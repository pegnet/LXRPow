// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"fmt"

	"github.com/pegnet/LXRPow/pow"
)

// All that is needed to create a Hasher instance.  Once it is created,
// It will run and feed back solutions
type Hashers struct {
	Instances   []*Hasher
	BlockHashes chan []byte
	Solutions   chan PoWSolution
	Control     chan string
	Nonce       uint64
	Lx          *pow.LxrPow
	Started     bool
}

// NewHashers
// Create a set of Hashers to utilize the available cores on a computer
// system.  A Solutions Channel is used by all the go routines to report
// their best results.
func NewHashers(Instances int, Nonce uint64, Lx *pow.LxrPow) *Hashers {
	h := new(Hashers)
	h.Nonce = Nonce
	h.Lx = Lx
	h.BlockHashes = make(chan []byte, 1)
	h.Solutions = make(chan PoWSolution, 1)
	h.Control = make(chan string,1)

	for i := 0; i < Instances; i++ {
		n := h.Nonce
		n = n<<19 ^ n>>11 ^ uint64((i + 17))

		instance := NewHasher(n, Lx)
		instance.Solutions = h.Solutions // override Solutions channel

		h.Instances = append(h.Instances, instance) // Collect all our instances

		instance.Start()
	}

	return h
}

func (h *Hashers) Stop() {
	if !h.Started {
		return
	}
	h.Control <- "stop"
	h.Started = false
	fmt.Println("\nStopping All Hashers")
	for _, i := range h.Instances {
		i.Stop()
	}
}

func (h *Hashers) Start() {
	if h.Started {
		return
	}
	h.Started = true

   	go func() {
		for {
			select {
			case hash := <-h.BlockHashes:
				for _, i := range h.Instances {
					i.BlockHashes <- hash
				}
			case cmd := <-h.Control:
				if cmd == "stop" {
					return
				}
			}
		}
	}()
}
