// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"crypto/sha256"
	"fmt"
	//"net/url"
	"sync"
	"time"

	"github.com/pegnet/LXRPow/pow/miner/hashing"
)

var start = time.Now()
var firstBlock = 10000
var MiningRecords []*hashing.PoWSolution
var rw, h sync.Mutex

// WriteSolution
// When the Miner finds a solution
// We should write the solution to Accumulate!
func WriteSolution(height uint64, solution *hashing.PoWSolution) {
	fmt.Printf("Token URL              %x\n", solution.TokenURL)
	fmt.Printf("Directory Network Hash %x\n", solution.Hash)
	fmt.Printf("Nonce                  %x\n", solution.Nonce)
	fmt.Printf("Self Reported PoW      %016x\n", solution.Pow)
	fmt.Printf("Total Hashes           %16x\n", solution.Hash)
	fmt.Println()
	rw.Lock()
	defer rw.Unlock()
	MiningRecords = append(MiningRecords, solution)
}

var cDNHash = []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// GetDNHeight
// Returns
//
//	DNHeight - Height of the Directory Network
//	DNHash   - Hash of the Directory Network at DNHeight
func GetDNHeight() (DNHeight uint64, DNHash []byte) {
	rw.Lock()
	defer rw.Unlock()
	DNHeight = uint64(time.Since(start).Nanoseconds()/1e9) + uint64(firstBlock)
	h := sha256.Sum256(cDNHash)
	copy(cDNHash, h[:])
	return DNHeight, h[:]
}

// GetUnread
// Read the submitted records to the submission chain since we last looked
// Returns the same height and an empty list if no updates have been made
func GetUnread(height uint64) (newHeight uint64, entries []*hashing.PoWSolution) {
	rw.Lock()
	defer rw.Unlock()
	if height == uint64(len(MiningRecords)) {
		return height, entries
	}
	newHeight = uint64(len(MiningRecords))
	entries = append(entries, MiningRecords[height:]...)
	return newHeight, entries
}


// Detect when blocks are complete. The height informs this code as to the state
// of the miner, i.e. has the state changed since the last call from the miner,
// so it needs a new hash.
func GetNextPow(Cfg *Config,height uint64) (newHeight uint64, hash []byte) {
	var entries []*hashing.PoWSolution
	newHeight, entries = GetUnread(height)
	
	_=entries


	return newHeight,hash
}
