// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"time"

	"github.com/pegnet/LXRPow/pow"
)

type PoWSolution struct {
	Block    uint64    // Block Number
	TokenURL string    // Payout token account
	Instance int16     // ID of the hasher producing this block
	time     time.Time // Timestamp when block was produced
	DNHash   [32]byte  // The hash of the block
	Nonce    uint64    // Nonce that is the solution
	Pow      uint64    // Self-reported Difficulty
	HashCnt  uint64    // Count of hashes performed so far
}

// All that is needed to create a Hasher instance.  Once it is created,
// It will run and feed back solutions
type Hasher struct {
	Instance    int
	CurrentHash [32]byte
	Nonce       uint64
	BlockHashes chan [32]byte
	Solutions   chan PoWSolution
	Control     chan string
	Best        uint64
	Started     bool
	HashCnt     uint64
	LX          *pow.LxrPow
}

func NewHasher(instance int, nonce uint64, lx *pow.LxrPow) *Hasher {
	m := new(Hasher)
	m.Instance = instance
	m.Nonce = nonce
	m.LX = lx

	// Inputs to the Hasher
	m.Control = make(chan string, 1)       // The Hasher stops when told on this channel
	m.BlockHashes = make(chan [32]byte, 1) // Hashes to Hash are read from this channel

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
}

func (m *Hasher) Start() {
	if m.Started {
		return
	}

	m.Started = true
	go func() {
		var best uint64
		var hash [32]byte

		hash = <-m.BlockHashes
		for {
			select {
			case hash = <-m.BlockHashes:
				m.CurrentHash = hash
				best = 0
				continue // Read Hashes until the channel is empty
			case cmd := <-m.Control:
				if cmd == "stop" {
					return
				}
			default:
			}
			m.HashCnt++
			m.Nonce ^= m.Nonce<<17 ^ m.Nonce>>9 ^ m.HashCnt // diff nonce for each instance
			nPow := m.LX.LxrPoW(hash[:], m.Nonce)
			if nPow > best || nPow > 0xFFFF000000000000 {
				m.Solutions <- PoWSolution{
					m.HashCnt, "", int16(m.Instance), time.Now(), hash, m.Nonce, nPow, m.HashCnt,
				}
				best = nPow

			}
		}
	}()
}
