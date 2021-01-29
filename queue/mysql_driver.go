package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

// MySQLDriver MySQL Driver
type MySQLDriver struct {
	db        *sqlx.DB
	tableName string
}

// NewMySQLDriver Returns a new mysql driver, initialised.  readTimeout is in seconds
func NewMySQLDriver(username string, password string, host string, db string, table string, readTimeout int64) (*MySQLDriver, error) {
	var err error

	m := &MySQLDriver{}

	connString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&readTimeout=%ds", username, password, host, db, readTimeout)

	m.db, err = sqlx.Open("mysql", connString)

	m.tableName = table

	return m, err
}

func (m *MySQLDriver) taskQueryColumns() string {
	return "a." + m.tableName + "_id, a.task_key, a.task_name, a.created, a.data, a.state"
}

// clear Removes all entries from the queue.  Be careful.  Generally you should cancel entries rather than delete.
func (m *MySQLDriver) clear() error {
	_, err := m.db.Exec(fmt.Sprintf("DELETE FROM %s", m.tableName))

	return err
}

func (m *MySQLDriver) name() string {
	return "MySQLDriver"
}

// AddTask Adds a task to the queue
func (m *MySQLDriver) addTask(taskName string, taskKey string, doAfter time.Time, data map[string]interface{}) error {
	if time.Now().Before(doAfter) {
		log.Printf("WARNING: doAfter not implemented for MySQLDriver")
	}
	// Store data as json:
	dataString, err := json.Marshal(data)

	created := time.Now()
	// Convert
	_, err = m.db.Exec("INSERT INTO "+m.tableName+" (nsync_queue_id, data, state, task_key, task_name, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?)", uuid.NewV4(), dataString, "READY", taskKey, taskName, created, created)

	return err
}

func (m *MySQLDriver) getTask(taskName string) (Task, error) {
	var task Task
	var err error

	query := `SELECT ` + m.taskQueryColumns() + ` FROM ` + m.tableName + ` WHERE task_name = ? ORDER BY created DESC LIMIT 1`

	task, err = m.scanTask(m.db.QueryRow(query))

	return task, err
}

func (m *MySQLDriver) pop() (Task, error) {
	var task Task
	var data string

	query := `SELECT ` + m.taskQueryColumns() + `
FROM ` + m.tableName + ` a
LEFT OUTER JOIN ` + m.tableName + ` b
ON a.task_key = b.task_key 
AND (a.created < b.created OR (a.created = b.created AND a.` + m.tableName + `_id < b.` + m.tableName + `_id))
WHERE b.task_key IS NULL
AND a.state = 'READY'
ORDER BY a.created ASC LIMIT 1`

	err := m.db.QueryRow(query).Scan(&task.id, &task.Key, &task.Name, &task.Created, &data, &task.State)

	if err == sql.ErrNoRows {
		return task, ErrNoTasks
	}

	if err != nil {
		return task, err
	}

	err = json.Unmarshal([]byte(data), &task.Data)

	return task, err
}

func (m *MySQLDriver) refreshRetry(age time.Duration) error {
	when := time.Now().Add(-age)
	_, err := m.db.Exec("UPDATE "+m.tableName+" SET state=?, UPDATED=? WHERE state=? AND UPDATED < ?", string(TaskReady), time.Now(), string(TaskRetry), when)

	return err
}

func (m *MySQLDriver) getQueueLength() (int64, error) {
	var length int64

	err := m.db.Get(&length, "SELECT count(*) FROM "+m.tableName+" LIMIT 1")

	return length, err
}

func (m *MySQLDriver) complete(id string, message string) error {
	if len(message) > 0 {
		log.Printf("WARNING: message not implemented")
	}
	return m.setTaskState(id, TaskDone)
}

func (m *MySQLDriver) cancel(id string, message string) error {
	if len(message) > 0 {
		log.Printf("WARNING: message not implemented")
	}
	return m.setTaskState(id, TaskCancelled)
}

func (m *MySQLDriver) fail(id string, message string) error {
	if len(message) > 0 {
		log.Printf("WARNING: message not implemented")
	}
	return m.setTaskState(id, TaskFailed)
}

func (m *MySQLDriver) retry(id string, message string) error {
	if len(message) > 0 {
		log.Printf("WARNING: message not implemented")
	}
	return m.setTaskState(id, TaskRetry)
}

func (m *MySQLDriver) setTaskState(id string, state TaskState) error {
	_, err := m.db.Exec("UPDATE "+m.tableName+" SET state=?, UPDATED=? WHERE "+m.tableName+"_id = ?", string(state), time.Now(), id)

	return err
}

func (m *MySQLDriver) scanTask(scanner *sql.Row) (Task, error) {
	var task Task
	var data string

	err := scanner.Scan(&task.id, &task.Key, &task.Name, &task.Created, &data, &task.State)

	if err != nil {
		return task, err
	}

	err = json.Unmarshal([]byte(data), &task.Data)

	return task, err
}
