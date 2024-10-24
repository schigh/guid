package guid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGUID_Sign(t *testing.T) {
	type test struct {
		name   string
		gStr   string
		in     []byte
		expect []byte
	}

	tests := []test{
		{
			name:   "tjl77zbrfh43jk00qs4un0qr57",
			gStr:   "tjl77zbrfh43jk00qs4un0qr57",
			in:     []byte(`music television`),
			expect: []byte(`bfbbe69bfa65feffe9fb171ff76e246ba1fe5c17bf8f9f93bad39b35f6af9f7d`),
		},
		{
			name:   "tjl77zbrfh43jk00qs4un0qr57 v2",
			gStr:   "tjl77zbrfh43jk00qs4un0qr57",
			in:     []byte(`live mice sit on us`),
			expect: []byte(`feffe6dbfee3126df3bf1fa476ff4c59c252d9378d6fbab1fed19f9dffaf9ab1`),
		},
		{
			name:   "l6l77zbrfh43jk00rf4umdycxx",
			gStr:   "l6l77zbrfh43jk00rf4umdycxx",
			in:     []byte(`music television`),
			expect: []byte(`bfbbe69bfa65feffe9fb171fff76246ba1fe5c17bf8f9f93aad39b35eaddcf7d`),
		},
		{
			name:   "l6l77zbrfh43jk00rf4umdycxx v2",
			gStr:   "l6l77zbrfh43jk00rf4umdycxx",
			in:     []byte(`live mice sit on us`),
			expect: []byte(`feffe6dbfee3126df3bf1fa46eb74c59c252d937b76fbab1ded19f9dbbfddbb1`),
		},
		{
			name:   "test guid",
			gStr:   TestGUID.String(),
			in:     []byte(`music television`),
			expect: []byte(`bdfbaedfbae7ffffdbd0a71ff767246ba1fe5c17ffd39f93eeb7af35e087af7d`),
		},
		{
			name:   "test guid v2",
			gStr:   TestGUID.String(),
			in:     []byte(`live mice sit on us`),
			expect: []byte(`ecffeedfb6a3136dfbd6bfa576f74c59c252d937cdffbeb1d6f5b79dbba7beb1`),
		},
		{
			name:   "single char",
			gStr:   "tjl77zbrfh43jk00qs4un0qr57",
			in:     []byte(`x`),
			expect: []byte(`bffbd6d3ff66b044e1eb7fa9ffee32f5c8530fb1983fc4dbbaf59f17d6bfd881`),
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			g, err := ParseString(tt.gStr)
			if err != nil {
				t.Fatal(err)
			}

			sig := g.Sign(tt.in)
			fmt.Println(string(sig))
			assert.Equal(t, tt.expect, sig)
		})
	}
}

func TestGUID_DidSign(t *testing.T) {
	type test struct {
		name           string
		gStr           string
		input          string
		shouldBeSigned bool
	}

	tests := []test{
		{
			name:           "tjl77zbrfh43jk00qs4un0qr57",
			gStr:           "tjl77zbrfh43jk00qs4un0qr57",
			shouldBeSigned: true,
			input:          `bfbbe69bfa65feffe9fb171ff76e246ba1fe5c17bf8f9f93bad39b35f6af9f7d`,
		},
		{
			name:           "l6l77zbrfh43jk00rf4umdycxx",
			gStr:           "l6l77zbrfh43jk00rf4umdycxx",
			shouldBeSigned: true,
			input:          `bfbbe69bfa65feffe9fb171fff76246ba1fe5c17bf8f9f93aad39b35eaddcf7d`,
		},
		{
			name:           "test guid 1",
			gStr:           TestGUID.String(),
			shouldBeSigned: true,
			input:          `bdfbaedfbae7ffffdbd0a71ff767246ba1fe5c17ffd39f93eeb7af35e087af7d`,
		},
		{
			name:           "test guid 2",
			gStr:           TestGUID.String(),
			shouldBeSigned: true,
			input:          `ecffeedfb6a3136dfbd6bfa576f74c59c252d937cdffbeb1d6f5b79dbba7beb1`,
		},
		{
			name:           "tjl77zbrfh43jk00qs4un0qr57",
			gStr:           "tjl77zbrfh43jk00qs4un0qr57",
			shouldBeSigned: false,
			input:          `bfbbe69bfa65feffe9fb171ff76e246ba1fe5c17bf8f9f93bad39b35f6af9f7c`,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			g, err := ParseString(tt.gStr)
			if err != nil {
				t.Fatal(err)
			}

			signed := g.DidSign(tt.input)
			assert.Equal(t, tt.shouldBeSigned, signed)
		})
	}
}

/*
UNCOMMENT THIS TO MAKE YOU SOME TEST GUIDS
*/
/*
func Test_MkGuids(*testing.T) {
	// This is going to generate N guids as quickly as possible. Since each guid
	// is created in its own goroutine, these guids should be lexically close
	// together...e.g. this is how you make a ton of guids that can be parsed by
	// a binary tree efficiently
	rand.Seed(time.Now().UnixNano())
	r := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		p1 := r[rand.Intn(len(r))]
		p2 := r[rand.Intn(len(r))]

		go func(b1, b2 byte, wg *sync.WaitGroup) {
			defer wg.Done()
			g := New(WithPrefixBytes(b1, b2))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", g)
		}(p1, p2, &wg)
	}

	wg.Wait()
}
*/
