package blocks

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"time"
)

// ── Compress Primitive ────────────────────────────────────────────────
// Compress/decompress blocks using zlib (fast, built-in, no external deps).
// Used for log rotation, brain dumps, session exports, bench reports.

// Compress zlib-compresses data and returns a Block with the compressed content.
func Compress(data []byte, label string) (*Block, error) {
	start := time.Now()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("compress write: %w", err)
	}
	w.Close()

	compressed := buf.Bytes()
	b := &Block{
		Content: compressed,
		Ref:     Ref(compressed),
		Kind:    "compress",
		Path:    label,
	}

	ratio := float64(len(data)) / float64(len(compressed))
	Log(LogInfo, "blocks.Compress", label, b.Ref, "", time.Since(start),
		fmt.Errorf("%.1fx compression (%d→%d bytes)", ratio, len(data), len(compressed)))
	return b, nil
}

// Decompress decompresses zlib-compressed data.
func Decompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompress: %w", err)
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("decompress read: %w", err)
	}
	return buf.Bytes(), nil
}

// CompressFile compresses a file and writes the compressed version alongside it.
func CompressFile(path string) (*Block, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Compress(data, path+".zlib")
}

// CompressVakedFit documents this as a pure utility function.
func CompressVakedFit() string { return "PURE UTILITY — no Vaked layer. Performance-critical." }
