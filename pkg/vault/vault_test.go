package vault_test

import (
	"path/filepath"
	"testing"

	apex "github.com/Segniko/Apex/proto"
	"github.com/Segniko/Apex/pkg/vault"
)

func TestVaultEncryptRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // exactly 32 bytes
	v, err := vault.New(filepath.Join(t.TempDir(), "test.db"), key)
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	defer v.Close()

	in := &apex.CrashReport{
		ErrorId: "e1",
		Message: "sensitive-crash-data",
		Context: &apex.DeviceContext{Os: "linux", Arch: "amd64"},
	}
	if err := v.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}

	out, err := v.FetchAll()
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(out) != 1 || out[0].Message != "sensitive-crash-data" {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}

func TestVaultRejectsBadKey(t *testing.T) {
	if _, err := vault.New(filepath.Join(t.TempDir(), "x.db"), []byte("too-short")); err == nil {
		t.Fatal("expected error for non-32-byte key")
	}
}
