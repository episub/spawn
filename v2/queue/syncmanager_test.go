package queue

import (
	"testing"
	"time"
)

func TestStartStopSyncManager(t *testing.T) {
	// Simply test starting and stopping manager:

	sm := NewSyncManager(drivers[0])

	timeout := make(chan bool, 1)
	success := make(chan bool)

	go func() {
		// Don't want to wait for SM to stop indefinitely
		time.Sleep(2 * time.Second)
		timeout <- true
	}()

	go func() {
		sm.Run()
		success <- true
	}()

	go func() {
		sm.Stop()
	}()

	select {
	case <-success:
		return
	case <-timeout:
		t.Error("Timeout before sync manager stopped")
	}

}

func TestRegisterAction(t *testing.T) {
	taskName := "TestRegisterAction"
	sm := NewSyncManager(drivers[0])

	go func() {
		sm.Run()
	}()
	defer sm.Stop()

	ea := NewExampleTaskAction(nil)

	err := sm.RegisterTaskHandler(&ea, taskName)

	if err != nil {
		t.Error(err)
		return
	}

	storedAction := sm.getRegisteredAction(taskName)

	if storedAction == nil {
		t.Errorf("Register action returned no error, yet action was not registered")
	}
}

func TestRunScheduledAction(t *testing.T) {
	// Test that scheduled action runs
	sm := NewSyncManager(drivers[0])

	go func() {
		sm.Run()
	}()
	defer sm.Stop()

	results := make(chan bool, 10)
	ea := NewExampleScheduledAction(results)

	sm.Schedule(&ea, time.Millisecond*250)

	started := time.Now()
	// Gather four results:
	count := 0

	for count < 4 {
		<-results
		count++
	}

	diff := time.Since(started)

	if diff.Seconds() < 0.75 || diff.Seconds() > 1.25 {
		t.Errorf("Scheduled time should take around 1 second but took %0.2f", diff.Seconds())
	}
}

func TestRunTaskAction(t *testing.T) {
	// Tests running a task on some queue items.  We run this for each driver, since they may implement queue differently
	for _, driver := range drivers {
		taskName := "TestRunTaskAction"
		sm := NewSyncManager(driver)
		sm.driver.clear()
		tm := NewTaskManager(driver)

		go func() {
			sm.Run()
		}()
		defer sm.Stop()

		result := make(chan bool)
		ea := NewExampleTaskAction(result)

		err := sm.RegisterTaskHandler(&ea, taskName)

		if err != nil {
			t.Error(err)
			return
		}

		// Create our sample task
		taskKey := hashKey(taskName + "1")
		data := map[string]interface{}{}

		err = tm.AddTask(taskName, taskKey, time.Now(), data)

		if err != nil {
			t.Error(err)
		}

		timeout := make(chan bool, 1)

		go func() {
			// Don't want to wait for SM to stop indefinitely
			time.Sleep(4 * time.Second)
			timeout <- true
		}()

		select {
		case <-result:
			break
		case <-timeout:
			t.Error("Timeout before task action was run")
		}

		// Now, if we run 'pop', there should be no waiting tasks, waiting a moment for the thread to write the state:
		time.Sleep(time.Millisecond * 250)
		task, err := sm.driver.pop()

		if err != ErrNoTasks {
			t.Errorf("Should have had ErrNoTasks, but had %+v: %+v", err, task)
		}

	}
}
