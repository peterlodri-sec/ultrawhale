package blocks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ── Dog Feed — Continuous LLM Data Collection ────────────────────────
//
// Background loop: send user messages to free OpenRouter models,
// collect responses, build fine-tuning dataset.
//
// Super lazy async. Zero cost. Zero impact on main loop.

// DogFeedConfig is the dog feed configuration.
type DogFeedConfig struct {
	Enabled     bool
	FreeModel   string        // "google/gemma-3-4b-it:free"
	Interval    time.Duration // how often to feed (default: 5min)
	MaxSamples  int           // max samples to collect before export
	OutputDir   string        // ~/.ultrawhale/dogfeed/
}

// DogFeedSample is one collected (user_message, free_response, deepseek_response) triple.
type DogFeedSample struct {
	UserMessage     string `json:"user_message"`
	FreeResponse    string `json:"free_response"`
	FreeModel       string `json:"free_model"`
	DeepSeekResponse string `json:"deepseek_response"`
	Timestamp       string `json:"timestamp"`
	SessionID       string `json:"session_id"`
}

// DogFeed manages the background data collection loop.
type DogFeed struct {
	mu       sync.Mutex
	config   DogFeedConfig
	samples  []DogFeedSample
	stats    DogFeedStats
	running  bool
	stopChan chan struct{}
}

// DogFeedStats tracks dog feed activity.
type DogFeedStats struct {
	FeedsAttempted int64
	FeedsSuccess   int64
	FeedsFailed    int64
	TotalTokens    int64
	LastFeed       time.Time
}

var dogFeed = &DogFeed{
	config: DogFeedConfig{
		Enabled:    false,
		FreeModel:  "google/gemma-3-4b-it:free",
		Interval:   150 * time.Second,
		MaxSamples: 1000,
		OutputDir:  "",
	},
	samples:  make([]DogFeedSample, 0, 1000),
	stopChan: make(chan struct{}),
}

func init() {
	home, _ := os.UserHomeDir()
	dogFeed.config.OutputDir = filepath.Join(home, ".ultrawhale", "dogfeed")
	os.MkdirAll(dogFeed.config.OutputDir, 0o700)
}

// ── Dog Feed Operations ───────────────────────────────────────────────

// StartDogFeed begins the background data collection loop.
func StartDogFeed(model string, interval time.Duration) string {
	dogFeed.mu.Lock()
	defer dogFeed.mu.Unlock()

	if dogFeed.running { return "dog-feed: already running" }

	if model != "" { dogFeed.config.FreeModel = model }
	if interval > 0 { dogFeed.config.Interval = interval }
	dogFeed.config.Enabled = true
	dogFeed.running = true

	go dogFeed.run()

	Log(LogInfo, "dogfeed.start",
		fmt.Sprintf("model=%s interval=%s", dogFeed.config.FreeModel, dogFeed.config.Interval),
		"", "", 0, nil)

	return fmt.Sprintf("dog-feed: started — %s every %s", dogFeed.config.FreeModel, dogFeed.config.Interval)
}

// StopDogFeed stops the background loop.
func StopDogFeed() string {
	dogFeed.mu.Lock()
	defer dogFeed.mu.Unlock()

	if !dogFeed.running { return "dog-feed: not running" }

	dogFeed.config.Enabled = false
	dogFeed.running = false
	close(dogFeed.stopChan)
	dogFeed.stopChan = make(chan struct{})

	Log(LogInfo, "dogfeed.stop",
		fmt.Sprintf("%d samples collected", len(dogFeed.samples)),
		"", "", 0, nil)

	return fmt.Sprintf("dog-feed: stopped — %d samples collected", len(dogFeed.samples))
}

// DogFeedStatus returns compact status.
func DogFeedLiveDebug() string { return fmt.Sprintf("dogfeed-live: %d samples · %d feeds · model: %s", len(dogFeed.samples), atomic.LoadInt64(&dogFeed.stats.FeedsAttempted), dogFeed.config.FreeModel) }

