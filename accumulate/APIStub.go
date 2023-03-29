// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package accumulate

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	//"net/url"
	"sync"
	"time"

	"github.com/pegnet/LXRPow/cfg"
	"github.com/pegnet/LXRPow/hashing"
)

type APIStub struct {
	Cfg                *cfg.Config           // Configuration Parameters for the miner
	Start              time.Time             // Start time of the current block
	StartWindow        time.Time             // Start the window (adjusting difficulty)
	windowAvgBlockTime float64               // As we run, should approach the target block time
	BlockCnt           int                   // Count of blocks mined
	WindowCnt          int                   // Count of blocks in window
	Hash               []byte                // Hash of the highest DN Block being mined
	MiningRecords      []hashing.PoWSolution // Record of all Mining Records for this block
	rw                 sync.Mutex            // Because we may have many threads
	DNHash             []byte                // DN hash being mined
	BestPow            *hashing.PoWSolution  // BestPow so far
	Height             int                   // Height for writes; we track the best solution
	Difficulty         uint64                // Represents the Difficulty in the settings on Accumulate
}

// GetDifficulty
func (a *APIStub) GetDifficulty() uint64 {
	if a.Difficulty == 0 {
		a.Difficulty = a.Cfg.Difficulty
	}
	return a.Difficulty
}

// SetDifficulty
func (a *APIStub) SetDifficulty(difficulty uint64) {
	a.Difficulty = difficulty
}

// GetBlock
// Returns the slice of all the records found for DNHash.  Returns true
// if a complete block is found, returns false if the block is not complete (found the
// difficulty target)
func (a *APIStub) GetBlock(DNHash []byte, difficulty uint64) (block []hashing.PoWSolution, complete bool) {
	for _, mr := range a.MiningRecords {
		if bytes.Equal(a.DNHash, DNHash) {
			mrDifficulty := a.Cfg.LX.LxrPoW(DNHash, mr.Nonce)
			if mrDifficulty != mr.Difficulty {
				continue //                      If self reported difficulty isn't correct, ignore entry
			} //                                  TODO: Miner loses points for bad entry
			block = append(block, mr)
			if mrDifficulty >= difficulty {
				return block, true
			}
		}
	}
	return block, false
}

// NewAPIStub
// store away our parameters, and initialize the first DNHash
func NewAPIStub(cfg *cfg.Config) *APIStub {
	a := new(APIStub)
	a.Cfg = cfg
	h := sha256.Sum256([]byte{1})
	a.DNHash = h[:]
	a.BlockCnt++
	a.Start = time.Now()
	a.StartWindow = time.Now()
	a.PrintHeader()
	return a
}

// NextDNHash
// When a block is complete, we mine the next DN Block.
func (a *APIStub) NextDNHash() []byte {
	var h [32]byte
	if a.Hash == nil {
		h = sha256.Sum256([]byte{1})
		a.Hash = make([]byte, 32)
	} else {
		h = sha256.Sum256(a.Hash)
	}
	copy(a.Hash, h[:])
	return h[:]
}

// WriteSolution
// When the Miner finds a solution
// We should write the solution to Accumulate!
func (a *APIStub) WriteSolution(solution hashing.PoWSolution, HashCounts map[int]uint64) {

	var totalHashes uint64
	for _, v := range HashCounts {
		totalHashes += v
	}

	url := solution.TokenURL // Truncate URL to the field
	if len(solution.TokenURL) > 25 {
		url = url[:25]
	}

	elapsed := time.Since(a.Start)
	seconds := elapsed.Seconds()
	hps := float64(totalHashes) / seconds

	fmt.Printf("%6d ", solution.Instance)     // Note that the fields here
	fmt.Printf("%5d ", a.BlockCnt)            // are coded in PrintHeader to center
	fmt.Printf("%-25s ", url)                 // So make sure they all match up.  The URL
	fmt.Printf("%010x ", solution.DNHash[:5]) // length above, here, and in
	fmt.Printf("%016x ", solution.Nonce)      // PrintHeader()
	fmt.Printf("%016x ", solution.Difficulty) // The hasher's solution
	fmt.Printf("%016x ", a.GetDifficulty())   // Current difficulty
	fmt.Printf("%16d ", totalHashes)          // Estimate of all the hashes done by the hashers
	fmt.Printf("%16.2f\n", hps)               // Hashes per second

	a.rw.Lock()
	defer a.rw.Unlock()
	a.MiningRecords = append(a.MiningRecords, solution)
}

