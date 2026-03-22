package guid

import (
	"testing"
)

func TestGUID_Watermark(t *testing.T) {
	// Watermark should return nil for empty input
	g := MustNew()
	if wm := g.Watermark(nil); wm != nil {
		t.Fatalf("expected nil for empty input, got %s", wm)
	}
	if wm := g.Watermark([]byte{}); wm != nil {
		t.Fatalf("expected nil for empty input, got %s", wm)
	}

	// Watermark should produce consistent output
	data := []byte("music television")
	wm1 := g.Watermark(data)
	wm2 := g.Watermark(data)
	if string(wm1) != string(wm2) {
		t.Fatalf("expected consistent watermarks, got %s and %s", wm1, wm2)
	}

	// Watermark output should be 64 hex chars (SHA256 = 32 bytes = 64 hex)
	if len(wm1) != 64 {
		t.Fatalf("expected watermark length 64, got %d", len(wm1))
	}

	// Different data should produce different watermarks
	wm3 := g.Watermark([]byte("live mice sit on us"))
	if string(wm1) == string(wm3) {
		t.Fatal("expected different watermarks for different data")
	}

	// Different GUIDs should produce different watermarks for same data
	g2 := MustNew()
	wm4 := g2.Watermark(data)
	if string(wm1) == string(wm4) {
		t.Fatal("expected different watermarks from different GUIDs")
	}
}

func TestGUID_HasWatermark(t *testing.T) {
	g := MustNew()
	data := []byte("test data")
	wm := g.Watermark(data)

	// GUID should verify its own watermark
	if !g.HasWatermark(string(wm)) {
		t.Fatal("expected HasWatermark to return true for own watermark")
	}

	// HasWatermark should return false for invalid hex
	if g.HasWatermark("not-hex") {
		t.Fatal("expected HasWatermark to return false for non-hex input")
	}

	// HasWatermark should return false for wrong-length hex
	if g.HasWatermark("abcd") {
		t.Fatal("expected HasWatermark to return false for short hex")
	}

	// HasWatermark should return false for tampered data
	if g.HasWatermark("aaa0000000000000000000000000000000000000000000000000000000000000") {
		t.Fatal("expected HasWatermark to return false for tampered data")
	}

	// TestGUID watermark round trip
	testWm := TestGUID.Watermark([]byte("hello"))
	if !TestGUID.HasWatermark(string(testWm)) {
		t.Fatal("TestGUID failed to verify its own watermark")
	}
}
