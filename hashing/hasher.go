// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"fmt"

	"github.com/pegnet/LXRPow/pow"
)

type PoWSolution struct {
	TokenURL string
	Hash     []byte
	Nonce    uint64
	Pow      uint64
	HashCnt  uint64
}

// All that is needed to create a Hasher instance.  Once it is created,
// It will run and feed back solutions
type Hasher struct {
	CurrentHash []byte
	Nonce       uint64
	BlockHashes chan []byte
	Solutions   chan PoWSolution
	Control     chan string
	Best        uint64
	Started     bool
	HashCnt     uint64
	LX          *pow.LxrPow
}

func NewHasher(nonce uint64, lx *pow.LxrPow) *Hasher {
	m := new(Hasher)
	m.Nonce = nonce
	m.LX = lx

	// Inputs to the Hasher
	m.Control = make(chan string, 1)     // The Hasher stops when told on this channel
	m.BlockHashes = make(chan []byte, 1) // Hashes to Hash are read from this channel

	// Outputs from the Hasher (can be overwritten and thus shared across Hashers)
	m.Solutions = make(chan PoWSolution, 1) // Solutions are written to this channel

	return m
}

func (m *Hasher) Stop() {
	if !m.Started {
		return
	}
	m.Control <- "stop"
	m.Started = false
	fmt.Println("Stopping Instance")
}

func (m *Hasher) Start() {
	if m.Started {
		return
	}

	m.Started = true
	go func() {
		var best uint64
		var hash []byte

		hash = <-m.BlockHashes
		for {
			select {
			case hash = <-m.BlockHashes:
				m.CurrentHash = hash
				best = 0
			case cmd := <-m.Control:
				if cmd == "stop" {
					return
				}
			default:
			}
			m.HashCnt++
			m.Nonce ^= m.Nonce<<17 ^ m.Nonce>>9 ^ m.HashCnt // diff nonce for each instance
			nPow := m.LX.LxrPoW(hash, m.Nonce)
			if nPow > best {
				m.Solutions <- PoWSolution{"", hash, m.Nonce, nPow, m.HashCnt}
				best = nPow

			}
		}
	}()
}
