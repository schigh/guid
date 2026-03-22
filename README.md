# GUID

**G**lobally **U**nique **Id**entifiers

A zero-dependency Go library for generating globally unique, collision-resistant, introspectable 28-character identifiers. Similar in purpose to UUID v7, but with a compact base36 encoding and a richer internal structure.

## Structure

A GUID is a 28-byte array that serializes to a 28-character base36 string:

```
[prefix 2][timestamp 8][fingerprint 4][counter 4][random 10]
```

| Field       | Size     | Description                                      |
|-------------|----------|--------------------------------------------------|
| Prefix      | 2 chars  | Application-defined tag (default: `id`)           |
| Timestamp   | 8 chars  | Millisecond-precision UTC time                    |
| Fingerprint | 4 chars  | Derived from PID and hostname                     |
| Counter     | 4 chars  | Monotonic counter (wraps at 36^4)                 |
| Random      | 10 chars | Cryptographic randomness from `crypto/rand`       |

Collision resistance comes from combining all five fields. Within a single process on a single machine, the counter and ~51.7 bits of randomness (36^10 possible values) make collisions extremely unlikely even at high throughput.

## Install

Library:

```
go get github.com/schigh/guid
```

CLI:

```
go install github.com/schigh/guid/cmd/guid@latest
```

## Library Usage

### Generating

`New` returns a GUID and an error. `MustNew` is a convenience wrapper that panics on error instead (safe in practice, since `crypto/rand` rarely fails).

```go
package main

import (
	"fmt"
	"log"

	"github.com/schigh/guid"
)

func main() {
	// returns an error
	g, err := guid.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(g)
	// output: idlen38z4r2w1r0000rq9az8y8xv

	// panics on error (convenient when failure is not expected)
	g2 := guid.MustNew()
	fmt.Println(g2)
}
```

### Parsing

Parse a GUID from a string or byte slice. `ParseString` calls `Parse` internally.

```go
s := "idlen38z4r2w1r0000rq9az8y8xv"
g, err := guid.ParseString(s)
if err != nil {
	log.Fatal(err)
}
fmt.Println(g) // idlen38z4r2w1r0000rq9az8y8xv
```

### Introspection

Every component of a GUID can be extracted:

```go
g := guid.MustNew()

b1, b2 := g.PrefixBytes()
fmt.Printf("Prefix:      %c%c\n", b1, b2)

fmt.Printf("Timestamp:   %v\n", g.Time())
fmt.Printf("Fingerprint: %d\n", g.Fingerprint())
fmt.Printf("Counter:     %d\n", g.Counter())
fmt.Printf("Random:      %d\n", g.Random())
```

### Slugs

A slug is a lossy 12-character abbreviation of a GUID, useful as a short disambiguation key in URLs or small documents. The original GUID cannot be recovered from a slug.

```go
g := guid.MustNew()
fmt.Println(g.Slug()) // e.g. "8z4r00y8xv3q"
```

### Prefix Customization

The default prefix bytes are `i` and `d`. You can change them globally (once, at startup) or per GUID.

**Global prefix (called once, subsequent calls are no-ops):**

```go
err := guid.SetGlobalPrefixBytes('u', 's')
if err != nil {
	log.Fatal(err)
}
g := guid.MustNew()
fmt.Println(g) // us...
```

Prefix bytes must be lowercase base36 characters (`0-9`, `a-z`). `SetGlobalPrefixBytes` returns an error for invalid bytes. `MustSetGlobalPrefixBytes` panics instead.

**Per-GUID prefix using an option:**

```go
g := guid.MustNew(guid.WithPrefixBytes('a', 'b'))
fmt.Println(g) // ab...
```

**Per-GUID prefix by direct assignment:**

```go
g := guid.MustNew()
g[0], g[1] = 'a', 'b'
```

### Watermarking

A GUID can watermark data by folding its bytes into a SHA256 hash. This is not cryptographic signing. It is a lightweight tracing mechanism for associating a GUID with a piece of data.

