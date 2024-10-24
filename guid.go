package guid

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// byteSize is the size of a GUID, in bytes
	byteSize = 26

	// blockSize is the size in bits of byte-to-int32 conversions
	blockSize = 64

	// fieldSize is the standard size string for each 32bit integer field.
	fieldSize = 4

	// these alias start/end indexes for the GUID components
	tsStart = 2
	tsEnd   = tsStart + 2*fieldSize
	fpStart = tsEnd
	fpEnd   = fpStart + fieldSize
	icStart = fpEnd
	icEnd   = icStart + fieldSize
	dcStart = icEnd
	dcEnd   = dcStart + fieldSize
	rdStart = dcEnd
	rdEnd   = rdStart + fieldSize

	// base is used for all encoding operations. CUIDs use a base36 encoding of
	// the binary data to generate a string.
	base = 36

	maxInt  = 1679616 // 36^4 or base^fieldSize
	i32Buff = 1048576 // buffer for lower int32 byte sums (for random number generation)
)

// GUID is a globally unique identifier
//
//	prefix  timestamp                 fingerprint   incr          decr          random
//
// [[b, b], [b, b, b, b, b, b, b, b], [b, b, b, b], [b, b, b, b], [b, b, b, b], [b, b, b, b]]
type GUID [byteSize]byte

// Option is a function that allows for mutating a GUID
type Option func(GUID) GUID

// WithPrefixBytes sets the prefix bytes for a single GUID. To set GUID prefix
// bytes globally, use the SetGlobalPrefixBytes function.
func WithPrefixBytes(b1, b2 byte) Option {
	return func(g GUID) GUID {
		g[0] = b1
		g[1] = b2
		return g
	}
}

// New will return only a GUID using the global generator or panic (less safe way, but unlikely to fail)
func New(opts ...Option) GUID {
	g, err := NewRandom(opts...)
	if err != nil {
		panic(err)
	}
	return g
}

// NewRandom will create a GUID using the global generator or return an error (this is the safer way)
func NewRandom(opts ...Option) (GUID, error) {
	out, err := globalGenerator.Generate()
	if err != nil {
		return GUID{}, err
	}

	for i := range opts {
		out = opts[i](out)
	}

	return out, nil
}

// PrefixBytes returns the two GUID prefix
// bytes individually
func (g GUID) PrefixBytes() (byte, byte) {
	return g[0], g[1]
}

// SetTime inserts the unix timestamp into the GUID
func (g GUID) SetTime(t time.Time) GUID {
	_ = binary.PutVarint(g[tsStart:tsEnd], time.Duration(t.UnixNano()).Milliseconds())
	return g
}

// Time returns the timestamp embedded in the GUID
func (g GUID) Time() time.Time {
	msec, _ := binary.Varint(g[tsStart:tsEnd])
	return time.Unix(0, msec*1e6)
}

// SetFingerprint adds the device fingerprint Glyph to the GUID
func (g GUID) SetFingerprint(v int32) GUID {
	_ = binary.PutVarint(g[fpStart:fpEnd], filter(v))
	return g
}

// Fingerprint returns the device fingerprint
func (g GUID) Fingerprint() int32 {
	v, _ := binary.Varint(g[fpStart:fpEnd])
	return int32(v)
}

// SetCounters sets the increment and decrement counters
func (g GUID) SetCounters(incr, decr int32) GUID {
	_ = binary.PutVarint(g[icStart:icEnd], filter(incr))
	_ = binary.PutVarint(g[dcStart:dcEnd], filter(decr))

	return g
}

// Counters returns the incrementer and decrementer counters.
func (g GUID) Counters() (int32, int32) {
	incr, _ := binary.Varint(g[icStart:icEnd])
	decr, _ := binary.Varint(g[dcStart:dcEnd])

	return int32(incr), int32(decr)
}

// SetRandom adds an 8-byte random Word to the GUID
func (g GUID) SetRandom(v int32) GUID {
	_ = binary.PutVarint(g[rdStart:rdEnd], filter(v))
	return g
}

// Random returns the random value encoded in the GUID.
func (g GUID) Random() int32 {
	v, _ := binary.Varint(g[rdStart:rdEnd])
	return int32(v)
}

func (g GUID) String() string {
	nanos, _ := binary.Varint(g[tsStart:tsEnd])
	fingerprint, _ := binary.Varint(g[fpStart:fpEnd])
	incr, _ := binary.Varint(g[icStart:icEnd])
	decr, _ := binary.Varint(g[dcStart:dcEnd])
	random, _ := binary.Varint(g[rdStart:rdEnd])

	sb := strings.Builder{}
	sb.Grow(byteSize)
	sb.Write(g[0:2])
	sb.WriteString(leftPad(strconv.FormatInt(nanos, base), fieldSize*2))
	sb.WriteString(leftPad(strconv.FormatInt(fingerprint, base), fieldSize))
	sb.WriteString(leftPad(strconv.FormatInt(incr, base), fieldSize))
	sb.WriteString(leftPad(strconv.FormatInt(decr, base), fieldSize))
	sb.WriteString(leftPad(strconv.FormatInt(random, base), fieldSize))

	return sb.String()
}

