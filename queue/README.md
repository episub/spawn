esync is a library to help keep some services synchronised with other services.

Should be safe for multiple processors to run simultaneously.  A driver's pop() function should not return the same value at the next call.

# Usage

The following is a basic example of how to use this library.

```Go
package main

import (
	"log"
	"os"
	"time"

	"github.com/episub/spawn/queue"
)

// Use a scheduled action to run tasks at intervals
type myScheduledAction struct{}

func (m myScheduledAction) Do() error {
	log.Printf("myScheduledAction done")
	return nil
}

// Use a queue action to handle tasks in a queue
type myQueueAction struct{}

func (m myQueueAction) Do(task queue.Task) (queue.TaskResult, queue.TaskMessage) {
	log.Printf("myQueueAction done.  Task key: %s", task.Key)

	return queue.TaskResultSuccess, queue.TaskMessage("No Problem")
}

func queueErrorManager(err error) {
	log.Printf("Error from sync manager: %s", err)
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_DB")
	dbTable := os.Getenv("DB_TABLE")

	postgresDriver, err := queue.NewPostgresDriver(dbUser, dbPass, dbHost, dbName, dbTable)

	if err != nil {
		log.Fatal(err)
	}

	sm := queue.NewSyncManager(postgresDriver)

	// Optionally set an error handler so that you can catch errors from the running loop and put them through your own logging solution
	sm.SetErrorHandler(queueErrorManager)

	// Schedule a regular action to perform at specified intervals
	sm.Schedule(myScheduledAction{}, time.Second*3)

	// Register our queue handler for tasks with name "exampleTask"
	sm.RegisterTaskHandler(myQueueAction{}, "exampleTask")
	data := map[string]interface{}{
		"testData": true,
	}

	tm := queue.NewTaskManager(postgresDriver)

	// Add one task to the queue
	err = tm.AddTask("exampleTask", "myKey", data)
	if err != nil {
		panic(err)
	}

	// Off we go
	sm.Run()
}
```

It may be possible in future versions for the same action to be used simultaneously, so be careful with pointer functions that may end up sharing values across goroutines.  Avoid pointer functions where possible.

# Details
You need to create an actionManager object, providing it a database driver object that meets the 'Driver' interface.  Included drivers:

* PostgreSQL (PostgresDriver)

You can use this library for either creating a service to run the synchronising actions, or for creating entries in a queue to be acted on by the synchronisation service.  At the very least you need a SyncManager.

## Concepts

* Data: an action can store data in the queue
* Task Key: this should uniquely identify a particular action.  Think of it as the primary key, though it may not be the actual primary key, depending on driver implementation.  **If there is more than one READY entry for the same task key, only the most recent will be performed**.
* Task Name: this identifies the type of task.  Action managers may handle particular task types.  For example, you may have a task name such as "CUSTOMER_UPDATE", with multiple database entries of that sort.  Try to keep one action per task name.

## Actions

Actions are descriptions of an act to perform.  You provide them with a function that will run when the action is to be performed.  You will then need to schedule the action to occur, either at specific intervals, or as the action to be performed by a particular queue item.

Actions should be designed to be safe to be used by multiple processes.  Therefore, avoid pointers.

Actions should gracefully return if they take too long, as they will block the main loop.

### Register action

Actions need to be registered for each task name.  If there is no registered action for a task name, then the particular task is cancelled when its turn comes.

## Queue

One will wish to create actions in the queue to be performed in good time.  Not every action needs to form part of a queue, but it is helpful to be able to queue actions to be performed in time.  To use the queue, you need a driver that provides a connection to the queue.  The driver needs to fulfil the 'Driver' interface.

## SyncManager

# Running

In some cases, another service may not handle multiple connections well -- for example, NetSuite.  In these cases you should ensure that you are only running one instance of this service.

# Driver

When designing a driver, you need to be careful that you don't implement a 'pop' that will ignore newer tasks.  Suppose that a task to update a customer is added, actioned, but before the action is finished a new update customer task is added.  You then return the action and mark it as finished.  This task should be performed again, so you need to be careful that the "mark as finished" task does not override the newer update task.

## PostgreSQL

```
CREATE TABLE public.message_queue(
	message_queue_id uuid NOT NULL DEFAULT gen_random_uuid(),
	data jsonb NOT NULL DEFAULT '{}',
	task_key varchar(64) NOT NULL,
	task_name varchar(64) NOT NULL,
	created_at timestamptz DEFAULT Now(),
	last_attempted timestamptz NOT NULL DEFAULT Now(),
	state varchar(16) NOT NULL,
	last_attempt_message varchar NOT NULL,
	CONSTRAINT message_queue_id_pk PRIMARY KEY (message_queue_id)
);
```

# TODO:

* Implement a timeout so that if some action blocks, then we can perform other actions
* Per-task retry intervals so that some tasks can be tried every minute, while others wait every hour.  Solution: separate queue runners per task type.  They can then check for and pop tasks that are specific to their task name.  When a task is registered, an interval is passed in the register call, and a new queue is spun up for that task.
