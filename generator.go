package guid

import (
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	globalGen.Store(Generator(&stdGenerator{
		Random: rand.Reader,
		Now: func() time.Time {
			return time.Now().UTC()
		},
		Fingerprint: defaultFingerprint(),
	}))
}

// Generator defines the contract for generating GUIDs
type Generator interface {
	Generate() (GUID, error)
}

// stdGenerator generates GUIDs
type stdGenerator struct {
	Fingerprint int32
	Random      io.Reader
	Now         func() time.Time
	Counter     int32

	mu sync.Mutex
}

var (
	// globalGenerator is stored in an atomic.Value for safe concurrent access.
	// nolint: gochecknoglobals
	globalGen atomic.Value

	setOnce sync.Once
)

// SetGlobalGenerator allows for the manual assignment of the GUID generator.
// The main usefulness of this function is primarily for testing, but
// this function can also be used to inject custom time and randomness
// providers.
// Note that this function can be called only once per runtime.
// Subsequent calls are no-ops.
func SetGlobalGenerator(g Generator) {
	setOnce.Do(func() {
		globalGen.Store(g)
	})
}

func (g *stdGenerator) randomInt64() (int64, error) {
	return randomInt64(g.Random)
}

// Generate will create a new GUID.
func (g *stdGenerator) Generate() (GUID, error) {
	g.mu.Lock()
	counter := g.Counter
	g.Counter++
	if g.Counter >= maxInt {
		g.Counter = 0
	}
	g.mu.Unlock()

	r, err := g.randomInt64()
	if err != nil {
		return GUID{}, err
	}

	v := (GUID{}).SetTime(g.Now()).SetCounter(counter).SetFingerprint(g.Fingerprint).SetRandom(r)
	// set prefix bytes
	pfx := globalPrefix.Load().([2]byte)
	v[0] = pfx[0]
	v[1] = pfx[1]

	return v, nil
}
