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

func (ea *ExampleScheduledAction) Stream() string {
	return "test"

}

func NewExampleTaskAction(result chan bool) ExampleTaskAction {
	ea := ExampleTaskAction{}

	ea.result = result

	return ea
}

type ExampleTaskAction struct {
	result chan bool
}

func (ea *ExampleTaskAction) Do(task Task) (TaskResult, string) {
	ea.result <- true
	return TaskResultSuccess, "Done"
}
