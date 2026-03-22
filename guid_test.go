package guid

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

// the name here is funky because TestGUID is a global convenience var
func TestGUIDX(t *testing.T) {
	const (
		// note that this nanosecond timestamp is millisecond
		// precision...we lose that when we convert between
		// int64 and time
		ts = 1622222222222000000 // 2021-05-28 12:17:02.222 -0500 EST
	)
	var (
		rando = rand.New(rand.NewSource(ts))
	)

	t.Run("basic build up and tear down", func(t *testing.T) {
		// this test covers most of the guid internals...it exists
		// mainly as an aid to using the GUID itself

		rng := int64(rando.Int31n(maxInt))
		{
			g := GUID{}
			g[0], g[1] = 'x', 'o'

			var s string

			s = g.String()
			// test string after first two bytes
			expect := "xo" + "00000000" + "0000" + "0000" + "0000000000"
			if s != expect {
				t.Fatalf("expected '%s', got '%s'", expect, s)
			}

			// test string after random is set
			g = g.SetRandom(rng)
			s = g.String()
			t.Logf("after random(%d): %s", rng, s)

			// test string after counter is set
			g = g.SetCounter(500)
			s = g.String()
			t.Logf("after counter(500): %s", s)

			// test string after time is set
			g = g.SetTime(time.Unix(0, ts))
			s = g.String()
			t.Logf("after time: %s", s)

			// test string after fingerprint is set
			g = g.SetFingerprint(2222)
			s = g.String()
			t.Logf("after fingerprint(2222): %s", s)

			// round-trip: parse the string back
			roundTrip := s
			g2, err := ParseString(roundTrip)
			if err != nil {
				t.Fatalf("unexpected error parsing round trip: %v", err)
			}
			if g2.String() != roundTrip {
				t.Fatalf("round trip failed: expected '%s', got '%s'", roundTrip, g2.String())
			}

			// verify each field
			fp := g2.Fingerprint()
			if fp != int32(2222) {
				t.Fatalf("expected fingerprint value 2222, got %d", fp)
			}

			nano := g2.Time().UnixNano()
			if nano != ts {
				t.Fatalf("expected timestamp %d, got %d", ts, nano)
			}

			counter := g2.Counter()
			if counter != 500 {
				t.Fatalf("expected counter value 500, got %d", counter)
			}

			rd := g2.Random()
			if rd != rng {
				t.Fatalf("expected random value %d, got %d", rng, rd)
			}

			b1, b2 := g2.PrefixBytes()
			if b1 != 'x' || b2 != 'o' {
				t.Fatalf("expected prefix bytes %d and %d, got %d and %d", 'x', 'o', b1, b2)
			}
		}
	})

	t.Run("parsing", func(t *testing.T) {
		type test struct {
			name        string
			str         string
			expectErr   bool
			errContains string
		}

		tests := []test{
			{
				name:        "too short",
				str:         "nope",
				expectErr:   true,
				errContains: "bytes in length",
			},
			{
				name:        "too long",
				str:         "xo000000000000000000000000001",
				expectErr:   true,
				errContains: "bytes in length",
			},
			{
				// pr(2) + ts(8) + fp(4) + ctr(4) + rnd(10) = 28
				name:        "monkey business: bad time",
				str:         "xo!0000000000000000000000000000"[:byteSize],
				expectErr:   true,
				errContains: "invalid time value",
			},
			{
				name:        "monkey business: bad fingerprint",
				str:         "xo00000000!000000000000000000000"[:byteSize],
				expectErr:   true,
				errContains: "invalid fingerprint value",
			},
			{
				name:        "monkey business: bad counter",
				str:         "xo00000000000000!00000000000000"[:byteSize],
				expectErr:   true,
				errContains: "invalid counter value",
			},
			{
				name:        "monkey business: bad random",
				str:         "xo000000000000000000!000000000"[:byteSize],
				expectErr:   true,
				errContains: "invalid random value",
			},
		}

		for i := range tests {
			tt := tests[i]
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseString(tt.str)
				if err != nil {
					if tt.expectErr {
						if !strings.Contains(err.Error(), tt.errContains) {
							t.Fatalf("expected error string [%s] to contain [%s]", err.Error(), tt.errContains)
						}
						return
					}
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.expectErr {
					t.Fatal("expected error but got none")
				}
			})
		}
	})

	t.Run("options", func(t *testing.T) {
		t.Run("WithPrefixBytes", func(t *testing.T) {
			{
				b1, b2 := byte('4'), byte('2')
				g, err := New(WithPrefixBytes(b1, b2))
				if err != nil {
					t.Fatal(err)
				}
				e1, e2 := g.PrefixBytes()
				if e1 != b1 || e2 != b2 {
					t.Fatalf("expected '%s' '%s', got '%s', '%s'", string(b1), string(b2), string(e1), string(e2))
				}
			}

			{
				b1, b2 := byte('f'), byte('u')
				g, err := New(WithPrefixBytes(b1, b2))
				if err != nil {
					t.Fatal(err)
				}
				e1, e2 := g.PrefixBytes()
				if e1 != b1 || e2 != b2 {
					t.Fatalf("expected '%s' '%s', got '%s', '%s'", string(b1), string(b2), string(e1), string(e2))
				}
			}

			{
				b1, b2 := byte('x'), byte('x')
				g, err := New(WithPrefixBytes(b1, b2))
				if err != nil {
					t.Fatal(err)
				}
				e1, e2 := g.PrefixBytes()
				if e1 != b1 || e2 != b2 {
					t.Fatalf("expected '%s' '%s', got '%s', '%s'", string(b1), string(b2), string(e1), string(e2))
				}
			}

			{
				b1, b2 := byte('q'), byte('u')
				g, err := New(WithPrefixBytes(b1, b2))
				if err != nil {
					t.Fatal(err)
				}
				e1, e2 := g.PrefixBytes()
				if e1 != b1 || e2 != b2 {
					t.Fatalf("expected '%s' '%s', got '%s', '%s'", string(b1), string(b2), string(e1), string(e2))
				}
			}
		})
	})
}

