package guid

import (
	"io"
	"reflect"
	"testing"
	"time"
)

func void(...interface{}) {}

type testReader struct {
	buff []byte
}

func newTestReader(b [fieldSize]byte) *testReader {
	return &testReader{
		buff: b[:],
	}
}

func (t *testReader) Read(b []byte) (int, error) {
	copy(b, t.buff)
	return len(b), nil
}

func TestGenerator(t *testing.T) {
	type test struct {
		name      string
		fp        int32
		random    io.Reader
		now       func() time.Time
		ic        int32
		dc        int32
		expect    GUID
		expectErr bool
	}

	fp1, fp2, fp3 := int32(123456), int32(654321), int32(maxInt+1)
	ic1, ic2, ic3 := int32(0), int32(1000), int32(maxInt)
	dc1, dc2, dc3 := int32(maxInt-1), int32(1000), int32(0)
	rd1, rd2, rd3 := int32(111111), int32(222222), int32(333333)
	ts1 := int64(1600000000000000000) // 2020-09-13 07:26:40 -0500 EST
	ts2 := int64(1611111111111000000) // 2021-01-19 21:51:51.111 -0500 EST
	ts3 := int64(1633333333333000000) // 2021-10-04 02:42:13.333 -0500 EST

	// quiet!
	void(fp1, fp2, fp3, ic1, ic2, ic3, dc1, dc2, dc3, rd1, rd2, rd3, ts1, ts2, ts3)

	tests := []test{
		{
			name:   "happy path",
			fp:     fp1,
			random: newTestReader([4]byte{1, 1, 1, 1}),
			now: func() time.Time {
				return time.Unix(0, ts1)
			},
			ic: ic1,
			dc: dc1,
			expect: GUID{
				0x6e, 0x77, 0x80, 0x80, 0xf4, 0xf6, 0x90, 0x5d, 0x0, 0x0, 0x80,
				0x89, 0xf, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x83, 0xcd, 0x1, 0xb0,
				0x80, 0x81, 0x1,
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			gen := stdGenerator{
				Fingerprint: tt.fp,
				Random:      tt.random,
				Now:         tt.now,
				IncrCounter: tt.ic,
				DecrCounter: tt.dc,
			}
			g, err := gen.Generate()
			if err != nil {
				if tt.expectErr {
					// nothing to do
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(g, tt.expect) {
				t.Fatalf("expected:\n%s\n     got:\n%s", tt.expect.String(), g.String())
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewRandom()
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewRandom(WithPrefixBytes('f', 'u'))
	}
}
