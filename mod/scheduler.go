package mod

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ScheduledTask 定时任务
type ScheduledTask struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"caseId"`
	CaseName    string    `json:"caseName"`
	Action      string    `json:"action"` // "start" or "stop"
	ScheduledAt time.Time `json:"scheduledAt"`
	CreatedAt   time.Time `json:"createdAt"`
	Status      string    `json:"status"` // "pending", "completed", "failed", "cancelled"
	Error       string    `json:"error,omitempty"`
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	tasks      map[string]*ScheduledTask
	mu         sync.RWMutex
	stopChan   chan struct{}
	project    *RedcProject
	onExecute  func(caseID string, action string) error
	db         *sql.DB
	dbPath     string
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler(project *RedcProject, dbPath string) *TaskScheduler {
	return &TaskScheduler{
		tasks:    make(map[string]*ScheduledTask),
		stopChan: make(chan struct{}),
		project:  project,
		dbPath:   dbPath,
	}
}

// SetExecuteCallback 设置执行回调
func (s *TaskScheduler) SetExecuteCallback(callback func(string, string) error) {
	s.onExecute = callback
}

// InitDB 初始化数据库
func (s *TaskScheduler) InitDB() error {
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 创建表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduled_tasks (
		id TEXT PRIMARY KEY,
		case_id TEXT NOT NULL,
		case_name TEXT NOT NULL,
		action TEXT NOT NULL,
		scheduled_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		status TEXT NOT NULL,
		error TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_case_id ON scheduled_tasks(case_id);
	CREATE INDEX IF NOT EXISTS idx_status ON scheduled_tasks(status);
	CREATE INDEX IF NOT EXISTS idx_scheduled_at ON scheduled_tasks(scheduled_at);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return fmt.Errorf("创建表失败: %v", err)
	}

	s.db = db

	// 从数据库加载待执行的任务
	if err := s.loadTasksFromDB(); err != nil {
		return fmt.Errorf("加载任务失败: %v", err)
	}

	return nil
}

// loadTasksFromDB 从数据库加载待执行的任务
func (s *TaskScheduler) loadTasksFromDB() error {
	rows, err := s.db.Query(`
		SELECT id, case_id, case_name, action, scheduled_at, created_at, status, error
		FROM scheduled_tasks
		WHERE status = 'pending'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		task := &ScheduledTask{}
		var scheduledAtStr, createdAtStr string
		var errorStr sql.NullString

		err := rows.Scan(
			&task.ID,
			&task.CaseID,
			&task.CaseName,
			&task.Action,
			&scheduledAtStr,
			&createdAtStr,
			&task.Status,
			&errorStr,
		)
		if err != nil {
			continue
		}

		// 解析时间
		task.ScheduledAt, _ = time.Parse(time.RFC3339, scheduledAtStr)
		task.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if errorStr.Valid {
			task.Error = errorStr.String
		}

		s.tasks[task.ID] = task
	}

	return rows.Err()
}

// saveTaskToDB 保存任务到数据库
func (s *TaskScheduler) saveTaskToDB(task *ScheduledTask) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO scheduled_tasks 
		(id, case_id, case_name, action, scheduled_at, created_at, status, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		task.ID,
		task.CaseID,
		task.CaseName,
		task.Action,
		task.ScheduledAt.Format(time.RFC3339),
		task.CreatedAt.Format(time.RFC3339),
		task.Status,
		task.Error,
	)
	return err
}

// updateTaskStatusInDB 更新任务状态到数据库
func (s *TaskScheduler) updateTaskStatusInDB(taskID, status, errorMsg string) error {
	_, err := s.db.Exec(`
		UPDATE scheduled_tasks 
		SET status = ?, error = ?
		WHERE id = ?
	`, status, errorMsg, taskID)
	return err
}

