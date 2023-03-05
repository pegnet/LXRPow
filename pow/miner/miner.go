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
	pDifficulty := flag.Uint64("difficulty", 0x0300<<48, "Difficulty target (timed) or difficulty termination (not timed)")
	pDiffWindow := flag.Int("diffwindow", 1000, "Difficulty Target Valuation in blocks")
	pBlockTime := flag.Int("blocktime", 600, "Block Time in seconds (600 would be 10 minutes)")
	pTimed := flag.Bool("timed", false, "Blocks are timed, or blocks end with a given difficulty")
	flag.Parse()

	Instances := *pInstances
	Loop := *pLoop
	Bits := *pBits
	Data := *pData
	Phrase := *pPhrase
	Difficulty := *pDifficulty
	DiffWindow := *pDiffWindow
	BlockTime := *pBlockTime
	Timed := *pTimed

	lx := new(pow.LxrPow).Init(Loop, Bits, 6)
	oprHash := sha256.Sum256([]byte(Data))
	phraseHash := sha256.Sum256([]byte(Phrase))
	nonce := binary.BigEndian.Uint64(phraseHash[:]) + uint64(time.Now().UnixNano())

	fmt.Printf("\nminer --instances=%d --loop=%d --bits=%d --data=\"%s\""+
		" --phrase=\"%s\" --difficulty=0x%x --diffwindow=%d --blocktime=%d --timed=%v\n\n",
		Instances,
		Loop,
		Bits,
		Data,
		Phrase,
		Difficulty,
		DiffWindow,
		BlockTime,
		Timed,
	)
	fmt.Printf("Filename: out-instances%d-loop%d-difficulty0x%x-diffwindow%d-blocktime%d-timed_%v.txt\n\n",
		Instances, Loop, Difficulty, DiffWindow, BlockTime, Timed)

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
					avgSec := int(average.Seconds()) % 60
					avgMs := int((average.Seconds()-float64(avgSec))*1000)
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
