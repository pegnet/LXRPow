// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"
)

type Config struct {
	TokenURL   string // URL for rewards
	Instances  int    // How many hashers to run
	Loop       int    // How many times we loop over a hash computing PoW
	Bits       int    // Number of bits in the size of the ByteMap (30 == 1GB ByteMap)
	Phrase     string // A phrase used to create the seed nonce for mining
	Difficulty uint64 // The difficulty limit (if using difficulty to end mining blocks)
	Window     int    // Determines Difficulty adjustments, when ending blocks with difficulty
	BlockTime  int    // Used when ending blocks with time (uniform blocks)
	Timed      bool   // True if using timed blocks, false using difficulty

	Seed uint64
}

// Init
// What 
func (c *Config) Init() {

	pTokenURL := flag.String("tokenurl","","URL for where rewards go, and identify the ADI")
	pInstances := flag.Int("instances", 1, "Number of instances of the hash miners")
	pLoop := flag.Int("loop", 50, "Number of loops accessing ByteMap (more is slower)")
	pBits := flag.Int("bits", 30, "Number of bits addressing the ByteMap (more is bigger)")
	pPhrase := flag.String("phrase", "", "private phrase hashed to ensure unique nonces for the miner")
	pDifficulty := flag.Uint64("difficulty", 0x0300<<48, "Difficulty target (timed) or difficulty termination (not timed)")
	pDiffWindow := flag.Int("diffwindow", 1000, "Difficulty Target Valuation in blocks")
	pBlockTime := flag.Int("blocktime", 600, "Block Time in seconds (600 would be 10 minutes)")
	pTimed := flag.Bool("timed", false, "Blocks are timed, or blocks end with a given difficulty")
	flag.Parse()

	c.TokenURL = *pTokenURL
	c.Instances = *pInstances
	c.Loop = *pLoop
	c.Bits = *pBits
	c.Phrase = *pPhrase
	c.Difficulty = *pDifficulty
	c.Window = *pDiffWindow
	c.BlockTime = *pBlockTime
	c.Timed = *pTimed

	h := sha256.Sum256([]byte(c.Phrase))
	c.Seed = binary.BigEndian.Uint64(h[:]) + uint64(time.Now().UnixNano())

	fmt.Printf("\nminer --instances=%d --loop=%d --bits=%d --phrase=\"%s\" --difficulty=0x%x"+
		" --diffwindow=%d --blocktime=%d --timed=%v\n\n",
		c.Instances, c.Loop, c.Bits, c.Phrase, c.Difficulty, c.Window, c.BlockTime, c.Timed,
	)
	fmt.Printf("Filename: out-instances%d-loop%d-difficulty0x%x-diffwindow%d-blocktime%d-timed_%v.txt\n\n",
		c.Instances, c.Loop, c.Difficulty, c.Window, c.BlockTime, c.Timed)
	
	if	!ConfigIsValid(c) {
		os.Exit(1)
	}
}

func NewConfig() *Config {
	c := new(Config)
	c.Init()
	return c
}

// ConfigIsValid
// Do everything we can to validate the configuration provided to us.  Right
// now everything is on the commandline, and we are doing some checks.  But this
// could be more complex and more user friendly; that will be done here.
func ConfigIsValid(cfg *Config) bool {
	success := true
	if cfg.TokenURL == "" {
		fmt.Println("must have a ADI based token url for rewards")
		success = false
	}
	_, err := url.Parse(cfg.TokenURL)
	if err != nil {
		fmt.Println("token url provided is not a valid url")
		success = false
	}
	// Add other tests like query the protocol that the token account actually exists
	return success
}
