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

	humanize "github.com/dustin/go-humanize"
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

// Using this test to check the performance of the LxrPow
func Test_Sha256PoWSpeed(t *testing.T) {
	oprHash := sha256.Sum256([]byte("This is a test"))
	var hashCnt int64
	start := time.Now()
	fmt.Println("Starting")
	// pull the results, and print the best hashes
	for {
		for i := 0; i < 100000; i++ {
			hashCnt++
			oprHash = sha256.Sum256(oprHash[:])
		}
		current := time.Since(start)
		rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
		fmt.Printf("      time: %10s TH: %10d H/s: %12.5f \n",
			fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
			hashCnt, rate)
	}
}

// Using this test to check the performance of the LxrPow
func Test_StartValue(t *testing.T) {
	lx := new(LxrPow)
	_ = lx
	Hash := sha256.Sum256([]byte{1, 2, 4, 5, 7, 3, 1})

	for i := 0; i < 10; i++ {
		Hash = sha256.Sum256(Hash[:])	
		var last, d, diff, under, cnt uint64
		for i := 0; i < 100000; i++ {
			_, state := lx.mix(Hash[:],uint64(i))
			//state := binary.BigEndian.Uint64(Hash[:])
			//Hash = sha256.Sum256(Hash[:])
			
			
			cnt++
			if last > state {
				d += last - state
			} else {
				d += state - last
			}
			last = state
			d = d % (1024 * 1024 * 1024)
			if d < 1024 {
				under++
			}
			diff += d	
		}
		fmt.Printf("%d %x\n",i,Hash[:10])
		fmt.Printf("average jump: %s\n", humanize.Comma(int64(diff/cnt%(1024*1024*1024))))
		fmt.Printf("Under 1k: %d \n", under)
	}
}
