package blocks

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// ── Chaos Tests ───────────────────────────────────────────────────────

func TestChaosConcurrentWriteRollbackRead(t *testing.T) {
	if testing.Short() { t.Skip("chaos test") }

	dir := t.TempDir()
	path := filepath.Join(dir, "chaos.txt")
	Write(path, []byte("initial"))

	var wg sync.WaitGroup
	errs := make(chan error, 100)
	workers := 16
	duration := 2 * time.Second

	deadline := time.After(duration)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-deadline:
					return
				default:
					content := []byte(fmt.Sprintf("w%d-%d", id, rand.Intn(10000)))
					if b, err := Write(path, content); err == nil {
						_ = b.Ref
						time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
						Rollback(path)
					}
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	failures := 0
	for range errs { failures++ }
	if failures > workers*2 {
		t.Fatalf("%d failures in chaos test", failures)
	}

	// Final read should succeed
	if _, err := Read(path); err != nil {
		t.Fatalf("final read after chaos: %v", err)
	}

	t.Logf("Chaos OK: %d workers, %s, %d failures", workers, duration.Round(time.Second), failures)
}
