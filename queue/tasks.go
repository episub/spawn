package queue

import "time"

// TaskResult Result code for action
type TaskResult string

// TaskMessage Arbitrary message associated with action
type TaskMessage string

var (
	// TaskResultSuccess Task was a success
	TaskResultSuccess TaskResult = "SUCCESS"
	// TaskResultPermanentFailure Task failed with an error and is not to be retried
	TaskResultPermanentFailure TaskResult = "ERROR"
	// TaskResultRetryFailure Task resulted in an error, but can be retried later
	TaskResultRetryFailure TaskResult = "RETRY"
)

// Task A task to be performed
type Task struct {
	id      string // Optional internal reference for drivers to keep track of where this particular task was retrieved from.
	Key     string // A 'task' is a request to do something.  E.g., synchronise customer y.  The same task may be in the queue multiple times
	Name    string // What type of task is this?  This is used to determine which action will handle the task
	Created time.Time
	State   TaskState
	Data    map[string]interface{} // Storage of information that the action handler can use
}

// TaskState Allowable states for a task
type TaskState string

var (
	// TaskReady Task is ready to be actioned
	TaskReady TaskState = "READY"
	// TaskInProgress Task is currently being processed
	TaskInProgress TaskState = "IN PROGRESS"
	// TaskCancelled Task has been cancelled and will not be actioned
	TaskCancelled TaskState = "CANCELLED"
	// TaskFailed Task has failed and will not be actioned
	TaskFailed TaskState = "FAILED"
	// TaskRetry Task has failed and needs to be retried at a later time
	TaskRetry TaskState = "RETRY"
	// TaskDone Task is completed/finished/done
	TaskDone TaskState = "DONE"
)