```go
g := guid.MustNew()
data := []byte("some payload")

wm := g.Watermark(data)
fmt.Printf("Watermark: %s\n", wm) // hex-encoded

// check
if g.HasWatermark(string(wm)) {
	fmt.Println("verified")
}
```

Note: `Watermark` uses bitwise OR to fold GUID bytes into the hash. Multiple GUIDs can appear to match the same watermarked data (false positives are possible). This is suitable for tagging and tracing, not for authentication.

### Custom Generator

The library uses a global `Generator` for all GUID creation. You can replace it once at startup for testing or to inject custom time/randomness sources.

```go
type Generator interface {
	Generate() (GUID, error)
}
```

```go
guid.SetGlobalGenerator(myCustomGenerator{})
```

Subsequent calls to `SetGlobalGenerator` are no-ops. For testing, you can use `guid.TestGUID` as a fixed value.

### Serialization

GUID implements the following standard interfaces:

| Interface                    | Method            |
|------------------------------|-------------------|
| `fmt.Stringer`               | `String()`        |
| `json.Marshaler`             | `MarshalJSON()`   |
| `json.Unmarshaler`           | `UnmarshalJSON()` |
| `sql.Scanner`                | `Scan()`          |
| `driver.Valuer`              | `Value()`         |
| `encoding.BinaryMarshaler`   | `MarshalBinary()` |
| `encoding.BinaryUnmarshaler` | `UnmarshalBinary()` |
| `encoding.TextMarshaler`     | `MarshalText()`   |
| `encoding.TextUnmarshaler`   | `UnmarshalText()` |
| `gob.GobEncoder`             | `GobEncode()`     |
| `gob.GobDecoder`             | `GobDecode()`     |

This means GUIDs work out of the box with `encoding/json`, `database/sql`, `encoding/gob`, and any system that uses the standard marshaling interfaces.

## CLI Usage

```
go install github.com/schigh/guid/cmd/guid@latest
```

### Generate GUIDs

```shell
# generate one GUID
$ guid
idlen38z4r2w1r0000rq9az8y8xv

# generate 5 GUIDs
$ guid -n 5

# generate with a custom prefix
$ guid -p ab

# generate slugs instead of full GUIDs
$ guid -slug

# generate serially (default is concurrent)
$ guid -serial

# use a custom separator
$ guid -n 3 -sep ","

# write output to a file
$ guid -n 10 -o guids.txt
```

### Inspect a GUID

```shell
$ guid -scan idlen38z4r2w1r0000rq9az8y8xv
PREFIX:      nw
TIMESTAMP:   Mon Jan  2 15:04:05 2006
FINGERPRINT: 12345
COUNTER:     0
RANDOM:      987654321
```

JSON output:

```shell
$ guid -scan idlen38z4r2w1r0000rq9az8y8xv -json
{"counter":"0","fingerprint":"12345","prefix":"nw","random":"987654321","timestamp":"Mon Jan  2 15:04:05 2006"}
```

### Full Flag Reference

| Flag      | Default       | Description                              |
|-----------|---------------|------------------------------------------|
| `-n`      | `1`           | Number of GUIDs to generate              |
| `-p`      | (none)        | Two-character prefix for generated GUIDs |
| `-sep`    | newline       | Separator between multiple GUIDs         |
| `-serial` | `false`       | Generate GUIDs serially (not concurrent) |
| `-o`      | stdout        | Output file path                         |
| `-slug`   | `false`       | Output 12-character slugs instead        |
| `-scan`   | (none)        | Inspect a GUID and print its components  |
| `-json`   | `false`       | Output scan results as JSON              |

## Thread Safety

GUID generation is safe for concurrent use. The global generator uses a mutex to protect the monotonic counter, and all other fields (timestamp, fingerprint, random) are either goroutine-local or read from `crypto/rand`.

The global prefix and generator are set-once values protected by `sync.Once`.

## License

See [LICENSE](./LICENSE).
