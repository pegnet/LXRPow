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
	pLoop := flag.Int("loop", 50, "Number of loops accessing ByteMap (more is slower)")
	pBits := flag.Int("bits", 30, "Number of bits addressing the ByteMap (more is bigger)")
	pData := flag.String("data", "A Test Phrase", "A phrase that will be mined")
	pPhrase := flag.String("phrase", "", "private phrase hashed to ensure unique nonces for the miner")
	pReportBar := flag.Int64("powreport", 0x0300<<48, "All Pow greater than minPow will be reported")

	flag.Parse()
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

	fmt.Printf("miner --instances=%d --loop=%d --bits=%d --data=\"%s\" --phrase=\"%s\" --powreport=0x%x \n",
		Instances,
		Loop,
		Bits,
		Data,
		Phrase,
		ReportingBar)

	type result struct{ nonce, pow uint64 }
	results := make(chan result, 10)
	var updates [500]chan []byte
	post := func() {
		phrase := append([]byte{}, phraseHash[:]...)
		for i := 0; i < Instances; i++ { // Allocate a channel per instance
			updates[i] = make(chan []byte, 10) // If an go routine can't empty as fast as the main loop
			updates[i] <- phrase
		}
	}

	var best result
	var hashCnt int64
	start := time.Now()

	for i := 0; i < Instances; i++ {
		nonce = nonce<<9 ^ nonce>>3 ^ uint64(i)
		go func(update chan []byte, instance uint64) {
			var best uint64
			a := <-update
			for {
				for len(update) > 0 {
					a = <-update
					best = 0
				}
				for i := 0; i < 500; i++ {
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
			}
		}(updates[i], nonce)
	}

	// pull the results, and print the best hashes
	hashTime := time.Now()
	for i := 0; true; i++ {
		r := <-results
		time.Sleep(time.Second / 2)
		if time.Since(hashTime) > time.Minute {
			post()
			best = result{}
		}
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
