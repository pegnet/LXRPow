// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package miner

import (
	"crypto/sha256"
	"fmt"

	//"net/url"
	"sync"
	"time"

	"github.com/pegnet/LXRPow/hashing"
)

type APIStub struct {
	Cfg           *Config               // Configuration Parameters for the miner
	start         time.Time             // Start time of the current block
	hash          []byte                // Hash of the highest DN Block being mined
	MiningRecords []hashing.PoWSolution // Record of all Mining Records for this block
	rw            sync.Mutex            // Because we may have many threads
	DNHash        []byte                // The current DN hash being mined
	BestPow       *hashing.PoWSolution  // BestPow so far
	Height        int                   // Height for writes; we track the best solution
}

// NewAPIStub
// store away our parameters, and initialize the first DNHash
func NewAPIStub(cfg *Config) *APIStub {
	a := new(APIStub)
	a.Cfg = cfg
	h := sha256.Sum256([]byte{1})
	a.DNHash = h[:]
	return a
}

func (a *APIStub) NextDNHash() []byte {
	var h [32]byte
	if a.hash == nil {
		h = sha256.Sum256([]byte{1})
		a.hash = make([]byte, 32)
	} else {
		h = sha256.Sum256(a.hash)
	}
	copy(a.hash, h[:])
	return h[:]
}

// WriteSolution
// When the Miner finds a solution
// We should write the solution to Accumulate!
func (a *APIStub) WriteSolution(solution hashing.PoWSolution) {
	url := solution.TokenURL
	if len(solution.TokenURL) > 10 {
		url = url[:10]
	}
	fmt.Printf("%10s ", url)
	fmt.Printf("%06x ", solution.Hash[:5])
	fmt.Printf("%016x ", solution.Nonce)
	fmt.Printf("%016x\n", solution.Pow)

	a.rw.Lock()
	defer a.rw.Unlock()
	a.MiningRecords = append(a.MiningRecords, solution)
}

// GetDNHeight
// Returns
//
//	DNHeight - Height of the Directory Network
//	DNHash   - Hash of the Directory Network at DNHeight
func (a *APIStub) GetDNHeight() (DNHeight uint64, DNHash []byte) {
	a.rw.Lock()
	defer a.rw.Unlock()
	DNHeight = uint64(time.Since(a.start).Nanoseconds() / 1e9)
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
		if e.Pow >= a.Cfg.Difficulty {
			a.BestPow = new(hashing.PoWSolution)
			a.DNHash = a.NextDNHash()
			DNHash = a.DNHash
		}
	}

	return height + uint64(len(entries)), DNHash // Note DNHash is nil unless a new DNHash is found
}
