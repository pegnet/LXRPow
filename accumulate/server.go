package accumulate

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pegnet/LXRPow/pow"
)

// Submission
// Miners create submissions that are recorded on
// acc://miningService/submissions scratch data account
type Submission struct {
	Valid      bool      // Not Persisted. Assumed valid, set to false if invalid
	TimeStamp  time.Time //  8 Miner reported timestamp
	DNIndex    uint64    //  8 Directory Network minor block index
	DNHash     [32]byte  // 40 Directory Network Index
	BlockIndex uint64    // 48 Mining Block Index
	Nonce      uint64    // 56 Nonce solution
	MinerIdx   uint64    // 64 Index into the Miners Account
	PoW        uint64    // 72 Self reported Difficulty
}

// Miner
// Miners register their token accounts on
// acc://miningService/miners data account.  If a token account
// is registered multiple times, then the first registration is the index
// used to track points
type Miner struct {
	TokenURL string
}

// Validator
// Validators evaluate Miner submissions and evaluate points and
// make token distributions
type Validator struct {
	KeyBookURL string
}

// Points
// Every payout (once a day) reports the current point balances of all miners on
// acc://miningService/points.
// To compute the points of a miner, the validators run through all the records on
// this account to add up the points for all the miners.  The
type Points struct {
	MinersIdx uint64 // Index into acc://miningService/miners to find the Token URL
	Points    uint64 // Current points for the Token URL
}

type PointsReport struct {
	BlockIndex uint64
	Points     []*Points
}

// Settings
// Mostly we update the Difficult
type Settings struct {
	TimeStamp        time.Time //  8 - Timestamp (persisted in nanoseconds)
	WindowBlockIndex uint64    //  8 - Window Index of start of difficulty adjustment
	WindowTimestamp  time.Time //  8 - Timestamp (persisted in nanoseconds)
	DiffWindow       uint16    //  2 - Adjustment window in proof of work blocks
	DNIndex          uint64    //  8 - Miner block height
	LastDiff         uint64    //  8 - Difficulty of the last rewarded (300 or less)
	BlockIndex       uint64    //  8 - Block Index of activation for this set of settings
	Loops            uint16    //  2 - Loops over the hash (more loops, slower hash)
	Bits             uint16    //  2 - Number of bits addressing the ByteMap (30 = 1 GB)
	DNHash           [32]byte  // 32 - Hash to be mined
	Difficulty       uint64    //  8 - Difficulty that marks the end of the Block
	BlockTime        uint16    //  2 - Target block time in seconds per block
	PayoutFreq       uint64    //  8 - Payouts per 24 hours (starting at 0:00 UTC)
	Qualifies        uint64    //  8 - Number of submissions that are given points in a block
	//                           112 Bytes gross total bytes
}

// MAdi
// This struct represents the state kept in the mining ADI
type MAdi struct {
	Submissions  []Submission      // Submissions from Miners
	Accepted     []Submission      // The Accepted winning entry for each block
	Settings     []Settings        // Settings needed by miners to mine
	Miners       map[string]uint64 // A look up table to make finding a miner's index fast
	MinersIdx    []string          // The actual list of miners as they were registered
	Validators   map[string]uint64 // A look up table to make finding a validator index fast
	ValidatorIdx []string          // The actual list of validators as then are registered
	PointsReport []PointsReport    // Reports of points earned by miners
	LX           *pow.LxrPow       // Not persisted.
}

var MAdiMutex sync.RWMutex
var MiningADI = new(MAdi)

// NextDNHash
// In the real implementation, we get this stuff from Accumulate
func (m *MAdi) NextDNHash() (timeStamp time.Time, DNIndex uint64, DNHash []byte) {
	var dNHash [32]byte
	settings := m.Sync()
	dNHash = sha256.Sum256(settings.DNHash[:])

	return time.Now(), 100000, dNHash[:]
}

// init
// Set up the miner ADI (MAdi) while we are not actually running against accumulate.
// This needs to be replaced by a call to get the settings from the minder ADI.
// We will still need to create the Miners Map, and parse through the MinerIdx to
// create the Miners Map, compute the current point balances, etc.
func init() {

	MiningADI.Miners = make(map[string]uint64)

	// In development and standalone testing, we need a settings record
	// This will be done using the CLI for Accumulate
	settings := new(Settings)                              // Create a Settings record
	settings.TimeStamp = time.Now()                        // Current time
	settings.DNIndex = 10000                               // Just pick an arbitrary minor block height
	settings.BlockIndex = 1                                // Start at the first block
	settings.Loops = 32                                    // LXPow: number of loops over the hash
	settings.Bits = 16                                     // LXPow: Number of bits in the Byte Map lookup
	settings.DNHash = sha256.Sum256([]byte("first block")) // Hash of the first DNBlock to be mined
	settings.Difficulty = 0xFFFFF00000000000               // A Starting difficulty target
	settings.BlockTime = 600                               // 10 minutes
	settings.PayoutFreq = 60 * 60 * 4                      // Every 4 hours (6 times a day)
	settings.Qualifies = 100                               // How many submissions get points

	MiningADI.Settings = append(MiningADI.Settings, *settings)
}

