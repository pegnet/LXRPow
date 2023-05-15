package validator

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/pow"
)

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

// EndOfBlock
// Returns true if we have an end of block, and the index of the last
// entry of the block (first to be > difficulty)
// Assumes submissions are sorted
func (v *Validator) EndOfBlock(settings *accumulate.Settings, submissions []accumulate.Submission) (bool, int) {
	if len(submissions) == 0 || submissions[len(submissions)-1].PoW < settings.Difficulty {
		return false, 0
	}
	valid := -1
	i := len(submissions) - 1
	for ; i >= 0; i-- {
		if submissions[i].PoW < settings.Difficulty {
			i++ // Add back to the last valid
			break
		} else if accumulate.ValidateSubmission(v.LX, *settings, submissions[i]) {
			valid = i
		}

	}
	return valid >= 0, valid
}

// TrimToBlock
// Return a list of submissions that are trimmed to the 300 that will have points counted
func (v *Validator) TrimToBlock(settings accumulate.Settings, submissions []accumulate.Submission) []accumulate.Submission {

	var PointWinners []accumulate.Submission
	for _, s := range submissions {
		if accumulate.ValidateSubmission(v.LX, settings, s) {
			PointWinners = append(PointWinners, s)
		}
	}
	return PointWinners
}

// Start
// Updates the settings on Accumulate
func (v *Validator) Start() {
	for {
		time.Sleep(time.Second)
		settings := accumulate.MiningADI.Sync()
		dnHash, submissions := accumulate.MiningADI.GetBlock()
		EoB, lastEntry := v.EndOfBlock(&settings, submissions)
		if EoB {

			submissions = v.TrimToBlock(settings, submissions[:lastEntry+1])
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

			var sum float64
			for _, v := range v.BlockTimes {
				sum += v
			}
			btl := float64(len(v.BlockTimes))

			go func(submissions []accumulate.Submission) {
				out := "TimeStamp                 DNIndex DNHash           BlockIndex Nonce            MinerIdx  Pow              Url\n"
				for _, n := range submissions {
					t := n.TimeStamp.Format("01/02/06 03:04:05.000")
					url := accumulate.MiningADI.GetMinerUrl(n.MinerIdx)
					out += fmt.Sprintf("%25s %7d %016x %10d %016x %9d %016x %s\n",
						t, n.DNIndex, n.DNHash[:8], n.BlockIndex, n.Nonce, n.MinerIdx, n.PoW, url)
				}
				out += "TimeStamp                 DNIndex DNHash           BlockIndex Nonce            MinerIdx  Pow              Difficulty       Url\n"

				out += fmt.Sprintf("TargetBlockTime=   %4d\n", settings.BlockTime)
				out += fmt.Sprintf("AvgBlockTime=    %8.3f\n", sum/btl)
				out += fmt.Sprintf("Difficulty=         %016x\n", settings.Difficulty)
				out += fmt.Sprintf("PreviousDiff=       %016x\n", v.OldDiff)
				out += fmt.Sprintf("BlockTime =         %d:%06.3f\n", minutes, seconds)
				out += fmt.Sprintf("#submissions=       %d\n", len(submissions))
				out += fmt.Sprintf("block number=       %d\n", settings.BlockIndex)
				out += fmt.Sprintf("DNBlock number=     %d\n\n", settings.DNIndex)
				fmt.Print(out)
			}(submissions)
			accumulate.MiningADI.AddSettings(newSettings)

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
	if uint16(len(v.BlockTimes)) >= settings.DiffWindow {

		var sum float64
		var cnt float64
		for _, v := range v.BlockTimes {
			sum += v
			cnt += 1
		}
		avgBlkTime := sum / cnt

		dPercent := (float64(settings.BlockTime) - avgBlkTime) / 2
		if dPercent > .5 {
			dPercent = .5
		}
		if dPercent < -.5 {
			dPercent = -.5
		}
		fDifficulty := float64(-settings.Difficulty)  // Use the 2's complement of the Difficulty because leading F's (in hex) are hard to deal with
		nd := uint64(-(fDifficulty * (1 - dPercent))) // Edge closer by 1/5 the error since we are using a 128 blk average
		v.BlockTimes = v.BlockTimes[:0]               // Clear the list of block times

		return lastBlockTime, nd, uint16(settings.BlockIndex + 1)
	}
	return lastBlockTime, settings.Difficulty, uint16(settings.WindowBlockIndex)
}

func (v *Validator) Grade(settings accumulate.Settings, submissions []accumulate.Submission) (winner accumulate.Submission, difficulty uint64) {

	winner = submissions[len(submissions)-1]

	return winner, settings.Difficulty

}
