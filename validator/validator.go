package validator

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pegnet/LXRPow/accumulate"
)

type Validator struct {
	URL string
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
				// Need to add grading and point tracking
				fmt.Print("TimeStamp                DNIndex DNHash           BlockIndex Nonce            MinderIdx Pow              Url\n")
				for _, n := range submissions {
					t := n.TimeStamp.Format(time.RFC822)
					fmt.Printf("%24s %7d %016x %10d %016x %9d %016x %s\n",
						t, n.DNIndex, n.DNHash[:8], n.BlockIndex, n.Nonce, n.MinerIdx, n.PoW, n.TokenURL)
				}
				newSettings := settings.Clone()
				newSettings.DNHash = sha256.Sum256(dnHash[:])
				newSettings.TimeStamp = time.Now()
				newSettings.DNIndex = settings.DNIndex + 100
				newSettings.BlockIndex = settings.BlockIndex + 1
				// Add update Difficulty here

				accumulate.MiningADI.AddSettings(newSettings)
			}
		}
	}
}
