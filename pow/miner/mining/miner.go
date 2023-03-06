// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package mining

import (
	"github.com/pegnet/LXRPow/pow"
	"github.com/pegnet/LXRPow/pow/miner"
)

type PoWSolution struct {
	nonce   uint64
	pow     uint64
	HashCnt uint64
}

var lx pow.LxrPow

// All that is needed to create a Miner instance.  Once it is created,
// It will run and feed back solutions
type Miner struct {
	CurrentHash []byte
	Nonce       uint64
	BlockHashes chan []byte
	Solutions   chan PoWSolution
	Control     chan string
	Best        uint64
}

func CreatePowInstance(Loops, Bits,Passes int) {
	lx.Init(Loops,Bits,Passes)
}

func CreateMiner(Nonce uint64) *Miner {
	m := new(Miner)
	m.Nonce = Nonce
	m.BlockHashes = make(chan []byte, 100)
	m.Solutions = make(chan PoWSolution, 100)
	m.Control = make(chan string, 1)
	return m
}

func (m *Miner) Stop() {
	m.Control <- "stop"
}

func (m *Miner) Start() {
	var best, hashCnt uint64
	var hash []byte
	for {
		if len(m.BlockHashes) > 0 {
			hash = <-m.BlockHashes
		}
		hashCnt++
		m.Nonce ^= m.Nonce<<17 ^ m.Nonce>>9 ^ hashCnt // diff nonce for each instance
		nPow := miner.lx(hash, m.Nonce)
		if nPow > best {
			m.Solutions <- PoWSolution{m.Nonce, nPow, hashCnt}
			best = nPow

		}
	}

}
