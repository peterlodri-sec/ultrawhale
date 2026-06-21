package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Task Manager — Pure Concurrent Task Queue ─────────────────────────
//
// A minimal, pure task manager for the orchestrator.
//   - Submit tasks → queue → execute via DelegatePrompt
//   - Max concurrency prevents overload
//   - Timeout per task prevents hangs
//   - Cancel mid-flight
//   - Query status by ID

// Task is one orchestrator delegation.
type Task struct {
	ID          string
	Prompt      string
	Agent       string
	Status      string    // "queued", "running", "completed", "failed", "cancelled", "timeout"
	SpawnedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Result      string
	Error       string
	Timeout     time.Duration
}

// TaskManager manages the concurrent task queue.
type TaskManager struct {
	mu          sync.Mutex
	queue       []*Task
	running     map[string]*Task
	history     []*Task
	MaxConcurrent int
	MaxHistory    int
	Stats        TaskManagerStats
}

// TaskManagerStats tracks task activity.
type TaskManagerStats struct {
	Submitted   int64
	Completed   int64
	Failed      int64
	Cancelled   int64
	Timeouted   int64
}

var taskManager = &TaskManager{
	queue:         make([]*Task, 0),
	running:       make(map[string]*Task),
	history:       make([]*Task, 0),
	MaxConcurrent: 5,
	MaxHistory:    128,
}

// Submit adds a task to the queue.
func SubmitTask(prompt, agentRole string, timeout time.Duration) *Task {
	if timeout == 0 { timeout = 5 * time.Minute }

	task := &Task{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Prompt:    prompt,
		Agent:     agentRole,
		Status:    "queued",
		SpawnedAt: time.Now(),
		Timeout:   timeout,
	}

	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	taskManager.queue = append(taskManager.queue, task)
	taskManager.Stats.Submitted++

	Log(LogInfo, "task.submit", fmt.Sprintf("%s → %s (%s)", task.ID[:12], agentRole, prompt[:min(40, len(prompt))]),
		"", "", 0, nil)

	// Try to start if under concurrency limit
	go taskManager.drainQueue()

	return task
}

// Cancel cancels a queued or running task.
func CancelTask(taskID string) error {
	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	// Check queue
	for i, t := range taskManager.queue {
		if t.ID == taskID {
			t.Status = "cancelled"
			t.CompletedAt = time.Now()
			taskManager.queue = append(taskManager.queue[:i], taskManager.queue[i+1:]...)
			taskManager.history = append(taskManager.history, t)
			taskManager.Stats.Cancelled++
			Log(LogInfo, "task.cancel", taskID[:12], "", "", 0, nil)
			return nil
		}
	}

	// Check running
	if t, ok := taskManager.running[taskID]; ok {
		t.Status = "cancelled"
		t.CompletedAt = time.Now()
		delete(taskManager.running, taskID)
		taskManager.history = append(taskManager.history, t)
		taskManager.Stats.Cancelled++
		go taskManager.drainQueue() // free up a slot
		return nil
	}

	return fmt.Errorf("task %s not found", taskID[:12])
}

// TaskStatus returns the status of a task.
func TaskStatus(taskID string) string {
	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	for _, t := range taskManager.queue {
		if t.ID == taskID { return fmt.Sprintf("%s: queued", taskID[:12]) }
	}
	if t, ok := taskManager.running[taskID]; ok {
		elapsed := time.Since(t.StartedAt).Round(time.Second)
		return fmt.Sprintf("%s: running (%s)", taskID[:12], elapsed)
	}
	for _, t := range taskManager.history {
		if t.ID == taskID { return fmt.Sprintf("%s: %s — %s", taskID[:12], t.Status, t.Result[:min(40, len(t.Result))]) }
	}
	return fmt.Sprintf("task %s: not found", taskID[:12])
}

// ListTasks returns all tasks.
func ListTasks() []*Task {
	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	all := make([]*Task, 0, len(taskManager.queue)+len(taskManager.running)+len(taskManager.history))
	all = append(all, taskManager.queue...)
	for _, t := range taskManager.running { all = append(all, t) }
	all = append(all, taskManager.history...)
	return all
}

// drainQueue starts tasks from the queue if under concurrency limit.
func (tm *TaskManager) drainQueue() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for len(tm.running) < tm.MaxConcurrent && len(tm.queue) > 0 {
		task := tm.queue[0]
		tm.queue = tm.queue[1:]

		task.Status = "running"
		task.StartedAt = time.Now()
		tm.running[task.ID] = task

		// Execute via orchestrator
		go tm.executeTask(task)
	}
}

// executeTask runs a task and handles timeout.
func (tm *TaskManager) executeTask(task *Task) {
	orch := GetOrchestrator()
	if orch == nil {
		tm.completeTask(task, "failed", "orchestrator not initialized", "")
		return
	}

	// Execute with timeout
	done := make(chan struct{})
	go func() {
		agentID, agentRole := orch.DelegatePrompt(task.Prompt)
		task.Agent = agentRole
		tm.completeTask(task, "completed", fmt.Sprintf("delegated to %s (%s)", agentRole, agentID[:8]), agentID)
		close(done)
	}()

	select {
	case <-done:
		// completed normally
	case <-time.After(task.Timeout):
		tm.completeTask(task, "timeout",
			fmt.Sprintf("timeout after %s", task.Timeout.Round(time.Second)), "")
		CancelTask(task.ID)
	}
}

// completeTask marks a task as done.
func (tm *TaskManager) completeTask(task *Task, status, result string, details string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task.Status = status
	task.CompletedAt = time.Now()
	task.Result = result

	delete(tm.running, task.ID)
	tm.history = append(tm.history, task)
	if len(tm.history) > tm.MaxHistory { tm.history = tm.history[1:] }

	switch status {
	case "completed": tm.Stats.Completed++
	case "failed": tm.Stats.Failed++
	case "timeout": tm.Stats.Timeouted++
	}

	_ = details

	Log(LogInfo, "task."+status, fmt.Sprintf("%s: %s", task.ID[:12], result),
		"", "", time.Since(task.StartedAt), nil)

	// Drain queue — free slot
	go tm.drainQueue()
}

// TaskManagerStatus returns compact task manager status.
func TaskManagerStatus() string {
	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	return fmt.Sprintf("tasks: %d queued · %d running · %d done (%d ok, %d fail, %d cancel, %d timeout) · max:%d",
		len(taskManager.queue), len(taskManager.running),
		taskManager.Stats.Completed+taskManager.Stats.Failed+taskManager.Stats.Cancelled+taskManager.Stats.Timeouted,
		taskManager.Stats.Completed, taskManager.Stats.Failed,
		taskManager.Stats.Cancelled, taskManager.Stats.Timeouted,
		taskManager.MaxConcurrent)
}

// TaskManagerVakedFit returns task manager Vaked fit.
func TaskManagerVakedFit() string {
	return `TASK MANAGER = SUPERVISES LAYER QUEUE

  Pure, minimal, concurrent.
  Submit → queue → execute via DelegatePrompt.
  Max concurrency prevents overload.
  Timeout per task prevents hangs.
  Cancel mid-flight.
  Query status by ID.

  This is the orchestrator's concurrent proof-of-work.`
}