// deleteTaskFromDB 从数据库删除任务
func (s *TaskScheduler) deleteTaskFromDB(taskID string) error {
	_, err := s.db.Exec(`DELETE FROM scheduled_tasks WHERE id = ?`, taskID)
	return err
}

// Start 启动调度器
func (s *TaskScheduler) Start() {
	go s.run()
	// 启动定期清理任务
	go s.periodicCleanup()
}

// Stop 停止调度器
func (s *TaskScheduler) Stop() {
	close(s.stopChan)
	if s.db != nil {
		s.db.Close()
	}
}

// periodicCleanup 定期清理已完成的任务
func (s *TaskScheduler) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.CleanupCompletedTasks()
		}
	}
}

// run 运行调度器主循环
func (s *TaskScheduler) run() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndExecuteTasks()
		}
	}
}

// checkAndExecuteTasks 检查并执行到期的任务
func (s *TaskScheduler) checkAndExecuteTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, task := range s.tasks {
		if task.Status == "pending" && now.After(task.ScheduledAt) {
			// 执行任务
			go s.executeTask(id, task)
		}
	}
}

// executeTask 执行任务
func (s *TaskScheduler) executeTask(id string, task *ScheduledTask) {
	s.mu.Lock()
	task.Status = "executing"
	s.updateTaskStatusInDB(id, "executing", "")
	s.mu.Unlock()

	err := s.onExecute(task.CaseID, task.Action)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		s.updateTaskStatusInDB(id, "failed", err.Error())
	} else {
		task.Status = "completed"
		s.updateTaskStatusInDB(id, "completed", "")
	}
}

// AddTask 添加定时任务
func (s *TaskScheduler) AddTask(caseID, caseName, action string, scheduledAt time.Time) (*ScheduledTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证 action
	if action != "start" && action != "stop" {
		return nil, fmt.Errorf("无效的操作类型: %s", action)
	}

	// 验证时间
	if scheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("计划时间不能早于当前时间")
	}

	// 生成任务 ID
	taskID := fmt.Sprintf("%s-%s-%d", caseID, action, time.Now().Unix())

	// 创建任务
	task := &ScheduledTask{
		ID:          taskID,
		CaseID:      caseID,
		CaseName:    caseName,
		Action:      action,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
		Status:      "pending",
	}

	s.tasks[taskID] = task

	// 保存到数据库
	if err := s.saveTaskToDB(task); err != nil {
		delete(s.tasks, taskID)
		return nil, fmt.Errorf("保存任务到数据库失败: %v", err)
	}

	return task, nil
}

// CancelTask 取消任务
func (s *TaskScheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在")
	}

	if task.Status != "pending" {
		return fmt.Errorf("只能取消待执行的任务")
	}

	task.Status = "cancelled"

	// 更新数据库
	if err := s.updateTaskStatusInDB(taskID, "cancelled", ""); err != nil {
		return fmt.Errorf("更新数据库失败: %v", err)
	}

	return nil
}

// GetTask 获取任务
func (s *TaskScheduler) GetTask(taskID string) (*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在")
	}

	return task, nil
}

// ListTasks 列出所有任务
func (s *TaskScheduler) ListTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// ListTasksByCase 列出指定场景的任务
func (s *TaskScheduler) ListTasksByCase(caseID string) []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0)
	for _, task := range s.tasks {
		if task.CaseID == caseID {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// CleanupCompletedTasks 清理已完成的任务（保留最近24小时的）
func (s *TaskScheduler) CleanupCompletedTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	cutoffStr := cutoff.Format(time.RFC3339)

	// 从数据库删除
	if s.db != nil {
		s.db.Exec(`
			DELETE FROM scheduled_tasks 
			WHERE status IN ('completed', 'failed', 'cancelled') 
			AND created_at < ?
		`, cutoffStr)
	}

	// 从内存删除
	for id, task := range s.tasks {
		if (task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled") &&
			task.CreatedAt.Before(cutoff) {
			delete(s.tasks, id)
		}
	}
}
