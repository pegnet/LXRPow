// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package hashing

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/pegnet/LXRPow/pow"
)

func Test_HasherSet(t *testing.T) {
	lx := new(pow.LxrPow)
	lx.Init(32, 30, 6)
	m := NewHashers(7, 523452345, lx)
	m.Start()
	hash := sha256.Sum256([]byte{1, 2, 3, 4})

	fmt.Print("\nStart Hashing\n\n")
	m.BlockHashes <- Hash{Hash: hash, Limit: 0xFFF0000000000000}

	var best PoWSolution
	for i := 0; i < 1000; {
		s := <-m.Solutions
		if s.Pow > best.Pow || s.Pow > 0xffff000000000000 {
			var th uint64
			for _, v := range m.Instances {
				th += v.HashCnt
			}
			if best.Pow == 0 {
				fmt.Println()
				fmt.Printf("%6s%v%12s%v%12s%v%13s%v\n", "", "Hash", "", "Nonce", "", "PoW", "", "Hasher")
			}
			best = s
			fmt.Printf("%08x %016x %016x %10d\n", s.DNHash[:8], s.Nonce, s.Pow, th)
			if s.Pow > 0xFFFFF00000000000 {
				best.Pow = 0
				hash = sha256.Sum256(hash[:])
				m.BlockHashes <- Hash{Hash: hash, Limit: 0xFFF0000000000000}
				i++
			}
		}
	}
	m.Stop()
	time.Sleep(time.Second)
}