// Slug returns a shortened version of the GUID that may be used as a
// disambiguation key in small documents or in URLs.  Note that this
// is a ONE WAY PROCESS.  Generating a slug is lossy such that the
// original GUID cannot be recreated.
func (g GUID) Slug() string { //nolint:gocritic // complains about pointer semantics
	/*
		To create a slug, we take a regular guid and remove the prefix,
		remove the 32 MSBs (4 bytes) from the time bytes, truncate counters,
		and include the random.
		| PREFIX  | TIMESTAMP       | FP      | INCR    | DECR    | RANDOM  |
		| 0 0     | 0 0 0 0 0 0 0 0 | 0 0 0 0 | 0 0 0 0 | 0 0 0 0 | 0 0 0 0 |
		| PREFIX  | TIMESTAMP       | FP      | INCR    | DECR    | RANDOM  |
		| - -     | - - - - 0 0 0 0 | - - - - | - - 0 0 | - - 0 0 | 0 0 0 0 |
		  1 2       3 4 5 6 7 8 9 10  11121314  15161718  19202122  23242526
		  0 1       2 3 4 5 6 7 8 09  10111213  14151617  18192021  22232425
	*/
	gg := g.String()
	out := [12]byte{
		// TIMESTAMP                INCR            DECR            RANDOM
		gg[6], gg[7], gg[8], gg[9], gg[16], gg[17], gg[20], gg[21], gg[22], gg[23], gg[24], gg[25],
	}
	return string(out[:])
}

// Parse the byte slice into a guid
func Parse(in []byte) (GUID, error) {
	if len(in) != byteSize {
		return GUID{}, fmt.Errorf("guid.Parse: the byte slice must be exactly %d bytes in length", byteSize)
	}
	g := GUID{}
	g[0] = in[0]
	g[1] = in[1]

	t, err := strconv.ParseInt(string(in[tsStart:tsEnd]), base, blockSize)
	if err != nil {
		return GUID{}, fmt.Errorf("guid.Parse: invalid time value '%s': %w", in[tsStart:tsEnd], err)
	}
	g = g.SetTime(time.Unix(0, t*1e6))

	fingerprint, err := strconv.ParseInt(string(in[fpStart:fpEnd]), base, blockSize)
	if err != nil {
		return GUID{}, fmt.Errorf("guid.Parse: invalid fingerprint value '%s': %w", in[fpStart:fpEnd], err)
	}
	g = g.SetFingerprint(int32(fingerprint))

	incr, err := strconv.ParseInt(string(in[icStart:icEnd]), base, blockSize)
	if err != nil {
		return GUID{}, fmt.Errorf("guid.Parse: invalid increment counter value '%s': %w", in[icStart:icEnd], err)
	}
	decr, err := strconv.ParseInt(string(in[dcStart:dcEnd]), base, blockSize)
	if err != nil {
		return GUID{}, fmt.Errorf("guid.Parse: invalid decrement counter value '%s': %w", in[dcStart:dcEnd], err)
	}
	g = g.SetCounters(int32(incr), int32(decr))

	r, err := strconv.ParseInt(string(in[rdStart:rdEnd]), base, blockSize)
	if err != nil {
		return GUID{}, fmt.Errorf("guid.Parse: invalid random value '%s': %w", in[rdStart:rdEnd], err)
	}
	g = g.SetRandom(int32(r))

	return g, nil
}

// ParseString is a convenience func for parsing GUID strings
func ParseString(s string) (GUID, error) {
	return Parse([]byte(s))
}

// interface impls

// MarshalJSON implements json.Marshaler
func (g GUID) MarshalJSON() ([]byte, error) {
	s := g.String()
	sz := byteSize + 2
	b := make([]byte, sz)
	b[0] = '"'
	for i := 0; i < byteSize; i++ {
		b[i+1] = s[i]
	}
	b[sz-1] = '"'

	return b, nil
}

// UnmarshalJSON implements json.Unmarshaler
func (g *GUID) UnmarshalJSON(b []byte) error {
	lb := len(b)
	if lb <= 1 {
		return fmt.Errorf("guid.GUID.UnmarshalJSON: unable to parse bytes: %s", b)
	}
	b = b[1 : lb-1]
	gg, err := Parse(b)
	if err != nil {
		// wrap this error so it can be filtered by the caller if needed
		return fmt.Errorf("guid.GUID.UnmarshalJSON: parse error: %w", err)
	}
	*g = gg

	return nil
}

// Scan implements sql.Scanner
func (g *GUID) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	switch vv := v.(type) {
	case []byte:
		gg, err := Parse(vv)
		if err != nil {
			return fmt.Errorf("guid.GUID.Scan: parse error: %w", err)
		}
		*g = gg
		return nil
	case string:
		gg, err := ParseString(vv)
		if err != nil {
			return fmt.Errorf("guid.GUID.Scan: parse error: %w", err)
		}
		*g = gg
		return nil
	default:
		return fmt.Errorf("guid.GUID.Scan: unable to convert value of type %T", v)
	}
}

// Value implements driver.Valuer
func (g GUID) Value() (driver.Value, error) {
	return g.String(), nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (g GUID) MarshalBinary() (data []byte, err error) {
	data = g[:]
	return
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (g *GUID) UnmarshalBinary(data []byte) error {
	gg, err := Parse(data)
	if err != nil {
		return err
	}
	*g = gg

	return nil
}

// MarshalText implements encoding.TextMarshaler
func (g GUID) MarshalText() (text []byte, err error) {
	text = []byte(g.String())
	return
}

// UnmarshalText implements encoding.TextUnmarshaler
func (g *GUID) UnmarshalText(text []byte) error {
	return g.UnmarshalBinary(text)
}

// GobEncode implements gob.GobEncoder
func (g GUID) GobEncode() ([]byte, error) {
	return g.MarshalBinary()
}

// GobDecode implements gob.GobDecoder
func (g *GUID) GobDecode(data []byte) error {
	return g.UnmarshalBinary(data)
}
