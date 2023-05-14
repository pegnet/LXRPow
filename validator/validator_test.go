package validator

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/pow"
)

func Test_Difficulty(t *testing.T) {
	var s float64
	currentDiff := float64(20)

	var window []float64 // Used to compute the median

	w := 100 * 60 / 10 // 4 hours, 10 minute blocks
	target := float64(600)
	min := target * 100
	max := float64(0)

	for i := 1; i <= 45000; i++ {
		r := rand.ExpFloat64() * currentDiff
		s += r
		window = append(window, r)
		if r < min {
			min = r
		}
		if r > max {
			max = r
		}

		if i%w == 0 {
			a := s / float64(w)
			sort.Float64s(window)
			median := window[w/2]
			percentDelta := (target - a) / target
			nx := currentDiff * (1 + percentDelta)
			f := float64(10)
			if nx > f*currentDiff {
				nx = f * currentDiff
			}
			if nx < currentDiff/f {
				nx = currentDiff / f
			}

			fmt.Printf("Index avg = %8.3f, median = %8.3f min = %8.3f max = %8.3f "+
				"x = %8.3f window = %8d target = %8.1f \n", a, median, min, max, nx, w, target)

			currentDiff = nx
			s = 0
			min = target * 100
			max = 0
			window = window[:0]
		}

	}
}

func TestValidator_AdjustDifficulty(t *testing.T) {

	v := NewValidator("bob.acme/tokens", nil)
	settings := new(accumulate.Settings)
	settings.BlockTime = 10
	settings.DiffWindow = 10
	settings.Difficulty = 0xff00000000000000
	settings.TimeStamp = time.Now()
	settings.BlockIndex = 1
	for {
		newSettings := new(accumulate.Settings)
		newSettings.BlockIndex = settings.BlockIndex + 1
		newSettings.BlockTime = 10
		newSettings.DiffWindow = 10
		newSettings.Difficulty = settings.Difficulty
		newSettings.TimeStamp = settings.TimeStamp.Add(15000 * time.Millisecond)
		_, newDiff, _ := v.AdjustDifficulty(*settings, *newSettings)
		if newDiff != newSettings.Difficulty {
			v.OldDiff = newSettings.Difficulty
			newSettings.Difficulty = newDiff
			fmt.Printf("%x\n", newDiff)
			time.Sleep(time.Second / 2)
		}
		settings = newSettings
	}
}

func TestEndOfBlock(t *testing.T) {
	LX := pow.NewLxrPow(16, 30, 6)
	submissions := []accumulate.Submission{}
	DNHash := sha256.Sum256([]byte{1, 2, 3})
	v := NewValidator("a.acme", LX)
	settings := new(accumulate.Settings)
	settings.BlockIndex = 1
	settings.DNIndex = 1
	settings.DNHash = DNHash
	settings.DNIndex = 1
	settings.Difficulty = 0xFFF8000000000000
	var s accumulate.Submission
	found := false
	var n uint64
	for i := 0; ; {
		if found && i >= 300 {
			break
		}
		if i == len(submissions) {
			s.BlockIndex = 1
			s.DNHash = DNHash
			s.DNIndex = 1
			s.MinerIdx = accumulate.MiningADI.RegisterMiner("a.acme")
			s.Nonce = 1
			s.PoW = 1
			s.TimeStamp = time.Now()
			submissions = append(submissions, s)
		}
		pow := LX.LxrPoW(DNHash[:], n)
		if pow > 0xFF00000000000000 {
			submissions[i].Nonce = n
			submissions[i].PoW = pow
			i++
		}
		if pow > settings.Difficulty {
			found = true
		}
		n++
	}
	sort.Slice(submissions, func(i, j int) bool { return submissions[i].PoW < submissions[j].PoW })

	printSubs := func(s []accumulate.Submission) {
		printed := false
		for i, s := range s {
			printed = false
			fmt.Printf("%4d %x ", i, s.PoW)
			if (i+1)%10 == 0 {
				fmt.Println()
				printed = true
			}
		}
		if !printed {
			fmt.Println()
		}
		fmt.Println()
	}
	printSubs(submissions)
	end, idx := v.EndOfBlock(settings, submissions)
	if idx < 0 {
		idx = 0
	}
	fmt.Printf("%v %d %x\n", end, idx, submissions[idx].PoW)
	{
		end2, idx2 := v.EndOfBlock(settings, submissions[:idx])
		if idx2 < 0 {
			idx2 = 0
		}
		fmt.Printf("%v %d %x\n", end2, idx, submissions[idx2].PoW)
	}

	end, idx = v.EndOfBlock(settings, submissions[:idx+1])
	if idx < 0 {
		idx = 0
	}
	fmt.Printf("%v %d %x\n", end, idx, submissions[idx].PoW)

	block := v.TrimToBlock(*settings, submissions[:idx+1])
	fmt.Printf("Block Len %d \n", len(block))
	printSubs(block)

	corrupt := append([]accumulate.Submission{}, submissions...)
	for i := 0; i < 30; i++ {
		corrupt[(i*17)%len(corrupt)].Nonce++
	}
	cEnd, cIdx := v.EndOfBlock(settings, corrupt)
	cBlock := v.TrimToBlock(*settings, corrupt[:cIdx+1])
	fmt.Printf("%v %d %x\n", cEnd, cIdx, submissions[idx].PoW)
	printSubs(cBlock)
}
