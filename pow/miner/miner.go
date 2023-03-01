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
	pBlockTime := flag.Int("blocktime", 600, "Block Time in seconds (600 would be 10 minutes)")
	pTimed := flag.Bool("timed", false, "Blocks are timed, or end on pReportBar")
	flag.Parse()

	Instances := *pInstances
	Loop := *pLoop
	Bits := *pBits
	Data := *pData
	Phrase := *pPhrase
	ReportingBar := uint64(*pReportBar)
	BlockTime := *pBlockTime
	Timed := *pTimed

	lx := new(pow.LxrPow).Init(Loop, Bits, 6)
	oprHash := sha256.Sum256([]byte(Data))
	phraseHash := sha256.Sum256([]byte(Phrase))
	nonce := binary.BigEndian.Uint64(phraseHash[:]) + uint64(time.Now().UnixNano())

	fmt.Printf("\nminer --instances=%d --loop=%d --bits=%d --data=\"%s\""+
		" --phrase=\"%s\" --powreport=0x%x --blocktime=%d timed=%v\n\n",
		Instances,
		Loop,
		Bits,
		Data,
		Phrase,
		ReportingBar,
		BlockTime,
		Timed,
	)

	type result struct{ nonce, pow uint64 }
	results := make(chan result, 1000)

	var updates [500]chan []byte

	post := func() {
		for i := 0; i < Instances; i++ { // Allocate a channel per instance
			if updates[i] == nil {
				updates[i] = make(chan []byte, 1000) // If an go routine can't empty as fast as the main loop
			}
			updates[i] <- oprHash[:]
		}
		oprHash = sha256.Sum256(oprHash[:])
	}

	var best result
	var hashCnt int64
	start := time.Now()

	post()

	for i := 0; i < Instances; i++ {
		nonce = nonce<<32 ^ nonce<<9 ^ nonce>>3 ^ uint64(i)
		fmt.Printf("Nonce %d: %x\n", i, nonce)
		go func(update chan []byte, iNonce uint64) {
			var best uint64
			var a []byte
			a = <-update
			for {
				for len(update) > 0 {
					a = <-update
					best = 0
				}
				iNonce = iNonce<<17 ^ iNonce>>9 ^ uint64(hashCnt) ^ uint64(iNonce) // diff nonce for each instance
				atomic.AddInt64(&hashCnt, 1)
				nPow := lx.LxrPoW(a[:], iNonce)
				if nPow > best {
					results <- result{iNonce, nPow}
					if nPow > best {
						best = nPow
					}
				}
			}
		}(updates[i], nonce)
	}

	// pull the results, and print the best hashes
	hashTime := time.Now()
	var BlockCnt, TotalDifficulty uint64
	for i := 0; true; {
		if Timed && time.Since(hashTime) > time.Second*time.Duration(BlockTime) {
			BlockCnt++
			TotalDifficulty += best.pow
			AvgDifficulty := TotalDifficulty / BlockCnt
			current := time.Since(start)
			sinceLast := time.Since(hashTime)
			rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
			coreRate := rate / float64(Instances)
			average := current / time.Duration(BlockCnt)
			fmt.Printf("  %3s time: %10s Th: %10d Th/s: %12.5f Ch/s: %12.5f Pow: %016x Hash: %20x Nonce: %016x\n",
				"",
				fmt.Sprintf("%3d:%02d:%02d", int(sinceLast.Hours()), int(sinceLast.Minutes())%60, int(sinceLast.Seconds())%60),
				hashCnt, rate, coreRate, best.pow, oprHash[:16], best.nonce)
			fmt.Printf("Total Block Count: %d Average Block Time: %3d:%02d:%02d Average Difficulty: %016x\n\n",
				BlockCnt,
				int(average.Hours()), int(average.Minutes())%60, int(average.Seconds())%60,
				AvgDifficulty,
			)
			post()
			best = result{}
			hashTime = time.Now()
			fmt.Println()
			i = 0
		}
		time.Sleep(time.Millisecond * 200)
		for len(results) > 0 {
			r := <-results
			if r.pow > best.pow {
				i++
				best = r
				current := time.Since(start)
				sinceLast := time.Since(hashTime)
				rate := float64(hashCnt) / float64(current.Nanoseconds()) * 1000000000
				coreRate := rate / float64(Instances)
				fmt.Printf("  %3d time: %10s Th: %10d Th/s: %12.5f Ch/s: %12.5f Pow: %016x Hash: %20x Nonce: %016x\n",
					i,
					fmt.Sprintf("%3d:%02d:%02d", int(sinceLast.Hours()), int(sinceLast.Minutes())%60, int(sinceLast.Seconds())%60),
					hashCnt, rate, coreRate, r.pow, oprHash[:16], r.nonce)
				if !Timed && r.pow > ReportingBar {
					BlockCnt++
					TotalDifficulty += best.pow
					AvgDifficulty := TotalDifficulty / BlockCnt
					average := current / time.Duration(BlockCnt)
					fmt.Printf("Total Block Count: %d Average Block Time: %3d:%02d:%02d Average Difficulty: %016x\n\n",
						BlockCnt,
						int(average.Hours()), int(average.Minutes())%60, int(average.Seconds())%60,
						AvgDifficulty,
					)
					post()
					hashTime = time.Now()
					best = result{}
					i = 0
				}
			}
		}
	}
}
