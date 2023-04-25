// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"crypto/sha256"
	"encoding/binary"
)

type LxrPow struct {
	Loops   int    // The number of loops translating the ByteMap
	ByteMap []byte // Integer Offsets
	MapSize uint64 // Size of the translation table (must be a factor of 256)
	Passes  int    // Passes to generate the rand table
}

// NewLxrPow
//
// Return a new instance of the LxrPow work function
// Loops  - the number of passes through the hash made through the hash itself.  More 
//          loops will slow down the hash function.
// Bits   - number of bits used to create the ByteMap. 30 bits creates a 1 GB ByteMap
// Passes - Number of shuffles used to randomize the ByteMap. 6 seems sufficient
// 
// Any change to Loops, Bits, or Passes will map the PoW to a completely different 
// space.
func NewLxrPow(Loops, Bits, Passes int) *LxrPow {
	lx := new(LxrPow)
	lx.Init(Loops, Bits, Passes)
	return lx
}

// LxrPoW() returns a 64 byte value indicating the proof of work of a hash
// This is designed to allow the use of any hash function, but make the grading
// of the proof of work dependent on the random byte access limits of LXRHash.
//
// The bigger uint64, the more PoW it represents.  The first byte is the
// number of leading bytes of FF, followed by the leading "non FF" bytes of the pow.
func (lx LxrPow) LxrPoW(hash []byte, nonce uint64) (pow uint64) {
	mask := lx.MapSize - 1

	var LHash [40]byte // Compute a 40 byte value from the hash and nonce
	binary.BigEndian.PutUint64(LHash[:],nonce) // Put the nonce firsst
	copy(LHash[8:],hash) // Then the given hash
	lHash := sha256.Sum256(LHash[:]) // Do a sha256 of the 40 bytes 
	copy(LHash[:],lHash[:32]) // Copy over the result

	state := binary.BigEndian.Uint64(lHash[:])

	// Make the specified "loops" through the LHash.  This is 40 bytes; 32 from the sha256 and
	// 8 bytes from the trailing part of the hash.  Keeping or not keeping the trailing 8 only
	// changes how many translations per loop are done.
	for i := 0; i < lx.Loops; i++ {
		for j,v := range LHash {
			state = state<<17 ^ state>>7 ^ uint64(lx.ByteMap[state&mask]^v)
			LHash[j] = byte(state)
		}
	}

	final := sha256.Sum256(LHash[:]) // The Entire LHash translation must be done (can't skip anything)
	return lx.pow(final[:32])        // Return the pow of the sha256 of the translated hash   
}

// Return the first 8 bytes of the modified hash as the difficulty
func (lx LxrPow) pow(hash []byte) (pow uint64) {
	pow = binary.BigEndian.Uint64(hash)
	return pow // Return the pow
}
