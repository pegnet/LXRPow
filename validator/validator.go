package validator

import (
	"github.com/pegnet/LXRPow/accumulate"
	"github.com/pegnet/LXRPow/cfg"
)

type Validator struct {
	SubmissionHeight uint64      // Height of submissions by miners processed
	API              *accumulate.APIStub
}

func IsValidator() bool {
	return true
}

// Validator
// The PegNet Staking Validators have to run this code.  They
// validate the submissions of DN Hashes, their PoW, and process
// rewards to winning miners.
func (v *Validator) Init(cfg *cfg.Config) {
	v.API = accumulate.NewAPIStub(cfg)
}

// Start
func (v *Validator) Start() {
	
	for {

	}

}
