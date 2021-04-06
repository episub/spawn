package queue

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var drivers []Driver

func init() {
	// Add each driver to be tested:
	var err error

	dbConn := os.Getenv("PG_CONNSTRING")
	dbSchema := os.Getenv("PG_SCHEMA")
	dbTable := os.Getenv("PG_TABLE")

	mDriver, err := NewPostgresDriver(dbConn, dbSchema, dbTable)

	if err != nil {
		panic(err)
	}

	drivers = append(drivers, mDriver)
}

func TestClearQueue(t *testing.T) {

	for _, d := range drivers {
		err := d.clear()

		if err != nil {
			t.Error(err)
		}

		// Check queue length:
		if err = checkLength(d, 0); err != nil {
			t.Error(err)
			continue
		}

		err = d.addTask("testClear", "somekey", time.Now(), map[string]interface{}{})

		if err != nil {
			t.Error(err)
			continue
		}

		if err = checkLength(d, 1); err != nil {
			t.Error(err)
			continue
		}

		err = d.clear()

		if err != nil {
			t.Error(err)
		}

		if err = checkLength(d, 0); err != nil {
			t.Error(err)
			continue
		}
	}

}

func checkLength(d Driver, length int64) error {
	fetchedLength, err := d.getQueueLength()

	if err != nil {
		return err
	}

	if length != fetchedLength {
		return fmt.Errorf("Expected length %d, but had %d", length, fetchedLength)
	}

	return nil
}

func TestAddTask(t *testing.T) {

	for _, d := range drivers {
		taskKey := hashKey("something")
		data := map[string]interface{}{
			"exampleString": "val1",
			"exampleBool":   true,
		}
		taskName := "testMSAddTask"

		err := d.addTask(taskName, taskKey, time.Now(), data)

		if err != nil {
			t.Error(err)
		}
	}

}

func TestPop(t *testing.T) {
	taskKey := hashKey("customer_update:123")
	data := map[string]interface{}{
		"exampleString": "val1",
		"exampleBool":   true,
	}

	taskName := "customer_update"

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
		}

		// First check for ErrNoTasks if we pop with no tasks:
		_, err = d.pop()

		if err != ErrNoTasks {
			if err != nil {
				t.Error(err)
			} else {
				t.Error("Expected ErrNoTasks but had no error")
			}
			continue
		}

		// Add a task to test with:
		err = d.addTask(taskName, taskKey, time.Now(), data)

		if err != nil {
			t.Error(err)
		}

		// Pop task should return that task:

		task, err := d.pop()
		if err != nil {
			t.Error(err)
			continue
		}

		if task.Key != taskKey {
			t.Errorf("Key mismatch (%s): task.Key %s, taskKey %s", d.name(), task.Key, taskKey)
		}

		if task.Name != taskName {
			t.Errorf("Name mismatch (%s): task.Name %s, taskName %s", d.name(), task.Name, taskName)
		}

		if !reflect.DeepEqual(data, task.Data) {
			t.Errorf("Data mismatch (%s): task.Data: %+v, data: %+v", d.name(), task.Data, data)
		}
	}

}

func TestCompleteTask(t *testing.T) {

	// Case: items A, B, and C are created in order.  Pop needs to grab oldest items in queue, but *most recent* for any particular taskName.  In this case, C should be popped, and A and B never popped, and C's completion count as a completion for A and B.  This way we avoid updating multiple times with the same data.

	tasks := []Task{
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 1},
		},
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 2},
		},
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 3},
		},
	}

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		// Create tasks:
		for i, task := range tasks {
			log.Printf("Adding %d", i)
			err = d.addTask(task.Name, task.Key, time.Now(), task.Data)

			time.Sleep(100 * time.Millisecond)

			if err != nil {
				t.Error(err)
				continue
			}
		}

		// Pop, contrary to name, should fetch oldest first:

		task, err := d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		if int(task.Data["order"].(float64)) != 1 {
			t.Errorf("Expected task to have 'order' of 1, but was %d", int(task.Data["order"].(float64)))
			continue
		}

		// Mark task as completed:

		err = d.complete(task.id, "None")

		if err != nil {
			t.Error(err)
			continue
		}

		// Pop new task, to check that we have none returned after third:

		for i := 0; i < 2; i++ {
			err = popAndComplete(d)
			if err != nil {
				t.Error(err)
			}
		}

		_, err = d.pop()

		if err != ErrNoTasks {
			if err != nil {
				t.Errorf("Expected ErrNoTasks but had: %s", err)
			} else {
				t.Error("Expected ErrNoTasks, but no error returned")
			}
			continue
		}

	}
}

func popAndComplete(d Driver) error {
	task, err := d.pop()
	if err != nil {
		return err
	}

	return d.complete(task.id, "None")
}

func TestPopNotDoneYet(t *testing.T) {
	// Case 1: item A is created and popped, requesting a customer update.  Customer data is pulled from database and prepared to send to remote system.  In the meantime, item B is added before A's action is completed.  A now completes, and marks that taskName as completed.  The systems are now out of sync because B required an update based on newer data, but A's action completion marked taskName as completed.

	tasks := []Task{
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 1},
		},
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 2},
		},
		{
			Key:  "testCompleteTask1",
			Name: "testCompleteTask",
			Data: map[string]interface{}{"order": 3},
		},
	}

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		// Create the first two tasks:
		for i := 0; i < 2; i++ {
			task := tasks[i]
			err = d.addTask(task.Name, task.Key, time.Now(), task.Data)

			time.Sleep(100 * time.Millisecond)

			if err != nil {
				t.Error(err)
				continue
			}
		}

		// Pop oldest, and it should be the "order"=2 task
		task, err := d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		if int(task.Data["order"].(float64)) != 1 {
			t.Errorf("Expected task to have 'order' of 1, but was %d", int(task.Data["order"].(float64)))
			continue
		}

		// Add the third task:

		err = d.addTask(tasks[2].Name, tasks[2].Key, time.Now(), tasks[2].Data)
		if err != nil {
			t.Error(err)
			continue
		}

		// Mark task as completed:

		err = d.complete(task.id, "None")

		if err != nil {
			t.Error(err)
			continue
		}

		// Pop new task, to check that task with order 2 is returned:

		task, err = d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		if int(task.Data["order"].(float64)) != 2 {
			t.Errorf("Expected task to have 'order' of 2, but was %d", int(task.Data["order"].(float64)))
			continue
		}

	}
}

