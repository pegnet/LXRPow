package mine

import (
	"fmt"
	"time"

	"github.com/pegnet/LXRPow/cfg"
	"github.com/pegnet/LXRPow/hashing"
)

type Miner struct {
	Cfg       *cfg.Config
	Hashers   *hashing.Hashers
	Started   bool
	Solutions chan hashing.PoWSolution
	Control   chan string
	API       *APIStub
}

func (m *Miner) Init(cfg *cfg.Config) {
	m.Cfg = cfg

	// Input to the Miner
	m.Control = make(chan string, 1) // The Hasher stops when told on this channel
	// Outputs of hashers to the Miners
	m.Solutions = make(chan hashing.PoWSolution, 1) // Solutions are written to this channel

	m.API = NewAPIStub(cfg)

	m.Hashers = hashing.NewHashers(cfg.Instances, cfg.Seed, cfg.LX) // Allocate the Hashers
	m.Hashers.SetSolutions(m.Solutions)                             // Override their Solutions channel
	m.Hashers.Start()                                               // Start the Hashers
}

func (m *Miner) Stop() {
	fmt.Printf("Miner %2d has stopped", m.Cfg.Index)
	m.Hashers.Stop()
}

// Run
// The job of the miner is to find the best hash it can from its hashers
// When hashers find a solution, those are fed to WriteSolution.  WriteSolution
// implements strategies for optimal submission of mining records
func (m *Miner) Run() {
	if m.Started {
		return
	}
	var best *hashing.PoWSolution
	_ = best
	m.Started = true
	var height uint64
	var hash []byte

	m.Hashers.BlockHashes <- m.API.DNHash // Send the first DN Hash to mine

	for {
		select {
		case solution := <-m.Solutions:
			if best == nil || solution.Pow > best.Pow {
				solution.TokenURL = m.Cfg.TokenURL
				best = &solution
				m.API.WriteSolution(solution)
			}
		case cmd := <-m.Control:
			if cmd == "stop" {
				m.Stop()
				return
			}
		default:
		}
		height, hash = m.API.GetNextPow(height)
		if hash != nil {
			m.Hashers.BlockHashes <- hash
			fmt.Println()
			best = new(hashing.PoWSolution)
		} else {
			time.Sleep(time.Second / 1000) // Sleep for 1/10 a second
		}
	}

}
