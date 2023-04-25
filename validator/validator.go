package validator

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/pow"
)

var DifficultyPeriod = 75 // When Adjusting difficulty, the average time of the last
//                  128 blocks is used.  The exponential distribution of block times
//                  is such that a period less than 128 doesn't really work.

type Validator struct {
	URL        string
	LX         *pow.LxrPow
	BlockTimes []float64
	OldDiff    uint64 // A working value
}

// NewValidator
func NewValidator(url string, lx *pow.LxrPow) *Validator {
	v := new(Validator)
	v.URL = url
	v.LX = lx
	return v
}

// Start
// Updates the settings on Accumulate
func (v *Validator) Start() {
	for {
		time.Sleep(time.Second)
		settings := accumulate.MiningADI.Sync()
		dnHash, submissions := accumulate.MiningADI.GetBlock()
		if len(submissions) > 0 {
			last := submissions[len(submissions)-1]
			if last.PoW > settings.Difficulty {

				// If this instance is the leader, compute and write the settings.  Need to
				// add some code to wait for the new settings, then validate the new settings,
				// and select a new leader if the settings written by the leader are not valid.
				newSettings := settings
				newSettings.DNHash = sha256.Sum256(dnHash[:])
				newSettings.TimeStamp = time.Now()
				newSettings.DNIndex = settings.DNIndex + 100
				newSettings.BlockIndex = settings.BlockIndex + 1
				LastBlockTime, newDiff, windowBlockIndex := v.AdjustDifficulty(settings, newSettings)
				newSettings.WindowBlockIndex = uint64(windowBlockIndex)
				if newDiff != newSettings.Difficulty {
					v.OldDiff = newSettings.Difficulty
					newSettings.Difficulty = newDiff
				}

				// Need to add grading and point tracking
				minutes := int(LastBlockTime) / 60
				seconds := LastBlockTime - float64(minutes*60)

				fmt.Print("TimeStamp                 DNIndex DNHash           BlockIndex Nonce            MinderIdx Pow              Url\n")
				for _, n := range submissions {
					t := n.TimeStamp.Format("01/02/06 03:04:05.000")
					url := accumulate.MiningADI.GetMinerUrl(n.MinerIdx)
					fmt.Printf("%25s %7d %016x %10d %016x %9d %016x %s\n",
						t, n.DNIndex, n.DNHash[:8], n.BlockIndex, n.Nonce, n.MinerIdx, n.PoW, url)
				}
				var sum float64
				for _, v := range v.BlockTimes {
					sum += v
				}
				btl := float64(len(v.BlockTimes))
				fmt.Print("TimeStamp                 DNIndex DNHash           BlockIndex Nonce            MinderIdx Pow              Difficulty       Url\n")
				fmt.Printf("AvgBlockTime=   %8.3f\n", sum/btl)
				fmt.Printf("Difficulty=     %016x\n", settings.Difficulty)
				fmt.Printf("PreviousDiff=   %016x\n", v.OldDiff)
				fmt.Printf("BlockTime =     %d:%06.3f\n", minutes, seconds)
				fmt.Printf("#submissions=   %d\n", len(submissions))
				fmt.Printf("block number=   %d\n", last.BlockIndex)
				fmt.Printf("DNBlock number= %d\n\n", last.DNIndex)

				accumulate.MiningADI.AddSettings(newSettings)

			}
		}
	}
}

// AdjustDifficulty
// Returns the time it took to produce the last hash in seconds, and the new difficulty
// if an adjustment is required.
func (v *Validator) AdjustDifficulty(settings, newSettings accumulate.Settings) (
	lastBlockTime float64, difficulty uint64, WindowBlockIndex uint16) {

	lastBlockTime = newSettings.TimeStamp.Sub(settings.TimeStamp).Seconds()
	v.BlockTimes = append(v.BlockTimes, lastBlockTime)
	if uint16(settings.BlockIndex-settings.WindowBlockIndex+1) >= settings.Window {

		var sum float64
		var cnt float64
		for _, v := range v.BlockTimes {
			sum += v
			cnt += 1
		}
		btl := float64(len(v.BlockTimes))
		avgBlkTime := sum / cnt
		if len(v.BlockTimes) >= DifficultyPeriod {
			copy(v.BlockTimes[:settings.Window], v.BlockTimes[settings.Window:])
			v.BlockTimes = v.BlockTimes[:DifficultyPeriod-int(settings.Window)]
		}

		dPercent := (float64(settings.BlockTime) - avgBlkTime) / btl
		if dPercent > .5 {
			dPercent = .5
		}
		if dPercent < -.5 {
			dPercent = -.5
		}
		fDifficulty := float64(-settings.Difficulty)  // Use the 2's complement of the Difficulty because leading F's (in hex) are hard to deal with
		nd := uint64(-(fDifficulty * (1 - dPercent))) // Edge closer by 1/5 the error since we are using a 128 blk average
		return lastBlockTime, nd, uint16(settings.BlockIndex + 1)
	}
	return lastBlockTime, settings.Difficulty, uint16(settings.WindowBlockIndex)
}

func (v *Validator) ValidateSubmission(settings accumulate.Settings, submission accumulate.Submission) bool {
	switch {
	case settings.BlockIndex != submission.BlockIndex:
		return false
	case settings.DNHash != submission.DNHash:
		return false
	case submission.DNIndex != settings.DNIndex:
		return false
	}
	if accumulate.MiningADI.GetMinerUrl(submission.MinerIdx) == "" { // Miner must be registered
		return false
	}

	// Because testing the Pow is the most expensive test, we do it last
	actualPow := v.LX.LxrPoW(submission.DNHash[:], submission.Nonce)
	return actualPow == submission.PoW
}

func (v *Validator) Grade(settings accumulate.Settings, submissions []accumulate.Submission) (winner accumulate.Submission, difficulty uint64) {

	winner = submissions[len(submissions)-1]

	return winner, settings.Difficulty

}
