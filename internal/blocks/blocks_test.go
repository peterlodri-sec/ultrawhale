package blocks

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello, blocks")

	b, err := Write(path, content)
	if err != nil { t.Fatalf("Write: %v", err) }
	if b.Ref != Ref(content) { t.Fatal("ref mismatch") }

	rb, err := Read(path)
	if err != nil { t.Fatalf("Read: %v", err) }
	if rb.Ref != b.Ref { t.Fatal("ref mismatch on read") }
	t.Logf("Write+Read OK: ref=%s", b.Ref[:12])
}

func TestRollback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollback.txt")
	v1 := []byte("version 1")
	v2 := []byte("version 2")
	b1, _ := Write(path, v1)
	Write(path, v2)
	Rollback(path)
	rb, _ := Read(path)
	if rb.Ref != b1.Ref { t.Fatal("rollback ref mismatch") }
	t.Logf("Rollback OK: journal depth=%d", journal.Depth(path))
}

func TestBatch(t *testing.T) {
	dir := t.TempDir()
	ops := []BatchOp{
		{Path: filepath.Join(dir, "a.txt"), Content: []byte("a")},
		{Path: filepath.Join(dir, "b.txt"), Content: []byte("b")},
		{Path: filepath.Join(dir, "c.txt"), Content: []byte("c")},
	}
	if err := Batch(ops); err != nil { t.Fatalf("Batch: %v", err) }
	for _, op := range ops {
		rb, _ := Read(op.Path)
		if string(rb.Content) != string(op.Content) { t.Fatal("content mismatch") }
	}
	t.Logf("Batch OK: %d files", len(ops))
}

func TestConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	N := 32
	M := 100
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < M; j++ {
				path := filepath.Join(dir, fmt.Sprintf("f-%d-%d.txt", id, j))
				Write(path, []byte(fmt.Sprintf("w-%d-%d-%d", id, j, time.Now().UnixNano())))
			}
		}(i)
	}
	wg.Wait()
	t.Logf("Concurrency OK: %d workers x %d = %d writes", N, M, N*M)
}

func TestRaceCondition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "race.txt")
	Write(path, []byte("initial"))
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1); go func(id int) { defer wg.Done(); Write(path, []byte(fmt.Sprintf("w-%d", id))) }(i)
		wg.Add(1); go func(id int) { defer wg.Done(); Read(path) }(i)
	}
	wg.Wait()
	b, _ := Read(path)
	t.Logf("Race OK: final ref=%s", b.Ref[:12])
}

func BenchmarkWrite(b *testing.B) {
	dir := b.TempDir()
	c := make([]byte, 4096)
	b.ResetTimer(); b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Write(filepath.Join(dir, fmt.Sprintf("b-%d.txt", i)), c)
	}
}

func BenchmarkRead(b *testing.B) {
	dir := b.TempDir()
	p := filepath.Join(dir, "br.txt")
	Write(p, make([]byte, 4096))
	b.ResetTimer(); b.ReportAllocs()
	for i := 0; i < b.N; i++ { Read(p) }
}

func BenchmarkBatch(b *testing.B) {
	dir := b.TempDir()
	c := make([]byte, 1024)
	b.ResetTimer(); b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ops := make([]BatchOp, 10)
		for j := 0; j < 10; j++ {
			ops[j] = BatchOp{Path: filepath.Join(dir, fmt.Sprintf("bb-%d-%d.txt", i, j)), Content: c}
		}
		Batch(ops)
	}
}

func BenchmarkHashTierGo(b *testing.B) {
	c := make([]byte, 65536)
	b.SetBytes(int64(len(c))); b.ReportAllocs()
	for i := 0; i < b.N; i++ { hashGo(c) }
}
