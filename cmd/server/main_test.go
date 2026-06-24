package main

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeErrorID(t *testing.T) {
	valid := "550e8400-e29b-41d4-a716-446655440000"
	if got := normalizeErrorID(valid); got != valid {
		t.Fatalf("valid uuid should pass through, got %s", got)
	}
	if got := normalizeErrorID(""); got == "" {
		t.Fatal("empty id should be assigned a uuid")
	}
	a, b := normalizeErrorID("custom-id"), normalizeErrorID("custom-id")
	if a != b {
		t.Fatal("non-uuid ids should map deterministically")
	}
	if _, err := uuid.Parse(a); err != nil {
		t.Fatalf("normalized id is not a uuid: %s", a)
	}
}

func TestFingerprintScrubsVolatileData(t *testing.T) {
	f1 := generateFingerprint("boom", scrubStackTrace("main.go:42 panic 0xdeadbeef"))
	f2 := generateFingerprint("boom", scrubStackTrace("main.go:42 panic 0xcafef00d"))
	if f1 != f2 {
		t.Fatal("memory addresses should be scrubbed so fingerprints match")
	}
	f3 := generateFingerprint("different", scrubStackTrace("main.go:42 panic 0xdeadbeef"))
	if f1 == f3 {
		t.Fatal("different messages should produce different fingerprints")
	}
}
