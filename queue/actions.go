package queue

// ScheduledAction A scheduled action to be run
type ScheduledAction interface {
	Do() error      // Perform the action
	Stream() string // Returns the stream (for simultaneous running)
}

// TaskAction An action to perform given some task in the task queue
type TaskAction interface {
	Do(task Task) (TaskResult, string) // Perform the action for the task
}