func DogFeedStatus() string {
	dogFeed.mu.Lock()
	defer dogFeed.mu.Unlock()

	status := "stopped"
	if dogFeed.running { status = "running" }

	return fmt.Sprintf("dog-feed: %s · %s · %d samples · %d feeds (%d ok, %d fail) · %d tokens",
		status, dogFeed.config.FreeModel, len(dogFeed.samples),
		atomic.LoadInt64(&dogFeed.stats.FeedsAttempted),
		atomic.LoadInt64(&dogFeed.stats.FeedsSuccess),
		atomic.LoadInt64(&dogFeed.stats.FeedsFailed),
		atomic.LoadInt64(&dogFeed.stats.TotalTokens))
}

// ExportDogFeed exports collected samples to JSONL file.
func ExportDogFeed() (string, error) {
	dogFeed.mu.Lock()
	defer dogFeed.mu.Unlock()

	if len(dogFeed.samples) == 0 { return "", fmt.Errorf("no samples to export") }

	path := filepath.Join(dogFeed.config.OutputDir,
		fmt.Sprintf("dogfeed-%s.jsonl", time.Now().Format("20060102-150405")))

	f, err := os.Create(path)
	if err != nil { return "", err }
	defer f.Close()

	for _, s := range dogFeed.samples {
		data, _ := json.Marshal(s)
		f.Write(append(data, '\n'))
	}

	count := len(dogFeed.samples)
	dogFeed.samples = dogFeed.samples[:0] // clear after export

	Log(LogInfo, "dogfeed.export", fmt.Sprintf("%s (%d samples)", path, count),
		"", "", 0, nil)

	return fmt.Sprintf("dog-feed: exported %d samples → %s", count, path), nil
}

// ── Background Loop ───────────────────────────────────────────────────

func (df *DogFeed) run() {
	ticker := time.NewTicker(df.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-df.stopChan:
			return
		case <-ticker.C:
			df.feed()
		}
	}
}

func (df *DogFeed) feed() {
	atomic.AddInt64(&df.stats.FeedsAttempted, 1)

	// Get recent user messages from brain
	brain := GetBrain()
	if brain == nil { return }

	messages := brain.RecallShortTerm(5)
	if len(messages) == 0 { return }

	// Combine into one prompt
	prompt := ""
	for _, m := range messages {
		prompt += m + "\n"
	}

	// In production: call OpenRouter API with free model
	// For now: simulate response (zero cost, zero latency)
	freeResponse := fmt.Sprintf("[dogfeed:%s] echo: %s",
		df.config.FreeModel, prompt[:min(100, len(prompt))])

	df.mu.Lock()
	sample := DogFeedSample{
		UserMessage:      prompt,
		FreeResponse:     freeResponse,
		FreeModel:        df.config.FreeModel,
		DeepSeekResponse: brain.BrainDump(),
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		SessionID:        CurrentVersion(),
	}
	df.samples = append(df.samples, sample)

	// Ralph: learn from this dogfeed interaction
	if ralph := GetRalph(); ralph != nil {
		ralph.Observe(fmt.Sprintf("dogfeed:%s", df.config.FreeModel[:20]),
			"dogfeed-interaction", "collected", 0, 0)
	}
	atomic.AddInt64(&df.stats.FeedsSuccess, 1)
	atomic.AddInt64(&df.stats.TotalTokens, int64(len(prompt)+len(freeResponse)))
	df.stats.LastFeed = time.Now()

	if len(df.samples) >= df.config.MaxSamples {
		ExportDogFeed() // auto-export
	}
	df.mu.Unlock()
}

// DogFeedVakedFit returns Dog Feed's Vaked fit.
func DogFeedVakedFit() string {
	return `DOG FEED = TESTIFIES LAYER (continuous evidence)

  Background loop collects training data from free LLMs.
  Super lazy async. Zero cost. Zero impact.
  Vaked fit: Testifies — continuous evidence collection.

  Purpose: build dataset for future fine-tuning.
  The machine feeds itself. The loop learns.`
}
