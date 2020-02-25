package queue

import "time"

// NewTaskManager Returns a task manager
func NewTaskManager(driver Driver) TaskManager {
	var tm TaskManager
	tm.driver = driver

	return tm
}

// TaskManager Used for clients that want to work with the queue
type TaskManager struct {
	driver Driver
}

// AddTask Add a task to the queue
func (tm *TaskManager) AddTask(taskName string, taskKey string, doAfter time.Time, data map[string]interface{}) error {
	return tm.driver.addTask(taskName, taskKey, doAfter, data)
}
