// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package cfg

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"os"

	"github.com/pegnet/LXRPow/pow"
	"github.com/pegnet/LXRPow/sim"
)

type Config struct {
	Index      uint64              // Index of this mining instance
	TokenURL   string              // URL for rewards
	Instances  int                 // How many hashers to run
	MinerCnt   int                 // Number of miners to run
	Loop       int                 // How many times we loop over a hash computing PoW
	Bits       int                 // Number of bits in the size of the ByteMap (30 == 1GB ByteMap)
	Phrase     string              // A phrase used to create the seed nonce for mining
	Randomize  bool                // Use an OS generated random number to avoid seed collisions
	Difficulty uint64              // The difficulty limit (if using difficulty to end mining blocks)
	Window     int                 // Determines Difficulty adjustments, when ending blocks with difficulty
	BlockTime  float64             // Used when ending blocks with time (uniform blocks)
	Timed      bool                // True if using timed blocks, false using difficulty
	Seed       uint64              // Seed for all the miners
	LX         *pow.LxrPow         // The Proof of work function to be used.
}

// Return a shallow copy of the configuration settings.
func (c Config) Clone() *Config {
	c.TokenURL = sim.GetURL()
	c.RandomizeSeed()
	return &c
}

func (c *Config) RandomizeSeed() {
	max := big.NewInt(math.MaxInt64)     // Use the Cryptographic library to generate
	r, err := rand.Int(rand.Reader, max) //    a random number
	if err != nil {                      // If this fails, panic, because we have no recourse
		panic("randomize of seed failed")
	}
	c.Seed ^= r.Uint64() //                  XOR the random number into the seed
}

// Init
// What
func (c *Config) Init() {
	args := os.Args
	_ = args
	pIndex := flag.Uint64("index", 1, "Index of the miner, where many miners may work together")
	pTokenURL := flag.String("tokenurl", "RedWagon.acme/tokens", "URL for where rewards go, and identify the ADI")
	pInstances := flag.Int("instances", 1, "Number of instances of the hash miners")
	pMinerCnt := flag.Int("minercnt", 1, "Number of miners (with random URLs) to run")
	pLoop := flag.Int("loop", 50, "Number of loops accessing ByteMap (more is slower)")
	pBits := flag.Int("bits", 30, "Number of bits addressing the ByteMap (more is bigger)")
	pPhrase := flag.String("phrase", "", "private phrase hashed to ensure unique nonces for the miner")
	pRandomize := flag.Bool("randomize", true, "randomize seed to lesson chances of collision with other miners")
	pDifficulty := flag.Uint64("difficulty", 0xffff<<48, "Difficulty target (timed) or difficulty termination (not timed)")
	pDiffWindow := flag.Int("diffwindow", 1000, "Difficulty Target Valuation in blocks")
	pBlockTime := flag.Float64("blocktime", 600, "Block Time in seconds (600 would be 10 minutes)")
	pTimed := flag.Bool("timed", false, "Blocks are timed, or blocks end with a given difficulty")
	flag.Parse()

	c.Index = *pIndex
	c.TokenURL = *pTokenURL
	c.Instances = *pInstances
	c.MinerCnt = *pMinerCnt
	c.Loop = *pLoop
	c.Bits = *pBits
	c.Phrase = *pPhrase
	c.Randomize = *pRandomize
	c.Difficulty = *pDifficulty
	c.Window = *pDiffWindow
	c.BlockTime = *pBlockTime
	c.Timed = *pTimed

	h := sha256.Sum256([]byte(c.Phrase))

	c.Seed = binary.BigEndian.Uint64(h[:]) ^ c.Index // Use the Phrase and index to build the seed

	if c.Randomize { //                         The User may wish to randomize the seed
		c.RandomizeSeed()
	}

	if c.MinerCnt < 1 { // The least number of miners we will run is 1
		c.MinerCnt = 1
	}

	fmt.Printf("\nminer --index=%d --tokenurl=\"%s\" --instances=%d --minercnt=%d --loop=%d --bits=%d --phrase=\"%s\""+
		" --randomize=%v --difficulty=0x%x --diffwindow=%d --blocktime=%d --timed=%v\n\n",
		c.Index, c.TokenURL, c.Instances, c.MinerCnt, c.Loop, c.Bits, c.Phrase,
		c.Randomize, c.Difficulty, c.Window, c.BlockTime, c.Timed,
	)
	fmt.Printf("Filename: out-instances%d-minercnt%d-loop%d-difficulty0x%x-diffwindow%d-blocktime%d-timed_%v.txt\n\n",
		c.Instances, c.MinerCnt, c.Loop, c.Difficulty, c.Window, c.BlockTime, c.Timed)

	if !ConfigIsValid(c) {
		os.Exit(1)
	}

	c.LX = pow.NewLxrPow(c.Loop, c.Bits, 6)
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
		fmt.Println("must have a (tokenurl) ADI based token url for rewards")
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