// RegisterMiner
func (m *MAdi) RegisterMiner(tokenUrl string) uint64 {
	if _, err := url.Parse(tokenUrl); err != nil {
		panic(fmt.Sprintf("bad url: '%s' -- %v", tokenUrl, err))
	}

	tokenUrl = strings.ToLower(tokenUrl)
	if idx, registered := m.Miners[tokenUrl]; registered { // Check Registry
		return idx //                                         Ignore registered miners
	}

	idx := uint64(len(m.MinersIdx))
	m.Miners[tokenUrl] = idx                    // Register the miner in the map
	m.MinersIdx = append(m.MinersIdx, tokenUrl) // and in the list

	return idx
}

// GetMinerUrl
// Takes the miner index and returns the miner's URL
func (m *MAdi) GetMinerUrl(minerIdx uint64) string {
	MAdiMutex.Lock()
	defer MAdiMutex.Unlock()

	if int(minerIdx) >= len(m.MinersIdx) {
		return ""
	}
	return m.MinersIdx[minerIdx]
}

// RegisterValidator
// Adds a Validator to the Validator data account, and updates the
// Map for easy lookups
func (m *MAdi) RegisterValidator(bookUrl string) (validatorIdx uint64) {
	if _, err := url.Parse(bookUrl); err != nil {
		panic(fmt.Sprintf("bad url: '%s' -- %v", bookUrl, err))
	}

	bookUrl = strings.ToLower(bookUrl)
	if idx, registered := m.Miners[bookUrl]; registered { // Check Registry
		return idx //                                         Ignore registered miners
	}

	idx := uint64(len(m.ValidatorIdx))
	m.Validators[bookUrl] = idx                      // Register the miner in the map
	m.ValidatorIdx = append(m.ValidatorIdx, bookUrl) // and in the list

	return idx
}

// Sync
// Syncs with the Accumulate Protocol so we can trust our data.
// If a nil is returned, then Syncing failed.
func (m *MAdi) Sync() Settings {
	MAdiMutex.Lock()
	defer MAdiMutex.Unlock()
	return m.Settings[len(m.Settings)-1]

}

// GetBlock
// Return
// DNHash:      the current DNHash being mined,
// Submissions: all the submissions so far
// the current block, and the lowest difficulty submitted.  If not synced,
// the DNHash returned is nil
func (m *MAdi) GetBlock() (DNHash [32]byte, submissions []Submission) {
	settings := m.Sync()

	MAdiMutex.Lock()
	if len(m.Submissions) == 0 {
		MAdiMutex.Unlock()
		return settings.DNHash, submissions
	}
	raw := append([]Submission{}, m.Submissions[settings.BlockIndex:]...)
	MAdiMutex.Unlock()
	var clean []Submission

	for _, sub := range raw {
		if sub.DNHash != settings.DNHash {
			continue
		}
		if sub.BlockIndex != settings.BlockIndex {
			continue
		}

		clean = append(clean, sub)
	}
	sort.Slice(clean, func(i, j int) bool { return clean[i].PoW < clean[j].PoW })
	return settings.DNHash, clean
}

// AddSubmission
// Does quality checks on the submission to avoid adding submissions
// that cannot win points and wasting credits.
func (m *MAdi) AddSubmission(sub Submission) error {
	settings := m.Sync()
	_, submissions := m.GetBlock()
	if len(submissions) > 0 && submissions[len(submissions)-1].PoW >= settings.Difficulty {
		return nil // A solution has already been found
	}
	if sub.PoW < 0xFFF0000000000000 && sub.PoW < settings.Difficulty {
		return nil // Ignore trivial difficulties
	}
	if sub.DNHash == settings.DNHash {
		MAdiMutex.Lock()
		m.Submissions = append(m.Submissions, sub)
		MAdiMutex.Unlock()
	}
	return nil
}

// AddSettings
// Add a Settings Record to the Settings Account
func (m *MAdi) AddSettings(settings Settings) {
	MAdiMutex.Lock()
	defer MAdiMutex.Unlock()
	m.Settings = append(m.Settings, settings)
}
