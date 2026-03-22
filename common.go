package guid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

var globalPrefix atomic.Value // stores [2]byte

func init() {
	globalPrefix.Store([2]byte{'i', 'd'})
}

var (
	fpOnce     sync.Once
	fp         int32
	prefixOnce sync.Once

	// TestGUID is a nonsense GUID used for testing.
	TestGUID = GUID{
		0x74, 0x65, // prefix
		0xa8, 0xd9, 0xac, 0xde, 0xb2, 0x83, 0x1, 0x0, // ts
		0xda, 0xc0, 0xa7, 0x1, // fp
		0xc8, 0xd3, 0x4, 0x0, // counter
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // rnd
	}
)

// SetGlobalPrefixBytes is a global initializer for GUID prefixes.
// The default prefix bytes are 'i' and 'd'. This function can only
// be called once successfully per execution. Subsequent calls are no-ops.
// Prefix bytes must be lowercase base36 characters (0-9, a-z).
// Invalid bytes return an error without consuming the one-shot,
// so a subsequent call with valid bytes will still succeed.
func SetGlobalPrefixBytes(b1, b2 byte) error {
	if !(isValidPrefixByte(b1) && isValidPrefixByte(b2)) {
		return fmt.Errorf("guid.SetGlobalPrefixBytes: prefix bytes must be base36-compatible and lowercase")
	}
	prefixOnce.Do(func() {
		globalPrefix.Store([2]byte{b1, b2})
	})
	return nil
}

// MustSetGlobalPrefixBytes calls SetGlobalPrefixBytes and panics on error.
func MustSetGlobalPrefixBytes(b1, b2 byte) {
	if err := SetGlobalPrefixBytes(b1, b2); err != nil {
		panic(err)
	}
}

// filter out of band integers.  The integers produced by the
// default generator will always be within band
func filter(v int32) int64 {
	if v < 0 {
		v = -v
	}
	if v >= maxInt {
		return int64(v % maxInt)
	}
	return int64(v)
}

// filterRandom clamps random values to the valid range [0, maxRandom)
func filterRandom(v int64) int64 {
	if v < 0 {
		v = -v
	}
	if v >= maxRandom {
		return v % maxRandom
	}
	return v
}

// generate a random int64 from crypto/rand bytes
func randomInt64(reader io.Reader) (int64, error) {
	var b [8]byte
	_, err := reader.Read(b[:])
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b[:]) % uint64(maxRandom)), nil
}

// pad a byte slice with the zero value until
// it is the required size
func leftPad(in string, size int) string {
	l := len(in)
	// if the size of the input gte size or size
	// is invalid, just return the input
	if l >= size || size <= 0 {
		return in
	}
	b := append(bytes.Repeat([]byte{'0'}, size-l), []byte(in)...)

	return string(b)
}

// get the default hostname of the device
func defaultHostname() int32 {
	h, err := os.Hostname()
	if err != nil {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		h = string(b)
	}
	hb := []byte(h)
	final := len(hb) + 36
	for _, b := range hb {
		final = final + int(b)
	}
	return int32(final)
}

// get the current process id
func defaultPid() int32 {
	return int32(os.Getpid())
}

// get the default fingerprint of the device
func defaultFingerprint() int32 {
	fpOnce.Do(func() {
		fp = defaultPid()<<2 | defaultHostname()>>2
	})
	return fp
}

// prefix bytes must be printable base36 ASCII chars
func isValidPrefixByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z')
}
