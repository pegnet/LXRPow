package main

import (
	"fmt"

	"github.com/pegnet/LXRPow/pow"
	"github.com/pegnet/LXRPow/pow/miner/hashing"
)

type Miner struct {
	Cfg       *Config
	Index     uint64
	Hashers   *hashing.Hasher
	lx        pow.LxrPow
	Started   bool
	Solutions chan hashing.PoWSolution
	Control   chan string
}

func (m *Miner) Init(cfg *Config, index uint64) {
	m.Cfg = cfg
	m.Index = index
	// Input to the Miner
	m.Control = make(chan string, 1)                  // The Hasher stops when told on this channel
	m.Solutions = make(chan hashing.PoWSolution, 1)   // Solutions are written to this channel
	lx := pow.NewLxrPow(cfg.Loop, cfg.Bits, 6)        // Allocate the LxrPow
	m.Hashers = hashing.NewHasher(cfg.Seed^index, lx) // Allocate the Hashers
	m.Hashers.Solutions = m.Solutions                 // Override their Solutions channel
	m.Hashers.Start()                                 // Start the Hashers
}

func (m *Miner) Stop() {
	fmt.Printf("Miner %2d has stopped", m.Index)
	m.Hashers.Stop()

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
	m.Started = true
	var height uint64
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
			//height, hash := GetNextPow(Cfg,height)
		}
	}()
}

/*

// pull the results, and print the best hashes
hashTime := time.Now()
var BlockCnt, TotalDifficulty, WindowBlockCnt, WindowDifficulty float64
var WindowTime time.Duration
for i := 0; true; {
	if Timed && time.Since(hashTime) > time.Second*time.Duration(BlockTime) {
		BlockCnt++
		WindowBlockCnt++
		TotalDifficulty += float64(best.pow)
		WindowDifficulty += float64(best.pow)
		AvgDifficulty := TotalDifficulty / BlockCnt
		AvgWindowDifficulty := WindowDifficulty / WindowBlockCnt
		current := time.Since(start)
		sinceLast := time.Since(hashTime)
		WindowTime += sinceLast
		rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
		coreRate := rate / float64(Instances)
		average := current / time.Duration(BlockCnt)
		fmt.Printf("  %3s time: %10s Th: %10d Th/s: %12.5f Ch/s: %12.5f Pow: %016x Hash: %16x Nonce: %016x\n",
			"",
			fmt.Sprintf("%3d:%02d:%02d", int(sinceLast.Hours()), int(sinceLast.Minutes())%60, int(sinceLast.Seconds())%60),
			hashCnt, rate, coreRate, best.pow, oprHash[:8], best.nonce)
		fmt.Printf("Total Block Count: %d Average Block Time: %3d:%02d:%02d Average Difficulty: %016x Average Window Difficulty: %016x \n\n",
			int64(BlockCnt),
			int(average.Hours()), int(average.Minutes())%60, int(average.Seconds())%60,
			uint64(AvgDifficulty),
			uint64(AvgWindowDifficulty),
		)
		post()
		best = result{}
		hashTime = time.Now()
		fmt.Println()
		i = 0
		if uint64(BlockCnt)%uint64(DiffWindow) == 0 {
			WindowTime = WindowTime / time.Duration(WindowBlockCnt)
			avgSec := int(WindowTime.Seconds()) % 60
			avgMs := int(WindowTime.Seconds() * 1000)
			fmt.Printf("Window of %d blocks has an average Difficulty: %016x Average Blocktime %d:%02d:%02d.%03d\n\n",
				uint64(WindowBlockCnt),
				uint64(AvgWindowDifficulty),
				int(WindowTime.Hours()), int(WindowTime.Minutes()), avgSec, avgMs,
			)
			WindowBlockCnt = 0
			WindowDifficulty = 0
			WindowTime = 0

		}
	}
	time.Sleep(time.Millisecond * 200)
	for len(results) > 0 {
		r := <-results
		if r.pow > best.pow {
			i++
			best = r
			current := time.Since(start)
			sinceLast := time.Since(hashTime)
			WindowTime += sinceLast
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			coreRate := rate / float64(Instances)
			fmt.Printf("  %3d time: %10s Th: %10d Th/s: %12.5f Ch/s: %12.5f Pow: %016x Hash: %16x Nonce: %016x\n",
				i,
				fmt.Sprintf("%3d:%02d:%02d", int(sinceLast.Hours()), int(sinceLast.Minutes())%60, int(sinceLast.Seconds())%60),
				hashCnt, rate, coreRate, r.pow, oprHash[:8], r.nonce)
			if !Timed && r.pow > Difficulty {
				BlockCnt++
				WindowBlockCnt++
				TotalDifficulty += float64(best.pow)
				AvgDifficulty := TotalDifficulty / BlockCnt
				WindowDifficulty += float64(best.pow)
				AvgWindowDifficulty := WindowDifficulty / WindowBlockCnt
				average := current / time.Duration(BlockCnt)
				avgSec := int(average.Seconds())
				avgMs := int((average.Seconds() - float64(avgSec)) * 1000)
				fmt.Printf("Total Block Count: %d Average Block Time: %3d:%02d:%02d.%03d Average Difficulty: %016x Average Window Difficulty: %016x \n\n",
					int64(BlockCnt),
					int(average.Hours()), int(average.Minutes())%60, avgSec, avgMs,
					uint64(AvgDifficulty),
					uint64(AvgWindowDifficulty),
				)
				post()
				hashTime = time.Now()
				best = result{}
				i = 0
				if uint64(BlockCnt)%uint64(DiffWindow) == 0 {
					WindowTime = WindowTime / time.Duration(WindowBlockCnt)
					TimeError := WindowTime - time.Duration(BlockTime)*time.Second
					PercentError := TimeError.Seconds() / float64(BlockTime)

					oldDiff := Difficulty
					// Adjust the Difficulty by %50 percent of the error
					Adjust := 1 + PercentError/2
					Difficulty ^= 0xFFFFFFFFFFFFFFFF
					Difficulty = uint64(float64(Difficulty) * Adjust)
					Difficulty ^= 0xFFFFFFFFFFFFFFFF

					WTSeconds := int(WindowTime.Seconds())
					WTMilli := int(WindowTime.Seconds()*1000) - WTSeconds
					fmt.Printf("S Window %d blocks \nS  average Difficulty: %016x\nS  Average Blocktime %d:%02d:%02d.%03d\n"+
						"S  Old Difficulty: %016x\nS  New Difficulty: %016x\nS  Percent Error %3.4f\n"+
						"S  Adjust %3.4f\n\n",
						uint64(WindowBlockCnt),
						uint64(AvgWindowDifficulty),
						int(WindowTime.Hours()), int(WindowTime.Minutes()), WTSeconds, WTMilli,
						oldDiff, Difficulty, PercentError, Adjust,
					)

					WindowBlockCnt = 0
					WindowDifficulty = 0
					WindowTime = 0
				}
			}
		}
	}
}
}
*/
