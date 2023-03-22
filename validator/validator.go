package validator

import (
	"fmt"

	"github.com/pegnet/LXRPow/cfg"
)

type Validator struct {
	Cfg              *cfg.Config // Staking Config
	SubmissionHeight uint64      // Height of submissions by miners processed

}

func IsValidator() bool {

}

// Validator
// The PegNet Staking Validators have to run this code.  They
// validate the submissions of DN Hashes, their PoW, and process
// rewards to winning miners.
func Init(cfg *cfg.Config) {
	
}
