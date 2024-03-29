package hashing_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	. "github.com/pegnet/LXRPow/hashing"
	"github.com/pegnet/LXRPow/pow"
)

func Test_Hasher(t *testing.T) {
	lx := new(pow.LxrPow)
	lx.Init(32, 30, 6)
	m := NewHasher(1, 1000, lx)
	m.Start()
	hash := sha256.Sum256([]byte{1, 2, 3, 4})
	m.BlockHashes <- Hash{Hash: hash, Limit: 0xFFF0000000000000}
	start := time.Now()
	for i := 0; i < 40; i++ {
		s := <-m.Solutions
		fmt.Printf("%08x %016x %016x %016d\n", s.DNHash[:8], s.Nonce, s.Pow, s.HashCnt)
		if s.Pow > 0xFFFF000000000000 {
			hash = sha256.Sum256(hash[:])
			m.BlockHashes <- Hash{Hash: hash, Limit: 0xFFF0000000000000}
			fmt.Println()
			hs := float64(time.Since(start).Milliseconds()) / 1000
			fmt.Printf("Hashes Per Second = %8.3f\n", float64(s.HashCnt)/float64(hs))

		}
	}
}
