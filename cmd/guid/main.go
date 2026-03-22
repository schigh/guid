package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schigh/guid"
)

var (
	prefix   string
	times    uint
	sep      string
	serial   bool
	dest     string
	slug     bool
	scan     string
	scanJSON bool
)

const (
	nl     = "--newline--"
	stdout = "--stdout--"
)

func main() {
	flag.StringVar(&prefix, "p", "", "guid prefix")
	flag.UintVar(&times, "n", 1, "number of guids to generate")
	flag.StringVar(&sep, "sep", nl, "separator for multiple guids")
	flag.BoolVar(&serial, "serial", false, "generate all guids serially")
	flag.StringVar(&dest, "o", stdout, "output file")
	flag.BoolVar(&slug, "slug", false, "output a slug instead of a full guid")
	flag.StringVar(&scan, "scan", "", "inspect guid and print parts to console")
	flag.BoolVar(&scanJSON, "json", false, "sets the output of SCAN to json")
	flag.Parse()

	if scan != "" {
		scanGUID(scan, scanJSON)
		return
	}

	var writeTo io.WriteCloser = os.Stdout

	// determine where the guids will go
	if dest != stdout && dest != "" {
		if !filepath.IsAbs(dest) {
			dest = normalizeRelativePath(dest)
		}

		var openErr error
		writeTo, openErr = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if openErr != nil {
			log.Fatalf("open output file failed: %v", openErr)
		}
		defer func(writeTo io.WriteCloser) {
			if err := writeTo.Close(); err != nil {
				log.Fatal(err)
			}
		}(writeTo)
	}

	// must make at least 1
	if times == 0 {
		times = 1
	}

	// check prefix
	if len(prefix) >= 2 {
		if err := guid.SetGlobalPrefixBytes(prefix[0], prefix[1]); err != nil {
			log.Fatalf("invalid prefix: %v", err)
		}
	}

	var guids []guid.GUID

	if serial {
		guids = generateSerially(times)
	} else {
		guids = generateAsync(times)
	}

	guidStrs := make([]string, len(guids))
	for i := range guids {
		if slug {
			guidStrs[i] = guids[i].Slug()
			continue
		}
		guidStrs[i] = guids[i].String()
	}

	if sep == nl {
		sep = "\n"
	}
	out := strings.Join(guidStrs, sep)
	_, wErr := writeTo.Write([]byte(out))
	if wErr != nil {
		log.Fatalf("write error: %v", wErr)
	}
}

func scanGUID(s string, isJSON bool) {
	const (
		green   = "\u001b[0;32m"
		nocolor = "\u001b[0m"
	)
	g, err := guid.ParseString(s)
	if err != nil {
		if isJSON {
			_, _ = fmt.Fprintf(os.Stderr, `{"error":"Parse GUID failed. '%s' is not a valid guid. Only a full guid can be scanned."}`, s)
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Parse GUID failed\n'%s' is not a valid guid\nOnly a full guid can be scanned.\n", s)
		os.Exit(1)
	}

	p1, p2 := g.PrefixBytes()

	if isJSON {
		out := map[string]string{
			"prefix":      string([]byte{p1, p2}),
			"timestamp":   g.Time().Format(time.ANSIC),
			"fingerprint": fmt.Sprintf("%d", g.Fingerprint()),
			"counter":     fmt.Sprintf("%d", g.Counter()),
			"random":      fmt.Sprintf("%d", g.Random()),
		}
		data, _ := json.Marshal(out)
		_, _ = os.Stdout.Write(data)
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "%sPREFIX%s:      %s\n", green, nocolor, string([]byte{p1, p2}))
	_, _ = fmt.Fprintf(os.Stderr, "%sTIMESTAMP%s:   %s\n", green, nocolor, g.Time().Format(time.ANSIC))
	_, _ = fmt.Fprintf(os.Stderr, "%sFINGERPRINT%s: %d\n", green, nocolor, g.Fingerprint())
	_, _ = fmt.Fprintf(os.Stderr, "%sCOUNTER%s:     %d\n", green, nocolor, g.Counter())
	_, _ = fmt.Fprintf(os.Stderr, "%sRANDOM%s:      %d\n", green, nocolor, g.Random())
}

func normalizeRelativePath(in string) string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("normalize relative path failed: %v", err)
	}

	fullPath := filepath.Join(pwd, in)
	out, err := filepath.Abs(fullPath)
	if err != nil {
		log.Fatalf("normalize relative path failed: %v", err)
	}

	return out
}

func generateSerially(n uint) []guid.GUID {
	buffer := make([]guid.GUID, 0, n)
	var i uint
	for i < n {
		buffer = append(buffer, guid.MustNew())
		i++
	}

	return buffer
}

func generateAsync(n uint) []guid.GUID {
	buffer := make([]guid.GUID, n)
	var wg sync.WaitGroup
	for i := uint(0); i < n; i++ {
		wg.Add(1)
		go func(idx uint) {
			defer wg.Done()
			buffer[idx] = guid.MustNew()
		}(i)
	}
	wg.Wait()
	return buffer
}
