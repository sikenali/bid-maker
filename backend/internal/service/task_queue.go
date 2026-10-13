package service

import "sync"
import "time"

// TaskQueue 异步任务队列
type TaskQueue struct {
	mu        sync.Mutex
	tasks     map[string]*Task
}

type Task struct {
	TaskID      string
	SessionID   string
	Type        string // "generate_section", "generate_full", "parse", "outline"
	Status      string // "pending", "running", "completed", "failed"
	Progress    int    // 0-100
	SectionID   string
	Result      map[string]any
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make(map[string]*Task),
	}
}

func (tq *TaskQueue) CreateTask(taskID, sessionID, taskType, sectionID string) *Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	task := &Task{
		TaskID:    taskID,
		SessionID: sessionID,
		Type:      taskType,
		Status:    "pending",
		SectionID: sectionID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	tq.tasks[taskID] = task
	return task
}

func (tq *TaskQueue) GetTask(taskID string) (*Task, bool) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, ok := tq.tasks[taskID]
	return task, ok
}

func (tq *TaskQueue) UpdateStatus(taskID, status string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if task, ok := tq.tasks[taskID]; ok {
		task.Status = status
		task.UpdatedAt = time.Now().UTC()
	}
}

func (tq *TaskQueue) UpdateProgress(taskID string, progress int) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if task, ok := tq.tasks[taskID]; ok {
		task.Progress = progress
		task.UpdatedAt = time.Now().UTC()
	}
}

func (tq *TaskQueue) UpdateResult(taskID string, result map[string]any) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if task, ok := tq.tasks[taskID]; ok {
		task.Result = result
		task.Status = "completed"
		task.UpdatedAt = time.Now().UTC()
	}
}

func (tq *TaskQueue) UpdateError(taskID, errMsg string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if task, ok := tq.tasks[taskID]; ok {
		task.Error = errMsg
		task.Status = "failed"
		task.UpdatedAt = time.Now().UTC()
	}
}
