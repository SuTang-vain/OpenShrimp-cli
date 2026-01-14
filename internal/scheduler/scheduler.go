package scheduler

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/models"

	"github.com/robfig/cron/v3"
)

// TaskType defines the type of scheduled task
type TaskType string

const (
	TaskCleanup TaskType = "cleanup"
	TaskBackup  TaskType = "backup"
)

// ScheduledTask represents a scheduled task
type ScheduledTask struct {
	ID         string            `json:"id"`
	Type       TaskType          `json:"type"`
	Schedule   string            `json:"schedule"`
	Enabled    bool              `json:"enabled"`
	Tool       string            `json:"tool,omitempty"`
	LastRun    *time.Time        `json:"last_run,omitempty"`
	LastResult *TaskResult       `json:"last_result,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Output   string        `json:"output,omitempty"`
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	cfg     *config.Config
	cron    *cron.Cron
	tasks   map[string]*ScheduledTask
	mu      sync.RWMutex
	running bool
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[string]*ScheduledTask),
	}
}

// LoadTasks loads scheduled tasks from config
func (s *Scheduler) LoadTasks() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, taskCfg := range s.cfg.Scheduler.Tasks {
		task := &ScheduledTask{
			ID:        id,
			Type:      TaskType(taskCfg.Type),
			Schedule:  taskCfg.Schedule,
			Enabled:   taskCfg.Enabled,
			Tool:      taskCfg.Tool,
			CreatedAt: time.Now(),
		}
		s.tasks[id] = task
	}

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	s.cron.Start()
	log.Println("Scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// AddTask adds a new scheduled task
func (s *Scheduler) AddTask(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate schedule
	if _, err := cron.ParseStandard(task.Schedule); err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	s.tasks[task.ID] = task

	// Add to cron if enabled
	if task.Enabled {
		s.scheduleTask(task)
	}

	return nil
}

// scheduleTask adds a task to the cron scheduler
func (s *Scheduler) scheduleTask(task *ScheduledTask) {
	s.cron.AddFunc(task.Schedule, func() {
		s.RunTask(task)
	})
}

// RunTask executes a scheduled task
func (s *Scheduler) RunTask(task *ScheduledTask) {
	result := &TaskResult{
		StartTime: time.Now(),
	}

	var err error
	var output string

	switch task.Type {
	case TaskCleanup:
		output, err = s.runCleanup(task.Tool)
	case TaskBackup:
		output, err = s.runBackup()
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = output

	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Task failed: %v", err)
		log.Printf("Scheduled task %s failed: %v", task.ID, err)
	} else {
		result.Success = true
		result.Message = "Task completed successfully"
		log.Printf("Scheduled task %s completed", task.ID)
	}

	// Update task result
	s.mu.Lock()
	if t, ok := s.tasks[task.ID]; ok {
		now := time.Now()
		t.LastRun = &now
		t.LastResult = result
	}
	s.mu.Unlock()
}

// runCleanup runs a cleanup task
func (s *Scheduler) runCleanup(tool string) (string, error) {
	if tool == "" {
		// Run cleanup for all tools
		cmd := exec.Command("go", "run", ".", "cleanup", "--json")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), err
		}
		return string(out), nil
	}

	// Run cleanup for specific tool
	cmd := exec.Command("go", "run", ".", "cleanup", tool, "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// runBackup runs a backup task
func (s *Scheduler) runBackup() (string, error) {
	cmd := exec.Command("go", "run", ".", "backup")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// GetTasks returns all scheduled tasks
func (s *Scheduler) GetTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTask returns a specific task
func (s *Scheduler) GetTask(id string) (*ScheduledTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

// UpdateTask updates an existing task
func (s *Scheduler) UpdateTask(id string, enabled bool, schedule string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	if schedule != "" {
		if _, err := cron.ParseStandard(schedule); err != nil {
			return fmt.Errorf("invalid cron schedule: %w", err)
		}
		task.Schedule = schedule
	}

	task.Enabled = enabled

	// If enabling and was disabled, we need to reload scheduler
	// For simplicity, user should restart daemon
	s.tasks[id] = task
	return nil
}

// DeleteTask deletes a task
func (s *Scheduler) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(s.tasks, id)
	return nil
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() models.JSONMap {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasksCount := len(s.tasks)
	enabledCount := 0
	var lastRun *time.Time

	for _, task := range s.tasks {
		if task.Enabled {
			enabledCount++
		}
		if task.LastRun != nil {
			if lastRun == nil || task.LastRun.After(*lastRun) {
				lastRun = task.LastRun
			}
		}
	}

	result := models.JSONMap{
		"total_tasks":   tasksCount,
		"enabled_tasks": enabledCount,
		"running":       s.running,
		"last_run":      nil,
	}

	if lastRun != nil {
		result["last_run"] = lastRun.Format(time.RFC3339)
	}

	return result
}

// ToConfig converts scheduler tasks to config format
func (s *Scheduler) ToConfig() map[string]config.ScheduledTaskConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]config.ScheduledTaskConfig)
	for id, task := range s.tasks {
		result[id] = config.ScheduledTaskConfig{
			Type:     string(task.Type),
			Schedule: task.Schedule,
			Enabled:  task.Enabled,
			Tool:     task.Tool,
		}
	}
	return result
}
