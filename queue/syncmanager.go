package queue

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// NewSyncManager Returns a new and ready sync manager
func NewSyncManager(driver Driver) SyncManager {
	var sm SyncManager
	sm.driver = driver
	sm.registeredActions = make(map[string]TaskAction)
	sm.actionQueue = make(chan (ScheduledAction))
	sm.taskQueue = make(chan (taskQueueAction))
	sm.cancel = make(chan (bool))
	sm.registerMutex = &sync.Mutex{}

	sm.errorHandler = defaultErrorHandler

	return sm
}

type taskQueueAction struct {
	Task Task
	Done chan bool
}

// SyncManager is the central process for running actions
type SyncManager struct {
	actionQueue       chan (ScheduledAction)
	taskQueue         chan (taskQueueAction)
	cancel            chan (bool)
	driver            Driver
	registeredActions map[string]TaskAction
	registerMutex     *sync.Mutex
	errorHandler      func(error)
}

// Run Runs the main loop that keeps the queue running and performs actions at specified intervals
func (s *SyncManager) Run() {

	// Start the synchroniser queue handler:
	go s.runQueue()

	for {
		select {
		case action := <-s.actionQueue:
			err := action.Do()
			if err != nil {
				s.errorHandler(err)
			}
		case <-s.cancel:
			return
		case tqa := <-s.taskQueue:
			var err error
			task := tqa.Task
			action := s.getRegisteredAction(task.Name)

			if action == nil {
				err = fmt.Errorf("Cancelling task with ID %s because there is no action to handle it", task.id)
				s.errorHandler(err)
				err = s.driver.cancel(task.id, err.Error())
				if err != nil {
					s.errorHandler(err)
				}
			} else {
				result, message := action.Do(task)
				switch result {
				case TaskResultPermanentFailure, TaskResultRetryFailure:
					// Task failed
					s.errorHandler(fmt.Errorf("%s", message))

					switch result {
					case TaskResultPermanentFailure:
						err = s.driver.fail(task.id, message)
					case TaskResultRetryFailure:
						err = s.driver.retry(task.id, message)
					default:
						err = fmt.Errorf("Undefined task result %s", result)
					}

					if err != nil {
						s.errorHandler(err)
					}
				case TaskResultSuccess:
					// Complete the task
					err = s.driver.complete(task.id, message)
					if err != nil {
						s.errorHandler(err)
					}
				default:
					s.errorHandler(fmt.Errorf("Fell through.  Undefined task result %s", result))
				}
			}

			tqa.Done <- true
		}
	}
}

func (s *SyncManager) runQueue() {

	refreshDelay := time.Second * 4 // refreshDelay defines how soon before refreshing tasks that need to be retried
	refreshed := time.Now()

	for {
		// Refresh tasks marked for retry:
		if time.Now().Sub(refreshed) >= refreshDelay {
			err := s.driver.refreshRetry(time.Hour)

			if err != nil {
				s.errorHandler(err)
			}

			refreshed = time.Now()
		}

		// Check for new tasks in queue:
		task, err := s.driver.pop()

		if err != nil && err != ErrNoTasks {
			s.errorHandler(err)
		} else if err != ErrNoTasks {
			// We want to wait until this is executed before we begin the task again.  Otherwise "pop" might return the same value, since it's not truly pop'ing

			reply := make(chan bool)
			s.taskQueue <- taskQueueAction{Task: task, Done: reply}
			<-reply
		}

		time.Sleep(1 * time.Second)
	}
}

// Stop Stops the sync manager main loop
func (s *SyncManager) Stop() {
	s.cancel <- true
}

// Schedule Schedule an action to be performed at particular intervals
func (s *SyncManager) Schedule(act ScheduledAction, period time.Duration) {
	ticker := time.NewTicker(period)

	go func(act ScheduledAction, ticker *time.Ticker) {
		for {
			<-ticker.C

			s.actionQueue <- act
		}
	}(act, ticker)
}

// RegisterTaskHandler Specifies which action to be used to handle a task of name taskName
func (s *SyncManager) RegisterTaskHandler(act TaskAction, taskName string) error {
	s.registerMutex.Lock()
	s.registeredActions[taskName] = act
	s.registerMutex.Unlock()

	return nil
}

func (s *SyncManager) getRegisteredAction(taskName string) TaskAction {
	var taskAction TaskAction

	s.registerMutex.Lock()
	taskAction = s.registeredActions[taskName]
	s.registerMutex.Unlock()

	return taskAction
}

// SetErrorHandler Sets a function to handle errors from the run function
func (s *SyncManager) SetErrorHandler(handler func(err error)) {
	s.errorHandler = handler
}

func defaultErrorHandler(err error) {
	log.Print(err)
}
