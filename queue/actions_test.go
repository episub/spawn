package queue

func NewExampleScheduledAction(result chan bool) ExampleScheduledAction {
	ea := ExampleScheduledAction{}

	ea.result = result

	return ea
}

type ExampleScheduledAction struct {
	result chan bool
}

func (ea *ExampleScheduledAction) Do() error {
	ea.result <- true
	return nil

}

func NewExampleTaskAction(result chan bool) ExampleTaskAction {
	ea := ExampleTaskAction{}

	ea.result = result

	return ea
}

type ExampleTaskAction struct {
	result chan bool
}

func (ea *ExampleTaskAction) Do(task Task) (TaskResult, TaskMessage) {
	ea.result <- true
	return TaskResultSuccess, TaskMessage("Done")
}
