package guid

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

		rng := rando.Int31n(maxInt)
		{
			g := GUID{}
			g[0], g[1] = 'x', 'o'

			var s string

			s = g.String()
			// test string after first two bytes
			if s != "xo000000000000000000000000" {
				t.Fatalf("expected 'xo000000000000000000000000', got '%s", s)
			}

			// test string after random is set
			g = g.SetRandom(rng)
			s = g.String()
			if s != "xo000000000000000000006rgq" {
				t.Fatalf("expected 'xo000000000000000000006rgq', got '%s", s)
			}

			// test string after counters are set
			g = g.SetCounters(500, 1000)
			s = g.String()
			if s != "xo00000000000000dw00rs6rgq" {
				t.Fatalf("expected 'xo00000000000000dw00rs6rgq', got '%s", s)
			}

			// test string after time is set
			g = g.SetTime(time.Unix(0, ts))
			s = g.String()
			if s != "xokp8l85n2000000dw00rs6rgq" {
				t.Fatalf("expected 'xokp8l85n2000000dw00rs6rgq', got '%s", s)
			}

			// test string after fingerprint is set
			g = g.SetFingerprint(2222)
			s = g.String()
			if s != "xokp8l85n201pq00dw00rs6rgq" {
				t.Fatalf("expected 'xokp8l85n201pq00dw00rs6rgq', got '%s", s)
			}
		}

		{
			g, err := ParseString("xokp8l85n201pq00dw00rs6rgq")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// get the fingerprint
			fp := g.Fingerprint()
			if fp != int32(2222) {
				t.Fatalf("expected fingerprint value 2222, got %d", fp)
			}

			// get the time
			// we lose sub-millisecond precision on the conversion...see
			// the comment for ts above
			nano := g.Time().UnixNano()
			if nano != ts {
				t.Fatalf("expected timestamp %d, got %d", ts, nano)
			}

			// get counters
			incr, decr := g.Counters()
			if incr != 500 || decr != 1000 {
				t.Fatalf("expected counter values 500 and 1000, got %d and %d", incr, decr)
			}

			// get random
			rd := g.Random()
			if rd != rng {
				t.Fatalf("expected random value %d, got %d", rng, rd)
			}

			// get prefix
			b1, b2 := g.PrefixBytes()
			if b1 != 'x' || b2 != 'o' {
				t.Fatalf("expected prefix bytes %d and %d, got %d and %d", 'x', 'o', b1, b2)
			}

		}
	})

	t.Run("parsing", func(t *testing.T) {
		type test struct {
			name        string
			str         string
			expect      GUID
			expectErr   bool
			errContains string
		}

		tests := []test{
			{
				name: "happy path",
				str:  "xokp8l85n201pq00dw00rs6rgq",
				expect: GUID{
					0x78, 0x6f, 0x9c, 0xce, 0xd5, 0xbf, 0xb6, 0x5e, 0x0, 0x0,
					0xdc, 0x22, 0x0, 0x0, 0xe8, 0x7, 0x0, 0x0, 0xd0, 0xf, 0x0,
					0x0, 0x94, 0xc2, 0x26, 0x0,
				},
			},
			{
				name: "case insensitive after prefix bytes",
				str:  "xoKP8l85N201PQ00DW00RS6RGQ",
				expect: GUID{
					0x78, 0x6f, 0x9c, 0xce, 0xd5, 0xbf, 0xb6, 0x5e, 0x0, 0x0,
					0xdc, 0x22, 0x0, 0x0, 0xe8, 0x7, 0x0, 0x0, 0xd0, 0xf, 0x0,
					0x0, 0x94, 0xc2, 0x26, 0x0,
				},
			},
			{
				name:        "too short",
				str:         "nope",
				expectErr:   true,
				errContains: "bytes in length",
			},
			{
				name:        "too long",
				str:         "xokp8l85n201pq00dw00rs6rgqf",
				expectErr:   true,
				errContains: "bytes in length",
			},
			{
				name:        "monkey business: multibyte fail",
				str:         "xokp8l85n201pq🐵0dw00rs6rgq",
				expectErr:   true,
				errContains: "bytes in length",
			},
			{
				name:        "monkey business: bad time",
				str:         "xokp🐵l85n201pq00dw00rs6",
				expectErr:   true,
				errContains: "invalid time value",
			},
			{
				name:        "monkey business: bad fingerprint",
				str:         "xokp8l85n2🐵00dw00rs6rgq",
				expectErr:   true,
				errContains: "invalid fingerprint value",
			},
			{
				name:        "monkey business: bad increment counter",
				str:         "xokp8l85n201pq🐵0dw00rs6",
				expectErr:   true,
				errContains: "invalid increment counter value",
			},
			{
				name:        "monkey business: bad decrement counter",
				str:         "xokp8l85n201pq00dw00r🐵6",
				expectErr:   true,
				errContains: "invalid decrement counter value",
			},
			{
				name:        "monkey business: bad random",
				str:         "xokp8l85n201pq00dw00rs🐵",
				expectErr:   true,
				errContains: "invalid random value",
			},
		}

		for i := range tests {
			tt := tests[i]
			t.Run(tt.name, func(t *testing.T) {
				out, err := ParseString(tt.str)
				if err != nil {
					if tt.expectErr {
						if !strings.Contains(err.Error(), tt.errContains) {
							t.Fatalf("expected error string [%s] to contain [%s]", err.Error(), tt.errContains)
						}
						// nothing else to do
						return
					}
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(out, tt.expect) {
					t.Fatalf("expected:\n%s\n     got:\n%s", tt.expect.String(), out.String())
				}
			})
		}
	})

	t.Run("options", func(t *testing.T) {
		t.Run("WithPrefixBytes", func(t *testing.T) {
			{
				b1, b2 := byte('4'), byte('2')
				g, err := NewRandom(WithPrefixBytes(b1, b2))
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
				g, err := NewRandom(WithPrefixBytes(b1, b2))
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
				g, err := NewRandom(WithPrefixBytes(b1, b2))
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
				g, err := NewRandom(WithPrefixBytes(b1, b2))
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
	orig := `nwlen32f2p2s1g0001r1dsx5ls
nwlen32f2p2s1g0003r1drxy2f
nwlen32f2p2s1g0004r1dqz6mu
nwlen32f2p2s1g0005r1dpx7zu
nwlen32f2p2s1g0006r1dodre0
nwlen32f2p2s1g0007r1dnj1ji
nwlen32f2p2s1g0008r1dmf0kx
nwlen32f2p2s1g0002r1dlzk8h
nwlen32f2p2s1g0009r1dkfsb5
nwlen32f2p2s1g0000r1djhu8g
shlen330b42s8g0001r1ejl30p
shlen330b42s8g0000r1eisdfc
shlen330b42s8g0003r1ehpuk7
shlen330b42s8g0004r1eg9qwu
shlen330b42s8g0005r1efpukn
shlen330b42s8g0006r1eevyge
shlen330b42s8g0007r1edsabt
shlen330b42s8g0008r1ecpk2m
shlen330b42s8g0002r1ebck46
shlen330b42s8g0009r1eacku7
fulen33m4v2stg0000r1fcwcpv
fulen33m4v2stg0002r1fbc496
fulen33m4v2stg0003r1fa8kkg
fulen33m4v2stg0001r1f9n1zn
fulen33m4v2stg0004r1f8l31x
fulen33m4v2stg0005r1f75ndn
fulen33m4v2stg0006r1f62fo4
fulen33m4v2stg0007r1f5vx7v
fulen33m4v2stg0008r1f4on9j
fulen33m4v2stg0009r1f32foc
xxlen34mdp2ss80000r1gn8hi0
xxlen34mdp2ss80001r1gmddow
xxlen34mdp2ss80002r1glk6px
xxlen34mdp2ss80003r1gk8v0l
xxlen34mdp2ss80004r1gjqbmb
xxlen34mdp2ss80005r1gi18cy
xxlen34mdp2ss80006r1ghkm7p
xxlen34mdp2ss80007r1gg61qv
xxlen34mdp2ss80008r1gfwqzc
xxlen34mdp2ss80009r1gen1nn`
	want := `2f2p01dsx5ls
2f2p03drxy2f
2f2p04dqz6mu
2f2p05dpx7zu
2f2p06dodre0
2f2p07dnj1ji
2f2p08dmf0kx
2f2p02dlzk8h
2f2p09dkfsb5
2f2p00djhu8g
30b401ejl30p
30b400eisdfc
30b403ehpuk7
30b404eg9qwu
30b405efpukn
30b406eevyge
30b407edsabt
30b408ecpk2m
30b402ebck46
30b409eacku7
3m4v00fcwcpv
3m4v02fbc496
3m4v03fa8kkg
3m4v01f9n1zn
3m4v04f8l31x
3m4v05f75ndn
3m4v06f62fo4
3m4v07f5vx7v
3m4v08f4on9j
3m4v09f32foc
4mdp00gn8hi0
4mdp01gmddow
4mdp02glk6px
4mdp03gk8v0l
4mdp04gjqbmb
4mdp05gi18cy
4mdp06ghkm7p
4mdp07gg61qv
4mdp08gfwqzc
4mdp09gen1nn`

	guids := strings.Split(orig, "\n")
	slugs := strings.Split(want, "\n")

	assert.Exactly(t, len(guids), len(slugs))

	for i := range guids {
		g, err := Parse([]byte(guids[i]))
		if err != nil {
			t.Fatal(err)
		}
		slg := slugs[i]
		assert.Equal(t, slg, g.Slug())
	}
}

func BenchmarkParseString(b *testing.B) {
	str := "xokp8l85n201pq00dw00rs6rgq"
	for i := 0; i < b.N; i++ {
		_, _ = ParseString(str)
	}
}

func BenchmarkParse(b *testing.B) {
	bt := []byte("xokp8l85n201pq00dw00rs6rgq")
	for i := 0; i < b.N; i++ {
		_, _ = Parse(bt)
	}
}

func BenchmarkString(b *testing.B) {
	g1, _ := ParseString("nwkuh4byaw2w910000m6rlbqiq")
	g2, _ := ParseString("nwkuh4cowa2wsd0000m6sjr5b2")
	g3, _ := ParseString("abkuh4d5kz2wvx0000m6t5bsmk")

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
