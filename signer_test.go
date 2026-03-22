package guid

import (
	"testing"
)

func TestGUID_Sign(t *testing.T) {
	// Sign should return nil for empty input
	g := MustNew()
	if sig := g.Sign(nil); sig != nil {
		t.Fatalf("expected nil for empty input, got %s", sig)
	}
	if sig := g.Sign([]byte{}); sig != nil {
		t.Fatalf("expected nil for empty input, got %s", sig)
	}

	// Sign should produce consistent output
	data := []byte("music television")
	sig1 := g.Sign(data)
	sig2 := g.Sign(data)
	if string(sig1) != string(sig2) {
		t.Fatalf("expected consistent signatures, got %s and %s", sig1, sig2)
	}

	// Sign output should be 64 hex chars (SHA256 = 32 bytes = 64 hex)
	if len(sig1) != 64 {
		t.Fatalf("expected signature length 64, got %d", len(sig1))
	}

	// Different data should produce different signatures
	sig3 := g.Sign([]byte("live mice sit on us"))
	if string(sig1) == string(sig3) {
		t.Fatal("expected different signatures for different data")
	}

	// Different GUIDs should produce different signatures for same data
	g2 := MustNew()
	sig4 := g2.Sign(data)
	if string(sig1) == string(sig4) {
		t.Fatal("expected different signatures from different GUIDs")
	}
}

func TestGUID_DidSign(t *testing.T) {
	g := MustNew()
	data := []byte("test data")
	sig := g.Sign(data)

	// GUID should verify its own signature
	if !g.DidSign(string(sig)) {
		t.Fatal("expected DidSign to return true for own signature")
	}

	// DidSign should return false for invalid hex
	if g.DidSign("not-hex") {
		t.Fatal("expected DidSign to return false for non-hex input")
	}

	// DidSign should return false for wrong-length hex
	if g.DidSign("abcd") {
		t.Fatal("expected DidSign to return false for short hex")
	}

	// DidSign should return false for tampered signature
	if g.DidSign("aaa0000000000000000000000000000000000000000000000000000000000000") {
		t.Fatal("expected DidSign to return false for tampered signature")
	}

	// TestGUID sign/verify round trip
	testSig := TestGUID.Sign([]byte("hello"))
	if !TestGUID.DidSign(string(testSig)) {
		t.Fatal("TestGUID failed to verify its own signature")
	}
}