// PrintHeader
func (a *APIStub) PrintHeader() {

	Overall := float64(time.Since(a.Start).Seconds()) / float64(a.BlockCnt)
	fmt.Printf("Overall Window Block Time: %8.3f\n", Overall)

	fmt.Printf("Last Window Block Time:    %8.3f\n", a.windowAvgBlockTime)
	fmt.Printf("%6s %5s %-25s %-10s %-16s %-16s %-16s %-16s %-16s\n",
		"hasher", "Blk#", "       URL", "   Hash", "     Nonce",
		"      pow", "   difficulty", "   total hashes", "       hps",
	)
}

// GetDNHeight
// Returns
//
//	DNHeight - Height of the Directory Network
//	DNHash   - Hash of the Directory Network at DNHeight
func (a *APIStub) GetDNHeight() (DNHeight uint64, DNHash []byte) {
	a.rw.Lock()
	defer a.rw.Unlock()
	DNHeight = uint64(time.Since(a.Start).Nanoseconds() / 1e9)
	h := sha256.Sum256(a.DNHash)
	copy(a.DNHash, h[:])
	return DNHeight, h[:]
}

// GetUnread
// Read the submitted records to the submission chain since we last looked
// Returns the same height and an empty list if no updates have been made
func (a *APIStub) GetUnread(height uint64) (entries []hashing.PoWSolution) {
	a.rw.Lock()
	defer a.rw.Unlock() // Lock and unlock around accessing a.MiningRecords

	if height == uint64(len(a.MiningRecords)) { // If nothing has been added since the last touch
		return entries //                  return given height and a null set
	}

	entries = append(entries, a.MiningRecords[height:]...) // Copy over the new stuff
	return entries                                         // Return that stuff
}

// Detect when blocks are complete. The height informs this code as to the state
// of the miner, i.e. has the state changed since the last call from the miner,
// so it needs a new DNHash.
func (a *APIStub) GetNextPow(height uint64) (newHeight uint64, DNHash []byte) {

	entries := a.GetUnread(height)

	// Look for the current hash
	for _, e := range entries {
		if e.Difficulty >= a.GetDifficulty() {
			a.BlockCnt++
			a.WindowCnt++
			a.BestPow = new(hashing.PoWSolution)
			a.DNHash = a.NextDNHash()
			DNHash = a.DNHash
			fmt.Println()
			a.PrintHeader()
			if a.WindowCnt >= a.Cfg.Window {
				// Avg window block time is the time in the window / the window block cnt
				a.windowAvgBlockTime = time.Since(a.StartWindow).Seconds()
				a.windowAvgBlockTime /= float64(a.WindowCnt)
				a.StartWindow = time.Now()                               // Reset the Window start time
				delta := float64(a.Cfg.BlockTime) - a.windowAvgBlockTime // Delta from expected avg block time
				percentDelta := delta / float64(a.Cfg.BlockTime)
				nDifficulty := a.GetDifficulty()
				// In order to scale the difficulty up, the easiest thing to do is to
				// flip from a largest result to a smallest result by taking the 1's complement.
				// We then scale it up or down (depending on the error of the Delta).  Then
				// flip it again by taking the 1's complement of the adjusted value.
				nDifficulty = 0xFFFFFFFFFFFFFFFF ^ nDifficulty
				nDifficulty = uint64(float64(nDifficulty) * (1 - percentDelta/4)) // Make a 25% of the delta adjustment
				a.SetDifficulty(uint64(nDifficulty) ^ 0xFFFFFFFFFFFFFFFF)
				a.WindowCnt = 0
			}
		}
	}

	if height == 0 {
		DNHash = a.DNHash
	}

	return height + uint64(len(entries)), DNHash // Note DNHash is nil unless a new DNHash is found
}
