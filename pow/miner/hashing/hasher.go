// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"github.com/pegnet/LXRPow/pow"
)

type PoWSolution struct {
	Hash    []byte
	Nonce   uint64
	Pow     uint64
	HashCnt uint64
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
	lx          *pow.LxrPow
}

func NewMiner(nonce uint64, lx *pow.LxrPow) *Hasher {
	m := new(Hasher)
	m.Nonce = nonce
	m.BlockHashes = make(chan []byte, 100)
	m.Solutions = make(chan PoWSolution, 100)
	m.Control = make(chan string, 5)
	m.lx = lx
	return m
}

func CreatePowInstance(Loops, Bits, Passes int) *pow.LxrPow {
	lx := new(pow.LxrPow)
	lx.Init(Loops, Bits, Passes)
	return lx
}

func (m *Hasher) Stop() {
	m.Control <- "stop"
	m.Started = false
}

func (m *Hasher) Start() {
	if m.Started {
		return
	}
	go func() {
		var best, hashCnt uint64
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

			hashCnt++
			m.Nonce ^= m.Nonce<<17 ^ m.Nonce>>9 ^ hashCnt // diff nonce for each instance
			nPow := m.lx.LxrPoW(hash, m.Nonce)
			if nPow > best {
				m.Solutions <- PoWSolution{hash, m.Nonce, nPow, hashCnt}
				best = nPow

			}
		}
	}()
}
