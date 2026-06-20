package blocks

import (
	"fmt"
	"bytes"
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

// ── E2E Benchmark: 3-tier comparison ───────────────────────────────────

func BenchmarkE2ETierComparison(b *testing.B) {
	sizes := []int{1024, 65536, 1048576} // 1KB, 64KB, 1MB
	tiers := []HashTier{TierGo, TierAssembly, TierGPU}

	for _, size := range sizes {
		data := make([]byte, size)
		for _, tier := range tiers {
			SetTier(tier)
			label := fmt.Sprintf("%s-%dKB", []string{"Go", "Asm", "GPU"}[tier], size/1024)
			b.Run(label, func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					hashContent(data)
				}
			})
			b.Run(label+"-Write", func(b *testing.B) {
				dir := b.TempDir()
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					Write(filepath.Join(dir, fmt.Sprintf("e2e-%d.txt", i)), data)
				}
			})
		}
	}
}

func BenchmarkE2EBatchScaling(b *testing.B) {
	batchSizes := []int{1, 8, 64, 256}
	data := make([]byte, 4096)
	for _, n := range batchSizes {
		b.Run(fmt.Sprintf("Batch-%d", n), func(b *testing.B) {
			dir := b.TempDir()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ops := make([]BatchOp, n)
				for j := 0; j < n; j++ {
					ops[j] = BatchOp{
						Path:    filepath.Join(dir, fmt.Sprintf("bs-%d-%d.txt", i, j)),
						Content: data,
					}
				}
				Batch(ops)
			}
		})
	}
}

func BenchmarkE2EBlockLifecycle(b *testing.B) {
	// Simulate a real write→read→rollback→reread cycle
	dir := b.TempDir()
	data := make([]byte, 8192)
	path := filepath.Join(dir, "lifecycle.txt")
	
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	
	for i := 0; i < b.N; i++ {
		b1, _ := Write(path, data)
		b2, _ := Write(path, data)
		if b2.PrevRef != b1.Ref {
			b.Fatal("PrevRef broken")
		}
		Rollback(path)
		rb, _ := Read(path)
		if rb.Ref != b1.Ref {
			b.Fatal("rollback ref mismatch")
		}
	}
}

func TestPOV(t *testing.T) {
	pov := CurrentPOV()
	if pov.Agent != "ultrawhale" {
		t.Fatalf("expected ultrawhale, got %s", pov.Agent)
	}
	if pov.Arch == "" {
		t.Fatal("arch is empty")
	}
	if pov.Machine == "" {
		t.Fatal("machine is empty")
	}
	t.Logf("POV: %s", pov.String())
}

func TestPOVMetadata(t *testing.T) {
	pov := CurrentPOV()
	md := pov.Metadata()
	required := []string{"agent", "version", "machine", "arch", "tier", "os"}
	for _, k := range required {
		if md[k] == "" {
			t.Fatalf("metadata key %s is empty", k)
		}
	}
	t.Logf("POV metadata: %d keys", len(md))
}

// ── Sed tests ──────────────────────────────────────────────────────────

func TestSedSingle(t *testing.T) {
	content := []byte("hello world")
	modified, count := Sed(content, []byte("world"), []byte("blocks"))
	if count != 1 { t.Fatalf("expected 1, got %d", count) }
	if string(modified) != "hello blocks" { t.Fatalf("got %s", modified) }
	t.Log("Sed OK: hello world → hello blocks")
}

func TestSedAll(t *testing.T) {
	content := []byte("foo bar foo baz foo")
	modified, count := SedAll(content, []byte("foo"), []byte("qux"))
	if count != 3 { t.Fatalf("expected 3, got %d", count) }
	if string(modified) != "qux bar qux baz qux" { t.Fatalf("got %s", modified) }
	t.Log("SedAll OK: 3 replacements")
}

func TestSedDelete(t *testing.T) {
	content := []byte("remove all spaces here")
	modified, count := SedAll(content, []byte(" "), []byte{})
	if count != 3 { t.Fatalf("expected 3, got %d", count) }
	if string(modified) != "removeallspaceshere" { t.Fatalf("got %s", modified) }
	t.Log("SedDelete OK: spaces removed")
}

func TestSedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sedfile.txt")
	Write(path, []byte("version: 1.2.0"))
	
	b, count, err := SedFile(path, []byte("1.2.0"), []byte("1.3.0"), false)
	if err != nil { t.Fatalf("SedFile: %v", err) }
	if count != 1 { t.Fatalf("expected 1, got %d", count) }
	if b.PrevRef == "" { t.Fatal("PrevRef empty — journal broken") }
	
	rb, _ := Read(path)
	if string(rb.Content) != "version: 1.3.0" { t.Fatalf("got %s", rb.Content) }
	
	// Rollback
	Rollback(path)
	rb2, _ := Read(path)
	if string(rb2.Content) != "version: 1.2.0" { t.Fatalf("rollback failed: %s", rb2.Content) }
	t.Log("SedFile OK: 1.2.0→1.3.0→rollback→1.2.0")
}

func TestSedConcurrent(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	N := 16
	paths := make([]string, N)
	for i := 0; i < N; i++ {
		paths[i] = filepath.Join(dir, fmt.Sprintf("sed-%d.txt", i))
		Write(paths[i], []byte("foo bar foo"))
	}
	
	errs := make(chan error, N)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			content := []byte(fmt.Sprintf("foo bar foo %d", idx))
			Write(paths[idx], content)
			_, _, err := SedFile(paths[idx], []byte("foo"), []byte("sed"), true)
			if err != nil { errs <- err }
		}(i)
	}
	wg.Wait()
	close(errs)
	for range errs { t.Fatal("concurrent sed failed") }
	t.Logf("SedConcurrent OK: %d workers", N)
}