func TestSlugs(t *testing.T) {
	// generate GUIDs and verify slug properties
	for i := 0; i < 20; i++ {
		g := MustNew()
		slug := g.Slug()
		if len(slug) != 12 {
			t.Fatalf("expected slug length 12, got %d: %s", len(slug), slug)
		}
		gs := g.String()
		// slug should contain: timestamp[6:10], counter[16:18], random[22:28]
		expected := string([]byte{gs[6], gs[7], gs[8], gs[9], gs[16], gs[17], gs[22], gs[23], gs[24], gs[25], gs[26], gs[27]})
		if slug != expected {
			t.Fatalf("slug mismatch: expected %s, got %s (from guid %s)", expected, slug, gs)
		}
	}

	// slugs from different GUIDs should differ
	g1 := MustNew()
	g2 := MustNew()
	if g1.Slug() == g2.Slug() {
		t.Fatal("expected different slugs for different GUIDs")
	}
}

func BenchmarkParseString(b *testing.B) {
	str := MustNew().String()
	for i := 0; i < b.N; i++ {
		_, _ = ParseString(str)
	}
}

func BenchmarkParse(b *testing.B) {
	bt := []byte(MustNew().String())
	for i := 0; i < b.N; i++ {
		_, _ = Parse(bt)
	}
}

func BenchmarkString(b *testing.B) {
	g1 := MustNew()
	g2 := MustNew()
	g3 := MustNew(WithPrefixBytes('a', 'b'))

	b.Run(g1.String(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = g1.String()
		}
	})

	b.Run(g2.String(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = g2.String()
		}
	})

	b.Run(g3.String(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = g3.String()
		}
	})
}

func FuzzParse(f *testing.F) {
	// seed with valid GUIDs
	f.Add([]byte(MustNew().String()))
	f.Add([]byte(TestGUID.String()))
	f.Add([]byte(MustNew(WithPrefixBytes('a', 'b')).String()))
	// seed with edge cases
	f.Add([]byte(""))
	f.Add([]byte("short"))
	f.Add(make([]byte, byteSize))
	f.Add(make([]byte, 100))

	f.Fuzz(func(t *testing.T, data []byte) {
		g, err := Parse(data)
		if err != nil {
			// parsing failure is expected for arbitrary input
			return
		}
		// if it parsed, the string should be byteSize chars
		s := g.String()
		if len(s) != byteSize {
			t.Fatalf("parsed GUID string length %d, expected %d", len(s), byteSize)
		}
		// and it should round-trip
		g2, err := ParseString(s)
		if err != nil {
			t.Fatalf("round-trip parse failed: %v", err)
		}
		if g2.String() != s {
			t.Fatalf("round-trip mismatch: %s != %s", g2.String(), s)
		}
	})
}

func FuzzParseString(f *testing.F) {
	f.Add(MustNew().String())
	f.Add(TestGUID.String())
	f.Add("")
	f.Add("0000000000000000000000000000")

	f.Fuzz(func(t *testing.T, s string) {
		g, err := ParseString(s)
		if err != nil {
			return
		}
		// round-trip check
		s2 := g.String()
		g2, err := ParseString(s2)
		if err != nil {
			t.Fatalf("round-trip parse failed: %v", err)
		}
		if g2.String() != s2 {
			t.Fatalf("round-trip mismatch: %s != %s", g2.String(), s2)
		}
	})
}
