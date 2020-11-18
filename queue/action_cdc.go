package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx"
)

// CDCAction Action that should be taken based on row changes
type CDCAction string

// This implements the structure for a CDC (Change Data Capture) task, which
// will monitor for changes to specified records, and then call functions
// based on whether there is a new record (CREATE), a record has been removed
// (DELETE), or if a record has changed (UPDATE) based on selected fields.
//
// It works by maintaining a list of records that have been created, updated,
// or deleted, along with a hash at the time.  The hash is created within
// the database based on fields that we care to track.  For example, a record
// may have a field we don't care about, so we exclude that from the hash
// calculation when determining if a record has changed.

const (
	// CDCActionCreate A new record should be created
	CDCActionCreate CDCAction = "CREATE"
	// CDCActionDelete An existing record should be removed
	CDCActionDelete CDCAction = "DELETE"
	// CDCActionUpdate An existing record should be updated
	CDCActionUpdate CDCAction = "UPDATE"
)

// CDCRunnerAction Runs at regular intervals.  Takes in an arbitrary array
// of arrays.  The outer array represents rows of data with the changes to be
// tracked, while the inner array represents an ordered list of fields.  Order
// is important, because this is used to calculate the MD5 hash.
// array of arrays
// controllerID: A unique ID
type CDCRunnerAction struct {
	controllerID uuid.UUID
	sourceQuery  string
	schema       string
	db           *pgx.ConnPool
	createFunc   func(context.Context, CDCObjectAction) error
	updateFunc   func(context.Context, CDCObjectAction) error
	deleteFunc   func(context.Context, CDCObjectAction) error
}

// CDCObjectAction Object ID along with the action that should be taken
type CDCObjectAction struct {
	ObjectID     string
	Hash         string
	Action       CDCAction
	ControllerID string
	cdcAction    CDCRunnerAction
	stream       string
}

// NewCDCRunnerAction Creates and initialises a new Execute SQL Processor.
//
// sourceQuery: Must be a query that returns the following values:
// * object_id: a unique field identifying the object in question (likely a
//   primary key
// * hash: an md5 hash (fast) of the object, which should change when and only
//   when you want a change to be noted
// * schema: database schema that has the cdc_hash table
// * db a pgx ConnPool pointer with a live connection to database (postgres
//   only supported)
func NewCDCRunnerAction(
	controllerID uuid.UUID,
	sourceQuery string,
	schema string,
	db *pgx.ConnPool,
	createFunc func(context.Context, CDCObjectAction) error,
	updateFunc func(context.Context, CDCObjectAction) error,
	deleteFunc func(context.Context, CDCObjectAction) error,
) (CDCRunnerAction, error) {
	var err error

	c := CDCRunnerAction{
		controllerID: controllerID,
		sourceQuery:  sourceQuery,
		schema:       schema,
		db:           db,
		createFunc:   createFunc,
		updateFunc:   updateFunc,
		deleteFunc:   deleteFunc,
	}

	if db == nil {
		err = fmt.Errorf("DB cannot be nil")
	}

	return c, err
}

// Do Run a loop of trying to sync objects needing to sync
func (c CDCRunnerAction) Do() error {
	ctx := context.Background()
	fmt.Println("Attempting to get changes")
	objects, err := c.getChanges(1)

	if err != nil {
		return err
	}

	// Log out for now:
	fmt.Println("Objects retrieved for sync:")
	for _, v := range objects {
		log.Printf(" * %s (%s): %s", v.ObjectID, v.Hash, v.Action)
		err = v.do(ctx)
		if err != nil {
			log.Printf("Error marking as done: %s", err)
			return err
		}
	}

	return nil
}

// Stream Can't run actions for the same controller simultaneously
func (c CDCRunnerAction) Stream() string {
	return c.controllerID.String()
}

// MarkDone Mark this action as completed
func (c CDCObjectAction) do(ctx context.Context) error {
	var err error

	switch c.Action {
	case CDCActionCreate:
		err = c.cdcAction.createFunc(ctx, c)
	case CDCActionUpdate:
		err = c.cdcAction.updateFunc(ctx, c)
	case CDCActionDelete:
		err = c.cdcAction.deleteFunc(ctx, c)
	default:
		log.Printf("Error: Unknown action type %s", c.Action)
	}

	if err != nil {
		return err
	}

	switch c.Action {
	case CDCActionCreate, CDCActionUpdate:
		_, err = c.cdcAction.db.Exec(`
INSERT INTO `+c.cdcAction.schema+`.cdc_hash (cdc_controller_id, object_id, hash)
VALUES($1, $2, $3)
ON CONFLICT ON CONSTRAINT cdc_hash_controller_object_uq DO
UPDATE SET hash = EXCLUDED.hash, updated_at = Now()`, c.ControllerID, c.ObjectID, c.Hash)
	case CDCActionDelete:
		_, err = c.cdcAction.db.Exec(`
DELETE FROM `+c.cdcAction.schema+`.cdc_hash
WHERE cdc_controller_id = $1
AND object_id = $2
AND hash = $3::uuid`, c.ControllerID, c.ObjectID, c.Hash)
	default:
		log.Printf("Error: Unknown action type %s", c.Action)
	}

	return err
}

// getChanges Returns up to 'n' random rows that need updating/creating/deleting
func (c CDCRunnerAction) getChanges(
	n int,
) ([]CDCObjectAction, error) {
	// Find all the changes
	qry := fmt.Sprintf(`
WITH current AS (
	%s
)
SELECT
	COALESCE(s.object_id, c.object_id),
	CASE
		WHEN (s.object_id IS NULL) THEN 'CREATE'
		WHEN (c.object_id IS NULL) THEN 'DELETE'
		WHEN (COALESCE((c.hash != s.hash))) THEN 'UPDATE'
	END AS action,
	CASE
		WHEN (s.object_id IS NULL) THEN c.hash::varchar
		WHEN (c.object_id IS NULL) THEN s.hash::varchar
		WHEN (COALESCE((c.hash != s.hash))) THEN c.hash::varchar
	END AS hash,
	$1 AS controller_id
FROM %s.cdc_hash s
FULL OUTER JOIN current c
	ON c.object_id = s.object_id
	AND s.cdc_controller_id = $1
WHERE (
	s.hash != c.hash
	OR s IS NULL
	OR c IS NULL
)
ORDER BY RANDOM()
LIMIT %d
`, c.sourceQuery, c.schema, n)

	rows, err := c.db.Query(qry, c.controllerID)

	if err != nil {
		return []CDCObjectAction{}, err
	}
	defer rows.Close()

	var actions []CDCObjectAction
	for rows.Next() {
		action := CDCObjectAction{}
		if err := rows.Scan(
			&action.ObjectID,
			&action.Action,
			&action.Hash,
			&action.ControllerID,
		); err != nil {
			return []CDCObjectAction{}, err
		}
		action.cdcAction = c
		actions = append(actions, action)
	}

	return actions, nil
}
