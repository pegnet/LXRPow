package hashing_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	. "github.com/pegnet/LXRPow/hashing"
	"github.com/pegnet/LXRPow/pow"
)

func Test_Hasher(t *testing.T) {
	lx := new(pow.LxrPow)
	lx.Init(32, 30, 6)
	m := NewHasher(1, 1000, lx)
	m.Start()
	hash := sha256.Sum256([]byte{1, 2, 3, 4})
	m.BlockHashes <- hash[:]

	for i := 0; i < 40; i++ {
		s := <-m.Solutions
		fmt.Printf("%08x %016x %016x %016d\n", s.DNHash[:8], s.Nonce, s.Difficulty, s.HashCnt)
		if s.Difficulty > 0xFFFF000000000000 {
			hash = sha256.Sum256(hash[:])
			m.BlockHashes <- hash[:]
			fmt.Println()
		}
	}

}
