package blocks

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// LogLevel classifies log events.
type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// LogEvent is a structured, timestamped event record.
type LogEvent struct {
	ID        string
	Timestamp time.Time
	Level     LogLevel
	Source    string
	Operation string
	Path      string
	Ref       string
	PrevRef   string
	Duration  time.Duration
	Error     string
}

// LogSink receives log events. Implementations: ToastSink, NATSSink, FileSink.
type LogSink interface {
	Emit(LogEvent)
}

// Logger is the global event logger. Lock-free ring buffer.
type Logger struct {
	overflows int64 // ring buffer overflow counter
	mu      sync.RWMutex
	buffer  []LogEvent // ring buffer, capacity 4096
	head    int64 // atomic — lock-free writes
	count   int64 // atomic — best-effort tracking
	sinks   []LogSink
}

var globalLogger = &Logger{
	buffer: make([]LogEvent, 4096),
}

// SetSinks replaces the active log sinks.
func SetSinks(sinks ...LogSink) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.sinks = sinks
}

// AddSink appends a log sink.
func AddSink(sink LogSink) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.sinks = append(globalLogger.sinks, sink)
}

// Log records an event to the ring buffer and all sinks.
func Log(level LogLevel, operation, path, ref, prevRef string, duration time.Duration, err error) {
	event := LogEvent{
		ID:        Ref([]byte(fmt.Sprintf("%s:%s:%s:%d", level, operation, path, time.Now().UnixNano())))[:12],
		Timestamp: time.Now(),
		Level:     level,
		Source:    "blocks",
		Operation: operation,
		Path:      path,
		Ref:       ref,
		PrevRef:   prevRef,
		Duration:  duration,
	}
	if err != nil {
		event.Error = err.Error()
	}

	// Lock-free ring buffer write via atomic CAS
	head := atomic.LoadInt64(&globalLogger.head)
	for {
		next := (head + 1) % int64(len(globalLogger.buffer))
		if atomic.CompareAndSwapInt64(&globalLogger.head, head, next) {
			globalLogger.buffer[head] = event
			break
		}
		head = atomic.LoadInt64(&globalLogger.head)
	}
	// Count tracking (best-effort, ok to be slightly off)
	count := atomic.AddInt64(&globalLogger.count, 1)
	if count > int64(len(globalLogger.buffer)) {
		atomic.StoreInt64(&globalLogger.count, int64(len(globalLogger.buffer)))
		atomic.AddInt64(&globalLogger.overflows, 1)
	}
	
	globalLogger.mu.RLock()
	sinks := globalLogger.sinks
	globalLogger.mu.RUnlock()

	// Fan-out to sinks (async-safe — each sink is responsible for its own goroutine safety)
	for _, s := range sinks {
		go s.Emit(event)
	}
}

// Recent returns the last N log events.
func Recent(n int) []LogEvent {
	globalLogger.mu.RLock()
	defer globalLogger.mu.RUnlock()

	count := int(atomic.LoadInt64(&globalLogger.count))
	if n > count {
		n = count
	}
	out := make([]LogEvent, n)
	for i := 0; i < n; i++ {
		h := int(atomic.LoadInt64(&globalLogger.head))
		if h < 0 { h += len(globalLogger.buffer) * 2 }
		idx := (h - n + i + len(globalLogger.buffer)) % len(globalLogger.buffer)
		out[i] = globalLogger.buffer[idx]
	}
	return out
}

// ── ToastSink ──────────────────────────────────────────────────────────
// Renders log events as compact HUD-style toast messages.

type ToastSink struct {
	onEmit func(string) // called with compact status line
}

func NewToastSink(onEmit func(string)) *ToastSink {
	return &ToastSink{onEmit: onEmit}
}

	// POV suffix: show machine·arch·tier in toast

func (t *ToastSink) Emit(e LogEvent) {
	if t.onEmit == nil || e.Level == LogDebug {
		return
	}
	icon := map[LogLevel]string{LogInfo: "·", LogWarn: "⚠", LogError: "✗"}[e.Level]
	pov := CurrentPOV()
	msg := fmt.Sprintf("%s %s %s (%s) [%s]", icon, e.Operation, e.Path, e.Duration.Round(time.Millisecond), pov.String())
	if e.Error != "" {
		msg += " " + e.Error
	}
	t.onEmit(msg)
}


func OverflowCount() int64 { return atomic.LoadInt64(&globalLogger.overflows) }
