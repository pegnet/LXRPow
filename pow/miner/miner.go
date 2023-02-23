// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pegnet/LXRPow/pow"
)

const LxrPoW_time int64 = 120

func main() {

	pInstances := flag.Int("instances", 1, "Number of instances of the hash miners")
	pLoop := flag.Int64("loop", 50, "Number of loops accessing ByteMap (more is slower)")
	pBits := flag.Int64("bits", 30, "Number of bits addressing the ByteMap (more is bigger)")
	pData := flag.String("data", "A Test Phrase", "A phrase that will be mined")
	pPhrase := flag.String("phrase", "", "private phrase hashed to ensure unique nonces for the miner")
	pReportBar := flag.Int64("powreport", 0x0300<<48, "All Pow greater than minPow will be reported")
	ReportingBar := uint64(*pReportBar)
	Phrase := *pPhrase
	Bits := *pBits
	Data := *pData
	Loop := *pLoop
	Instances := *pInstances

	lx := new(pow.LxrPow).Init(Loop, Bits, 6)
	oprHash := sha256.Sum256([]byte(Data))
	phraseHash := sha256.Sum256([]byte(Phrase))
	nonce := binary.BigEndian.Uint64(phraseHash[:]) + uint64(time.Now().UnixNano())

	type result struct{ nonce, pow uint64 }
	results := make(chan result, 10)

	var best result
	var hashCnt int64
	start := time.Now()

	for i := 0; i < Instances; i++ {
		nonce = nonce<<9 ^ nonce>>3 ^ uint64(i)
		go func(a [32]byte, instance uint64) {
			var best uint64
			for {
				instance = instance<<17 ^ instance>>9 ^ uint64(hashCnt) ^ uint64(instance) // diff nonce for each instance
				h := atomic.AddInt64(&hashCnt, 1)
				_, nPow := lx.LxrPoW(a[:], uint64(h))
				if nPow > best || nPow > ReportingBar {
					results <- result{instance, nPow}
					if nPow > best {
						best = nPow
					}
				}
			}
		}(oprHash, uint64(i))
	}

	// pull the results, and print the best hashes
	for i := 0; true; {
		r := <-results
		if r.pow > best.pow {
			i++
			best = r
			current := time.Since(start)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			fmt.Printf("  %3d time: %10s TH: %10d H/s: %12.5f Pow: %016x Hash: %64x Nonce: %016x\n",
				i, fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
				hashCnt, rate, r.pow, oprHash, r.nonce)
		} else if r.pow > best.pow || r.pow > ReportingBar {
			current := time.Since(start)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			fmt.Printf("      time: %10s TH: %10d H/s: %12.5f Pow: %016x\n",
				fmt.Sprintf("%3d:%02d:%02d", int(current.Hours()), int(current.Minutes())%60, int(current.Seconds())%60),
				hashCnt, rate, r.pow)
		}
	}
}
