package queue

import "time"

// SyncClient Client for interacting with the queue
type SyncClient struct {
	driver Driver
}

// NewSyncClient Returns a properly configured sync client
func NewSyncClient(driver Driver) SyncClient {
	return SyncClient{driver: driver}
}

// AddTask Adds a task to the queue
func (s *SyncClient) AddTask(taskName string, taskKey string, doAfter time.Time, data map[string]interface{}) error {
	return s.driver.addTask(taskName, taskKey, doAfter, data)
}
