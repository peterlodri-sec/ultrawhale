package blocks

import (
	"os"
	"path/filepath"
	"testing"
)

// ── Fuzz Tests ────────────────────────────────────────────────────────

func FuzzWrite(f *testing.F) {
	f.Add([]byte("hello"))
	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fuzz.txt")
		b, err := Write(path, data)
		if err != nil {
			t.Skip() // disk full, etc
		}
		if b.Ref != Ref(data) {
			t.Fatalf("ref mismatch: %s vs %s", b.Ref[:8], Ref(data)[:8])
		}
		// Clean read
		rb, err := Read(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if rb.Ref != b.Ref {
			t.Fatalf("read ref mismatch")
		}
	})
}

func FuzzSed(f *testing.F) {
	f.Add([]byte("hello world"), []byte("world"), []byte("fuzz"))
	f.Fuzz(func(t *testing.T, content, find, replace []byte) {
		if len(find) == 0 { t.Skip() }
		result, count := SedAll(content, find, replace)
		if count > 0 && len(result) == 0 {
			t.Fatal("non-zero count but empty result")
		}
		_ = Ref(result)
	})
}

func FuzzCompress(f *testing.F) {
	f.Add([]byte("hello compress fuzz test"))
	f.Fuzz(func(t *testing.T, data []byte) {
		b, err := Compress(data, "fuzz-test")
		if err != nil {
			t.Skip()
		}
		decompressed, err := Decompress(b.Content)
		if err != nil {
			t.Fatalf("decompress: %v", err)
		}
		if string(decompressed) != string(data) {
			t.Fatal("round-trip mismatch")
		}
	})
}

func FuzzBatch(f *testing.F) {
	f.Add([]byte("a"), []byte("b"), []byte("c"))
	f.Fuzz(func(t *testing.T, a, b, c []byte) {
		dir := t.TempDir()
		ops := []BatchOp{
			{Path: filepath.Join(dir, "a.txt"), Content: a},
			{Path: filepath.Join(dir, "b.txt"), Content: b},
			{Path: filepath.Join(dir, "c.txt"), Content: c},
		}
		err := Batch(ops)
		if err != nil {
			t.Skip()
		}
		for _, op := range ops {
			if _, err := os.Stat(op.Path); err != nil {
				t.Fatalf("batch file missing: %s", op.Path)
			}
		}
	})
}