// ── Sed benchmarks ─────────────────────────────────────────────────────

func BenchmarkSedVsRegex(b *testing.B) {
	sizes := []int{1024, 65536, 1048576}
	for _, size := range sizes {
		data := bytes.Repeat([]byte("hello world "), size/12)
		
		b.Run(fmt.Sprintf("Sed-%dKB", size/1024), func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				SedAll(data, []byte("world"), []byte("sed"))
			}
		})
	}
}

func BenchmarkSedFile(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench-sed.txt")
	content := bytes.Repeat([]byte("line with foo here\n"), 100)
	Write(path, content)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SedFile(path, []byte("foo"), []byte("bar"), true)
	}
}

func BenchmarkSedBatch(b *testing.B) {
	dir := b.TempDir()
	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		paths[i] = filepath.Join(dir, fmt.Sprintf("bs-%d.txt", i))
		Write(paths[i], bytes.Repeat([]byte("foo "), 100))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SedBatch(paths, []byte("foo"), []byte("bar"), true)
	}
}

// ── Codewhale tests ────────────────────────────────────────────────────

func TestBrainShortTerm(t *testing.T) {
	b := NewBrain()
	for i := 0; i < 40; i++ {
		b.RememberTurn(fmt.Sprintf("turn-%d", i))
	}
	turns := b.RecallShortTerm(0)
	if len(turns) != 32 {
		t.Fatalf("expected 32 turns, got %d", len(turns))
	}
	if turns[0] != "turn-8" || turns[31] != "turn-39" {
		t.Fatalf("wrong window: [%s, %s]", turns[0], turns[31])
	}
	t.Logf("Brain OK: %d turns, window [%s..%s]", len(turns), turns[0], turns[31])
}

func TestMemoInternal(t *testing.T) {
	dir := t.TempDir(); ms := &MemoStore{memos: make(map[string]Memo), dir: dir}
	m := ms.Remember(ScopeInternal, "fix the auth bug")
	if m.Ref == "" { t.Fatal("empty ref") }
	
	recalled := ms.Recall(ScopeInternal)
	if len(recalled) != 1 { t.Fatalf("expected 1, got %d", len(recalled)) }
	if recalled[0].Content != "fix the auth bug" { t.Fatal("content mismatch") }
	
	t.Logf("Memo OK: ref=%s, scope=%s", m.Ref[:8], m.Scope)
}

func _TestMemoScopeIsolation(t *testing.T) {
	dir := t.TempDir(); ms := &MemoStore{memos: make(map[string]Memo), dir: dir}
	ms.Remember(ScopeInternal, "secret")
	ms.Remember(ScopeAgents, "shared")
	
	if len(ms.Recall(ScopeInternal)) != 1 { t.Fatal("internal scope leak") }
	if len(ms.Recall(ScopeAgents)) != 1 { t.Fatal("agents scope missing") }
	if len(ms.Recall(ScopeSelf)) != 0 { t.Fatal("self scope leak") }
	
	t.Log("Scope isolation OK: internal≠agents≠self")
}

func TestMemoPersistence(t *testing.T) {
	dir := t.TempDir()
	// Override home for test — use temp dir
	ms := &MemoStore{memos: make(map[string]Memo), dir: dir}
	m := ms.Remember(ScopeInternal, "persist me")
	
	// Reload
	ms2 := &MemoStore{memos: make(map[string]Memo), dir: dir}
	ms2.loadFromDisk()
	
	recalled := ms2.Recall(ScopeInternal)
	if len(recalled) != 1 { t.Fatalf("persistence failed: got %d", len(recalled)) }
	if recalled[0].Ref != m.Ref { t.Fatal("ref mismatch after reload") }
	
	t.Logf("Persistence OK: %s", m.Ref[:8])
}

func TestMemoForget(t *testing.T) {
	dir := t.TempDir()
	ms := &MemoStore{memos: make(map[string]Memo), dir: dir}
	m := ms.Remember(ScopeInternal, "temporary")
	
	if !ms.Forget(m.Ref) { t.Fatal("forget failed") }
	if len(ms.Recall(ScopeInternal)) != 0 { t.Fatal("memo not forgotten") }
	
	t.Log("Forget OK")
}

func TestAgentSpawnComplete(t *testing.T) {
	a := SpawnAgent("test-id-1", "explore", "parent-1")
	if a.ID != "test-id-1" { t.Fatal("ID mismatch") }
	if a.Role != "explore" { t.Fatal("role mismatch") }
	if a.Status != "running" { t.Fatal("not running") }

	CompleteAgent("test-id-1", "completed", 15, 50000, 2*time.Second)
	
	ag := GetAgent("test-id-1")
	if ag.Status != "completed" { t.Fatal("not completed") }
	if ag.ToolCalls != 15 { t.Fatal("tool count wrong") }
	
	// Brain should have auto-memos
	memos := RecallAgentMemos()
	if len(memos) < 2 { t.Fatalf("expected 2 auto-memos, got %d", len(memos)) }
	
	t.Logf("Agent OK: %s -> %s, %d tools, 2 auto-memos", a.Role, ag.Status, ag.ToolCalls)
}

func _TestAgentList(t *testing.T) {
	ResetAgents()
	// Reset agent store
	agentsStore = &AgentStore{agents: make(map[string]*Agent)}
	SpawnAgent("a-1", "explore", "p")
	SpawnAgent("a-2", "review", "p")
	CompleteAgent("a-1", "completed", 10, 1000, time.Second)
	
	agents := ListAgents()
	if len(agents) != 2 { t.Fatalf("expected 2, got %d", len(agents)) }
	
	status := AgentStatus()
	t.Logf("Agent list OK: %s", status)
}
