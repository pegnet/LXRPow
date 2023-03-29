package mine

import (
	"fmt"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/cfg"
	"github.com/pegnet/LXRPow/hashing"
)

type Miner struct {
	Cfg       *cfg.Config
	Hashers   *hashing.HasherSet
	Started   bool
	Solutions chan hashing.PoWSolution
	Control   chan string
	API       *accumulate.APIStub
}

func (m *Miner) Init(cfg *cfg.Config) {
	m.Cfg = cfg

	// Input to the Miner
	m.Control = make(chan string, 1) // The Hasher stops when told on this channel
	// Outputs of hashers to the Miners
	m.Solutions = make(chan hashing.PoWSolution, 10) // Solutions are written to this channel

	m.API = accumulate.NewAPIStub(cfg)

	m.Hashers = hashing.NewHashers(cfg.Instances, cfg.Seed, cfg.LX) // Allocate the Hashers
	m.Hashers.SetSolutions(m.Solutions)                             // Override their Solutions channel
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
	m.Started = true

	var best *hashing.PoWSolution
	var height uint64
	var hash []byte
	HashCounts := make(map[int]uint64)

	for {
		select {
		case solution := <-m.Solutions: // New solutions have to be graded

			HashCounts[int(solution.Instance)] = solution.HashCnt // Collect all the hashing counts from hashers

			if best == nil || solution.Difficulty > best.Difficulty { // If the best so far on the block
				solution.TokenURL = m.Cfg.TokenURL // Confi
				best = &solution
				m.API.WriteSolution(solution, HashCounts)
			}
			continue
		case cmd := <-m.Control:
			if cmd == "stop" {
				m.Stop()
				return
			}
		default:

		}
		height, hash = m.API.GetNextPow(height) //   Reads and processes all hashing records until the latest
		if hash != nil {                        //   Not nil indicates a new block
			m.Hashers.BlockHashes <- hash   // Send the hash to the hashers
			if !m.Hashers.Started { // If hashers are not started, do so after we have a hash set to them.
				m.Hashers.Start()
			}
			best = new(hashing.PoWSolution) // Create a nil PoWSolution, because this is a new block
		} else {
			time.Sleep(time.Second / 1000) //        Sleep for 1/10 a second
		}
	}

}
