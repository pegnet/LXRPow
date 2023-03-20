package miner

import (
	"fmt"

	"github.com/pegnet/LXRPow/hashing"
	"github.com/pegnet/LXRPow/pow"
)

type Miner struct {
	Cfg       *Config
	Hashers   *hashing.Hasher
	Started   bool
	Solutions chan hashing.PoWSolution
	Control   chan string
}

func (m *Miner) Init(cfg *Config) {
	m.Cfg = cfg

	// Input to the Miner
	m.Control = make(chan string, 1) // The Hasher stops when told on this channel
	// Outputs of hashers to the Miners
	m.Solutions = make(chan hashing.PoWSolution, 1)   // Solutions are written to this channel
	m.Cfg.LX = pow.NewLxrPow(cfg.Loop, cfg.Bits, 6)   // Allocate the LxrPow
	m.Hashers = hashing.NewHasher(cfg.Seed, m.Cfg.LX) // Allocate the Hashers
	m.Hashers.Solutions = m.Solutions                 // Override their Solutions channel
	m.Hashers.Start()                                 // Start the Hashers
}

func (m *Miner) Stop() {
	fmt.Printf("Miner %2d has stopped", m.Cfg.Index)
	//	m.Hashers.Stop()

}

// Start
// The job of the miner is to find the best hash it can from its hashers
// When hashers find a solution, those are fed to WriteSolution.  WriteSolution
// implements strategies for optimal submission of mining records
func (m *Miner) Start() {
	if m.Started {
		return
	}
	var best *hashing.PoWSolution
	_ = best
	m.Started = true
	var height uint64
	var hash []byte
	go func() {
		for {
			select {
			case solution := <-m.Solutions:
				if best == nil || solution.Pow > best.Pow {
					solution.TokenURL = m.Cfg.TokenURL
					best = &solution
					WriteSolution(height, &solution)
				}
			case cmd := <-m.Control:
				if cmd == "stop" {
					m.Stop()
					return
				}
			default:
			}
			height, hash = GetNextPow(m.Cfg, height)
			if hash != nil {
				m.Hashers.BlockHashes <- hash
			}
		}
	}()
}
