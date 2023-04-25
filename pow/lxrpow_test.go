// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// Using this test to check the performance of the LxrPow
func Test_LxrPoWSpeed(t *testing.T) {
	doLxPow := true

	lx := new(LxrPow).Init(16, 30, 6)

	type result struct{ nonce, pow uint64 }
	results := make(chan result, 100)

	oprHash := sha256.Sum256([]byte("This is a test"))
	var hashCnt int64
	nonce := uint64(time.Now().UnixNano())

	goCnt := 1
	for i := 0; i < goCnt; i++ {

		go func(instance int, a [32]byte) {
			var best uint64
			for {
				nonce = nonce<<17 ^ nonce>>9 ^ uint64(hashCnt) ^ uint64(instance) // diff nonce for each instance
				h := atomic.AddInt64(&hashCnt, 1)
				var nPow uint64
				if doLxPow {
					nPow = lx.LxrPoW(a[:], uint64(h))
				} else {
					var n [8]byte
					binary.BigEndian.PutUint64(n[:], nPow)
					x := append(a[:], n[:]...)
					shaHash := sha256.Sum256(x)
					nPow = binary.BigEndian.Uint64(shaHash[:])
				}
				if nPow > best {
					results <- result{nonce, nPow}
					if nPow > best {
						best = nPow
					}
				}
			}
		}(i, oprHash)
	}
	start := time.Now()
	time.Sleep(5 * time.Second)
	fmt.Println("Starting")
	var best result
	// pull the results, and print the best hashes
	for {
		var r result
		select {
		case r = <-results:
			if r.pow > best.pow {
				best = r
			}
		default:
			current := time.Since(start)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			fmt.Printf("      time: %10s TH: %10d H/s: %12.5f Pow: %016x\n",
				fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
				hashCnt, rate, best.pow)
			time.Sleep(5 * time.Second)
		}

	}
}
