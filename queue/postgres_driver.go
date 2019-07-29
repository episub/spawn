package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx"
)

// PostgresDriver PostgreSQL Driver
type PostgresDriver struct {
	tableName string
	pool      *pgx.ConnPool
	config    pgx.ConnConfig
}

// NewPostgresDriver Returns a new postgres driver, initialised.  readTimeout is in seconds
func NewPostgresDriver(dbUser string, dbPass string, dbHost string, dbName string, dbTable string) (*PostgresDriver, error) {
	var err error

	d := &PostgresDriver{
		tableName: dbTable,
	}

	connString := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", dbUser, dbName, dbPass, dbHost)
	connConfig, err := pgx.ParseConnectionString(connString)

	if err != nil {
		return nil, err
	}

	d.config = connConfig

	d.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 5 /* https://wiki.postgresql.org/wiki/Number_Of_Database_Connections#How_to_Find_the_Optimal_Database_Connection_Pool_Size */})

	if err != nil {
		return nil, err
	}

	return d, err
}

func (d *PostgresDriver) taskQueryColumns() string {
	return "a." + d.tableName + "_id, a.task_key, a.task_name, a.created_at, a.data, a.state"
}

// clear Removes all entries from the queue.  Be careful.  Generally you should cancel entries rather than delete.
func (d *PostgresDriver) clear() error {
	_, err := d.pool.Exec(fmt.Sprintf("DELETE FROM %s", d.tableName))

	return err
}

func (d *PostgresDriver) name() string {
	return "PostgresDriver"
}

// AddTask Adds a task to the queue
func (d *PostgresDriver) addTask(taskName string, taskKey string, data map[string]interface{}) error {
	// Store data as json:
	dataString, err := json.Marshal(data)

	created := time.Now()
	// Convert
	_, err = d.pool.Exec(`
INSERT INTO `+d.tableName+`
	(`+d.tableName+`_id, data, state, task_key, task_name, created_at, last_attempted, last_attempt_message)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, 'Created')`,
		dataString,
		"READY",
		taskKey,
		taskName,
		created,
		created,
	)

	return err
}

func (d *PostgresDriver) getTask(taskName string) (Task, error) {
	var task Task
	var err error

	query := `SELECT ` + d.taskQueryColumns() + ` FROM ` + d.tableName + ` WHERE task_name = $1 ORDER BY created_at DESC LIMIT 1`

	task, err = d.scanTask(d.pool.QueryRow(query))

	return task, err
}

func (d *PostgresDriver) pop() (Task, error) {
	var task Task
	var data string

	query := `
UPDATE message_queue a SET last_attempted=Now(), last_attempt_message='Attempting', state='` + string(TaskInProgress) + `'
WHERE message_queue_id IN (
	SELECT message_queue_id
	FROM message_queue
	WHERE state IN ('` + string(TaskReady) + `')
	OR (
		last_attempted < Now() - INTERVAL '10 minute'
		AND state IN ('` + string(TaskInProgress) + `', '` + string(TaskRetry) + `')
	)
	ORDER BY last_attempted ASC
)
RETURNING ` + d.taskQueryColumns()

	err := d.pool.QueryRow(query).Scan(&task.id, &task.Key, &task.Name, &task.Created, &data, &task.State)

	if err == sql.ErrNoRows {
		return task, ErrNoTasks
	}

	if err != nil {
		return task, err
	}

	err = json.Unmarshal([]byte(data), &task.Data)

	return task, err
}

func (d *PostgresDriver) refreshRetry(age time.Duration) error {
	when := time.Now().Add(-age)
	_, err := d.pool.Exec("UPDATE "+d.tableName+" SET state=$1, last_attempted=$2 WHERE state=$3 AND last_attempted < $4", string(TaskReady), time.Now(), string(TaskRetry), when)

	return err
}

func (d *PostgresDriver) getQueueLength() (int64, error) {
	var length int64

	err := d.pool.QueryRow("SELECT count(*) FROM " + d.tableName + " LIMIT 1").Scan(&length)

	return length, err
}

func (d *PostgresDriver) complete(id string, message string) error {
	return d.setTaskState(id, TaskDone, message)
}

func (d *PostgresDriver) cancel(id string, message string) error {
	return d.setTaskState(id, TaskCancelled, message)
}

func (d *PostgresDriver) fail(id string, message string) error {
	return d.setTaskState(id, TaskFailed, message)
}

func (d *PostgresDriver) retry(id string, message string) error {
	return d.setTaskState(id, TaskRetry, message)
}

func (d *PostgresDriver) setTaskState(id string, state TaskState, message string) error {
	_, err := d.pool.Exec("UPDATE "+d.tableName+" SET state=$1, last_attempted=$2, last_attempt_message=$3 WHERE "+d.tableName+"_id = $4", string(state), time.Now(), message, id)

	return err
}

func (d *PostgresDriver) scanTask(scanner *pgx.Row) (Task, error) {
	var task Task
	var data string

	err := scanner.Scan(&task.id, &task.Key, &task.Name, &task.Created, &data, &task.State)

	if err != nil {
		return task, err
	}

	err = json.Unmarshal([]byte(data), &task.Data)

	return task, err
}
