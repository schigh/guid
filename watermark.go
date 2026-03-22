package guid

import (
	"crypto/sha256"
	"encoding/hex"
)

// Watermark folds the GUID's bytes into a SHA256 hash of the input data.
// This is not cryptographic signing. It is a lightweight tracing mechanism
// for associating a GUID with a piece of data.
func (g GUID) Watermark(in []byte) []byte {
	// cant digest what we don't have
	if len(in) == 0 {
		return nil
	}

	// get a hash of the input
	sum := sha256.Sum256(in) // [32]byte

	// fold all GUID bytes into the hash
	// forward: g[2..14] into sum[0..12]
	// reverse: g[27..15] into sum[31..19]
	j := sha256.Size - 1
	for i := 2; i < 15; i++ {
		sum[i-2] = sum[i-2] | g[i]
		sum[j] = sum[j] | g[j-4]
		j--
	}
	// fold in the prefix
	sum[13] = sum[13] | g[0]
	sum[14] = sum[14] | g[1]

	out := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(out, sum[:])

	return out
}

// HasWatermark returns true when this GUID's bytes are present in the
// given hex-encoded hash. Returns false if the input is not valid hex
// or not SHA256-sized.
func (g GUID) HasWatermark(in string) bool {
	sum, err := hex.DecodeString(in)
	if err != nil {
		return false
	}
	if len(sum) != sha256.Size {
		return false
	}

	if sum[13]&g[0] != g[0] {
		return false
	}
	if sum[14]&g[1] != g[1] {
		return false
	}
	j := sha256.Size - 1
	for i := 2; i < 15; i++ {
		if sum[i-2]&g[i] != g[i] {
			return false
		}
		if sum[j]&g[j-4] != g[j-4] {
			return false
		}
		j--
	}

	return true
}
