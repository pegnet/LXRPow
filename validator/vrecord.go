package validator

import (
	"net/url"
)

type Payout struct {
	tokenURL *url.URL
	amount   uint64
}
// ValidatorRecord
// Persisted out as a set of payouts, sorted by ADI
// Only one payout to an ADI is allowed per block
type ValidatorRecord struct {
	payouts map [string] *Payout  // Payout indexed by ADI
}


