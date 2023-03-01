// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"crypto/sha256"
)

type LxrPow struct {
	Loops   int    // The number of loops translating the ByteMap
	ByteMap []byte // Integer Offsets
	MapSize uint64 // Size of the translation table (must be a factor of 256)
	Passes  uint64 // Passes to generate the rand table
}

// LxrPoW() returns a 64 byte value indicating the proof of work of a hash
// This is designed to allow the use of any hash function, but make the grading
// of the proof of work dependent on the random byte access limits of LXRHash.
//
// The bigger uint64, the more PoW it represents.  The first byte is the
// number of leading bytes of FF, followed by the leading "non FF" bytes of the pow.
func (lx LxrPow) LxrPoW(hash []byte, nonce uint64) (pow uint64) {
	mask := lx.MapSize - 1

	LHash := append([]byte{},
		byte(nonce), byte(nonce>>8), byte(nonce>>16), byte(nonce>>24),
		byte(nonce>>32), byte(nonce>>40), byte(nonce>>48), byte(nonce>>56))
	LHash = append(LHash, hash...)
	lHash := sha256.Sum256(LHash)
	LHash = lHash[:]

	var state uint64
	// We assume the hash provided is from a good cryptographic hash function, something like Sha256.
	// Initialize state with the first 8 bytes of the given hash.
	for i := 0; i < 8; i++ {
		state = state<<8 ^ uint64(LHash[i])
	}

	// Make a number of passes through the hash.  Note that we make complete passes over the hash,
	// from the least significant byte to the most significant byte.  This ensures that we have to go
	// through the complete process prior to knowing the PoW value.
	// 
	
	for i := 0; i < lx.Loops; i++ {
		for j := len(LHash) - 1; j >= 0; j-- {
			state = state<<17 ^ state>>7 ^ uint64(lx.ByteMap[state&mask]^LHash[j])
			LHash[j] = byte(state)
		}
	}
	return lx.pow(LHash)
}

// Return a uint64 as the difficulty.  
//   - Byte 0 is the Count of leading bytes of "FF"
//   - Byte 1-7 the last 7 bytes of the hash
//
// What this does is allow difficulty to be computed from the hash in 8 bytes
//
// Larger values represent larger difficulties.
func (lx LxrPow) pow(hash []byte) (pow uint64) {
	idx := uint64(0) // idx is the index of bytes to collect into the difficulty
	for hash[idx] == 0xff && idx < uint64(len(hash)) {
		idx++
	}
	pow = uint64(idx) // The first byte is the count of leading FF's
	end := len(hash)-8
	for i := 0; i < 7; i++ { // Add the last 7 bytes of the hash
		pow = pow << 8
		pow += uint64(hash[end])
		end++
	}
	return pow // Return the pow 
}
