package guid

import (
	"sync"
	"testing"
	"time"
)

type testReader struct {
	buff []byte
}

func newTestReader(b [8]byte) *testReader {
	return &testReader{
		buff: b[:],
	}
}

func (t *testReader) Read(b []byte) (int, error) {
	copy(b, t.buff)
	return len(b), nil
}

func TestGenerator(t *testing.T) {
	ts1 := int64(1600000000000000000) // 2020-09-13 07:26:40 -0500 EST

	gen := stdGenerator{
		Fingerprint: 123456,
		Random:      newTestReader([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		Now: func() time.Time {
			return time.Unix(0, ts1)
		},
		Counter: 0,
	}
	g, err := gen.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify fields round-trip correctly
	if g.Counter() != 0 {
		t.Fatalf("expected counter 0, got %d", g.Counter())
	}
	if g.Fingerprint() != 123456%maxInt {
		t.Fatalf("expected fingerprint %d, got %d", 123456%maxInt, g.Fingerprint())
	}
	nano := g.Time().UnixNano()
	if nano != ts1 {
		t.Fatalf("expected time %d, got %d", ts1, nano)
	}
	// verify counter incremented
	if gen.Counter != 1 {
		t.Fatalf("expected generator counter to be 1, got %d", gen.Counter)
	}
}

func TestConcurrentGeneration(t *testing.T) {
	const goroutines = 100
	const perGoroutine = 100

	results := make(chan GUID, goroutines*perGoroutine)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				g, err := New()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				results <- g
			}
		}()
	}

	wg.Wait()
	close(results)

	seen := make(map[string]struct{}, goroutines*perGoroutine)
	for g := range results {
		s := g.String()
		if len(s) != byteSize {
			t.Fatalf("expected string length %d, got %d: %s", byteSize, len(s), s)
		}
		if _, exists := seen[s]; exists {
			t.Fatalf("duplicate GUID detected: %s", s)
		}
		seen[s] = struct{}{}
	}

	if len(seen) != goroutines*perGoroutine {
		t.Fatalf("expected %d unique GUIDs, got %d", goroutines*perGoroutine, len(seen))
	}
}

func TestCounterWraparound(t *testing.T) {
	gen := &stdGenerator{
		Fingerprint: 42,
		Random:      newTestReader([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		Now: func() time.Time {
			return time.Now().UTC()
		},
		Counter: int32(maxInt - 1),
	}

	// generate with counter at maxInt-1
	g, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if g.Counter() != int32(maxInt-1) {
		t.Fatalf("expected counter %d, got %d", maxInt-1, g.Counter())
	}
	// counter should have wrapped to 0 (maxInt is out of range for filter)
	if gen.Counter != 0 {
		t.Fatalf("expected counter to wrap to 0, got %d", gen.Counter)
	}

	// next generate should use counter 0
	g2, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if g2.Counter() != 0 {
		t.Fatalf("expected counter 0 after wrap, got %d", g2.Counter())
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = New()
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = New(WithPrefixBytes('f', 'u'))

	}
}
