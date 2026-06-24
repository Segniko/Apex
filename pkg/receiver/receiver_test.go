package receiver

import (
	"bytes"
	"strings"
	"testing"

	apex "github.com/Segniko/Apex/proto"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

func zstdCompress(b []byte) []byte {
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	zw.Write(b)
	zw.Close()
	return buf.Bytes()
}

func TestUnpackProtobuf(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatal(err)
	}
	batch := &apex.BatchReport{Reports: []*apex.CrashReport{{ErrorId: "x", Message: "boom"}}}
	raw, _ := proto.Marshal(batch)

	out, err := r.Unpack(zstdCompress(raw))
	if err != nil {
		t.Fatalf("unpack proto: %v", err)
	}
	if len(out.Reports) != 1 || out.Reports[0].Message != "boom" {
		t.Fatalf("unexpected batch: %+v", out)
	}
}

func TestUnpackJSONFallback(t *testing.T) {
	r, _ := New()
	body := []byte(`{"reports":[{"message":"nil pointer dereference"}]}`)

	out, err := r.Unpack(zstdCompress(body))
	if err != nil {
		t.Fatalf("unpack json: %v", err)
	}
	if len(out.Reports) != 1 || out.Reports[0].Message != "nil pointer dereference" {
		t.Fatalf("unexpected json batch: %+v", out)
	}
}

func TestUnpackRejectsGarbage(t *testing.T) {
	r, _ := New()
	if _, err := r.Unpack([]byte("not-zstd")); err == nil {
		t.Fatal("expected error for non-zstd payload")
	}
}

func TestAnalyzeKnownPattern(t *testing.T) {
	r, _ := New()
	got := r.Analyze(&apex.CrashReport{Message: "runtime error: invalid memory address or nil pointer dereference"})
	if !strings.Contains(got, "TACTICAL_FIX") {
		t.Fatalf("expected tactical fix, got: %q", got)
	}
}
