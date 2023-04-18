// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"fmt"

	"github.com/pegnet/LXRPow/pow"
)

// All that is needed to create a Hasher instance.  Once it is created,
// It will run and feed back solutions
type HasherSet struct {
	Instances   []*Hasher
	BlockHashes chan [32]byte
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
func NewHashers(Instances int, Nonce uint64, Lx *pow.LxrPow) *HasherSet {
	h := new(HasherSet)
	h.Nonce = Nonce
	h.Lx = Lx
	h.BlockHashes = make(chan [32]byte, 10)
	h.Solutions = make(chan PoWSolution, 10)
	h.Control = make(chan string, 10)

	for i := 0; i < Instances; i++ {
		n := h.Nonce ^ uint64(i)
		n = n<<19 ^ n>>11

		instance := NewHasher(i, n, Lx)
		instance.Solutions = h.Solutions // override Solutions channel

		h.Instances = append(h.Instances, instance) // Collect all our instances

		instance.Start()
	}

	return h
}

// SetSolutions
// Direct solutions to the given solutions channel
func (h *HasherSet) SetSolutions(solutions chan PoWSolution) {
	h.Solutions = solutions
	for _, hasher := range h.Instances {
		hasher.Solutions = solutions
	}
}

func (h *HasherSet) Stop() {
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

func (h *HasherSet) Start() {
	if h.Started {
		return
	}
	h.Started = true

	go func() {
		for {
			select {
			case hash := <-h.BlockHashes:
				for _, i := range h.Instances {
					for !i.Started && len(i.BlockHashes) > 0 {
						<-i.BlockHashes
					}
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
