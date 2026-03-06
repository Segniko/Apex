package receiver

import (
	"bytes"
	"fmt"
	"io"

	apex "github.com/apex/monitor/proto"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

// Receiver identifies a server capable of processing Apex Syphon batches.
type Receiver struct {
	decoder *zstd.Decoder
}

// New creates a new Receiver.
func New() (*Receiver, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	return &Receiver{decoder: dec}, nil
}

// Unpack decompresses and deserializes an Apex batch.
func (r *Receiver) Unpack(compressedData []byte) (*apex.BatchReport, error) {
	// 1. Decompress
	var decompressed bytes.Buffer
	zr, err := zstd.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if _, err := io.Copy(&decompressed, zr); err != nil {
		return nil, err
	}

	// 2. Unmarshal Protobuf
	batch := &apex.BatchReport{}
	if err := proto.Unmarshal(decompressed.Bytes(), batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	return batch, nil
}