func TestCancelTask(t *testing.T) {

	// Cancel a task.  A task is cancelled when no action is found for a task

	tasks := []Task{
		{
			Key:  "testCancelTask1",
			Name: "testCancelTask",
			Data: map[string]interface{}{},
		},
	}

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		// Create task:
		for _, task := range tasks {
			err = d.addTask(task.Name, task.Key, time.Now(), task.Data)
			if err != nil {
				t.Error(err)
				continue
			}
		}

		// Pop most recent
		task, err := d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		// Mark task as cancelled:

		err = d.cancel(task.id, "Cancelled")

		if err != nil {
			t.Error(err)
			continue
		}

		// Pop new task, to check that we have none returned:

		task, err = d.pop()

		if err != ErrNoTasks {
			if err != nil {
				t.Errorf("Expected ErrNoTasks but had: %s", err)
			} else {
				t.Error("Expected ErrNoTasks, but no error returned")
			}
			continue
		}

	}
}

func TestFailTask(t *testing.T) {

	// Fail a task.  A task is failed if there was an error when it returned

	tasks := []Task{
		{
			Key:  "testFailTask1",
			Name: "testFailTask",
			Data: map[string]interface{}{},
		},
	}

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		// Create task:
		for _, task := range tasks {
			err = d.addTask(task.Name, task.Key, time.Now(), task.Data)
			if err != nil {
				t.Error(err)
				continue
			}
		}

		// Pop most recent
		task, err := d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		// Mark task as failed:

		err = d.fail(task.id, "Cancelled")

		if err != nil {
			t.Error(err)
			continue
		}

		// Pop new task, to check that we have none returned:

		task, err = d.pop()

		if err != ErrNoTasks {
			if err != nil {
				t.Errorf("Expected ErrNoTasks but had: %s", err)
			} else {
				t.Error("Expected ErrNoTasks, but no error returned")
			}
			continue
		}

	}
}

func TestTaskOrders(t *testing.T) {
	// If task A is created first, then task B, task A should be popped first

	tasks := []Task{
		{
			Key:  "TestTaskOrders1",
			Name: "FakeTaskType",
			Data: map[string]interface{}{"order": 1},
		},
		{
			Key:  "TestTaskOrders2",
			Name: "FakeTaskType",
			Data: map[string]interface{}{"order": 2},
		},
		{
			Key:  "TestTaskOrders3",
			Name: "FakeTaskType",
			Data: map[string]interface{}{"order": 3},
		},
	}

	for _, d := range drivers {
		// Clear the queue:
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		// Create the tasks
		for _, task := range tasks {
			err = d.addTask(task.Name, task.Key, time.Now(), task.Data)

			time.Sleep(100 * time.Millisecond)

			if err != nil {
				t.Error(err)
				continue
			}
		}

		// Should now be able to fetch each task in the order they were added

		for _, task := range tasks {
			fetched, err := d.pop()

			if err != nil {
				t.Error(err)
				continue
			}

			if fetched.Key != task.Key {
				t.Errorf("Expected task with key %s, but had key %s", task.Key, fetched.Key)
			}

			err = d.complete(fetched.id, "Completed")

			if err != nil {
				t.Error(err)
				continue
			}
		}
	}
}

func TestTaskRetry(t *testing.T) {
	for _, d := range drivers {
		taskKey := "testTaskRetry1"
		// Clear the queue
		err := d.clear()

		if err != nil {
			t.Error(err)
			continue
		}

		err = d.addTask("testTaskRetry1", taskKey, time.Now(), map[string]interface{}{})

		if err != nil {
			t.Error(err)
			continue
		}

		// Fetch the task (which is only task in queue):
		task, err := d.pop()

		if err != nil {
			t.Error(err)
			continue
		}

		// Now we set this task as marked for retry:
		err = d.retry(task.id, "Retry")

		if err != nil {
			t.Error(err)
			continue
		}

		// Now if we pop, should get nothing:

		_, err = d.pop()

		if err != ErrNoTasks {
			t.Errorf("No result should have returned.  Err statement: %s", err)
			continue
		}

		// Refresh with time 1 hour, should still get no task:
		err = d.refreshRetry(time.Hour)

		if err != nil {
			t.Error(err)
			continue
		}

		_, err = d.pop()

		if err != ErrNoTasks {
			t.Errorf("No result should have returned.  Err statement: %s", err)
			continue
		}

		// Now if we refresh and pop with time 0, should get task back:
		err = d.refreshRetry(0)

		if err != nil {
			t.Error(err)
			continue
		}

		newTask, err := d.pop()

		if err != nil {
			t.Errorf("Should have refetched task, but didn't: %s", err)
			continue
		}

		if newTask.id != task.id {
			t.Error("waa1")
		}

		if newTask.State != TaskReady {
			t.Errorf("should have been ready, but was %s", newTask.State)
		}

	}
}
