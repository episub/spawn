// Code generated by gnorm, DO NOT EDIT!

package todo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/example/todo/gnorm"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// TableName is the primary table that this particular gnormed file deals with.
const TableName = "todo"

// Row represents a row from 'todo'.
type Row struct {
	TodoID  int    // todo_id (PK)
	Content string // content
	Done    bool   // done
	UserID  int    // user_id
}

// Field values for every column in Todo.
var (
	ContentCol = "content"
	DoneCol    = "done"
	TodoIDCol  = "todo_id"
	UserIDCol  = "user_id"
)

// All retrieves all rows from 'todo' as a slice of Row.
func All(ctx context.Context, db gnorm.DB) ([]Row, error) {
	qry := gnorm.Qry().Select(`content, done, todo_id, user_id`)
	qry.From(`public.todo`)
	sqlstr, _, err := qry.ToSql()
	if err != nil {
		return nil, err
	}

	var vals []Row
	q, err := db.Query(sqlstr)
	if err != nil {
		return nil, errors.Wrap(err, "query Todo")
	}
	for q.Next() {
		r := Row{}
		err := q.Scan(
			&r.Content,
			&r.Done,
			&r.TodoID,
			&r.UserID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "all Todo")
		}
		vals = append(vals, r)
	}
	return vals, nil
}

// CountQuery retrieve one row from 'todo'.
func CountQuery(ctx context.Context, db gnorm.DB, where []sq.Sqlizer) (int, error) {
	qry := gnorm.Qry().Select(`count(*) as count`)
	qry = qry.From("public.todo")
	for _, w := range where {
		qry = qry.Where(w)
	}

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return 0, err
	}

	count := 0
	err = db.QueryRow(sqlstr, args...).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "count Todo")
	}
	return count, nil
}

// Query retrieves rows from 'todo' as a slice of Row.
func Query(ctx context.Context, db gnorm.DB, where []sq.Sqlizer) ([]Row, error) {
	qry := gnorm.Qry().Select(`todo_id, content, done, user_id`)
	qry = qry.From("public.todo")
	for _, w := range where {
		qry = qry.Where(w)
	}

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return nil, err
	}

	var vals []Row
	q, err := db.Query(sqlstr, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query Todo")
	}
	for q.Next() {
		r := Row{}
		err := q.Scan(
			&r.TodoID,
			&r.Content,
			&r.Done,
			&r.UserID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "query Todo")
		}
		vals = append(vals, r)
	}
	return vals, nil
}

// PaginatedQuery Query used to get paginated results.  Can be replaced with
// a custom query of your own choosing that will allow you to sort or filter
// based on related fields as well
var PaginatedQuery = gnorm.
	Qry().
	Select("p.content, p.done, p.todo_id, p.user_id").
	From("public.todo as p")

// QueryPaginated retrieves rows from 'Todo' as a slice of Row.  If count == 0, then returns all results.  Returns true if there are more results to be had than those listed
// It will first grab a list of the relevant ID's, then fetch the full objects separately.  Done this way so that we can use custom queries that join more rows for use in sorting and filtering.
func QueryPaginated(ctx context.Context, db gnorm.DB, cursor *string, where []sq.Sqlizer, order gnorm.Order, count int64) (vals []Row, hasMore bool, total int, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "QueryPaginated Todo")
	defer span.Finish()

	qry := PaginatedQuery

	for _, w := range where {
		qry = qry.Where(w)
	}

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return
	}

	// Get the total number of rows possible with our existing query, before we add the cursor related conditional
	countSQL := fmt.Sprintf("SELECT count(*) FROM (%s) AS xyz", sqlstr)
	span.LogFields(
		log.String("countQuery", countSQL),
	)
	err = db.QueryRow(countSQL, args...).Scan(&total)

	if err != nil {
		return
	}

	if cursor != nil {
		w, args := gnorm.PaginateCursorWhere(*cursor, order, sqlstr, "todo_id")
		qry = qry.Where(w, args...)
	}

	err = order.AddField("todo_id")
	if err != nil {
		return
	}

	// Order of results:
	qry = qry.OrderBy(order.String())

	if count > 0 {
		qry = qry.Limit(uint64(count) + 1)
	}

	sqlstr, args, err = qry.ToSql()
	if err != nil {
		return
	}

	pageQuery := fmt.Sprintf("SELECT todo_id FROM (%s) AS xyz", sqlstr)
	span.LogFields(
		log.String("query", pageQuery),
	)
	q, err := db.Query(pageQuery, args...)
	if err != nil {
		return
	}

	// Collect the ID's together, and use them to fetch the full objects.  We do this in two stages because sometimes we will explicitly replace the auto-generated query used for this function, fullTodoIDPaginatedQuery.  This allows us to use custom queries that are tailored to allow sorting and filtering by fields that may not be available in just the parent and child (version) tables alone.
	var fetchIDs []int
	for q.Next() {
		var r int

		err = q.Scan(&r)
		if err != nil {
			return
		}

		fetchIDs = append(fetchIDs, r)
	}

	if len(fetchIDs) == 0 {
		return
	}

	fetched, err := Query(ctx, db, []sq.Sqlizer{gnorm.InInt(TodoIDCol, fetchIDs)})

	// Now create vals:
	vals = make([]Row, len(fetchIDs))

	found := 0
	for i, id := range fetchIDs {
		for _, v := range fetched {
			if id == v.TodoID {
				vals[i] = v
				found++
				continue
			}
		}
	}

	if found != len(fetchIDs) {
		err = fmt.Errorf("Could not find all required records")
	}

	// If count was more than 0 and we received more results than count, there are more rows to fetch
	if count > 0 && int64(len(fetchIDs)) > count {
		hasMore = true
		vals = vals[:len(vals)-1]
	}

	return
}

// Find retrieves a row from 'todo' by its primary key(s).
func Find(ctx context.Context, db gnorm.DB,
	todoID int,
) (Row, error) {
	const sqlstr = `SELECT
		content, done, todo_id, user_id
	FROM public.todo WHERE ( todo_id = $1 )`

	r := Row{}
	err := db.QueryRow(sqlstr,
		todoID,
	).Scan(&r.Content,
		&r.Done,
		&r.TodoID,
		&r.UserID,
	)
	if err != nil {
		return Row{}, errors.Wrap(err, "find Todo")
	}
	return r, nil
}

// One retrieve one row from 'todo'.
func One(ctx context.Context, db gnorm.DB, where []sq.Sqlizer, order *gnorm.Order) (Row, error) {
	qry := gnorm.Qry().Select(`todo_id, content, done, user_id`)
	qry = qry.From("public.todo")

	for _, w := range where {
		qry = qry.Where(w)
	}

	if order != nil {
		qry = qry.OrderBy(order.String())
	}

	qry = qry.Limit(1)

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return Row{}, err
	}

	r := Row{}
	err = db.QueryRow(sqlstr, args...).Scan(&r.TodoID,
		&r.Content,
		&r.Done,
		&r.UserID,
	)
	if err != nil {
		return Row{}, errors.Wrap(err, "queryOne Todo")
	}
	return r, nil
}

// Upsert Creates or updates record based on input
func Upsert(ctx context.Context, db gnorm.DB, o Row) (Row, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpsertTodo")
	defer span.Finish()

	var err error
	if err = prepareCreate(ctx, &o); err != nil {
		return o, err
	}

	// sql query
	const sqlstr = `INSERT INTO public.todo (` +
		`content, done, todo_id, user_id` +
		`) VALUES (` +
		`$1, $2, $3, $4` +
		`) ON CONFLICT (todo_id) DO UPDATE SET (` +
		`content, done, todo_id, user_id` +
		`) = (` +
		`EXCLUDED.content, EXCLUDED.done, EXCLUDED.todo_id, EXCLUDED.user_id` +
		`)`

	// run query
	_, err = db.Exec(sqlstr, o.Content, o.Done, o.TodoID, o.UserID)
	if err != nil {
		return o, err
	}

	return o, nil
}

// Delete deletes the Row from the database. Returns the number of items deleted.
func Delete(ctx context.Context,
	db gnorm.DB,
	todoID int,
) (int64, error) {
	const sqlstr = `DELETE FROM public.todo 
	WHERE
	  todo_id = $1
	`

	res, err := db.Exec(sqlstr, todoID)
	if err != nil {
		return 0, errors.Wrap(err, "delete Todo")
	}
	rows := res.RowsAffected()
	return rows, nil
}

// DeleteWhere deletes Rows from the database and returns the number of rows deleted.
func DeleteWhere(ctx context.Context, db gnorm.DB, where []sq.Sqlizer) (int64, error) {
	qry := gnorm.Qry().Delete("")
	qry = qry.From("public.todo")
	for _, w := range where {
		qry = qry.Where(w)
	}

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return 0, err
	}

	res, err := db.Exec(sqlstr, args...)
	if err != nil {
		return 0, errors.Wrap(err, "delete Todo")
	}
	return res.RowsAffected(), nil
}

// DeleteAll deletes all Rows from the database and returns the number of rows deleted.
func DeleteAll(ctx context.Context, db gnorm.DB) (int64, error) {
	const sqlstr = `DELETE FROM public.todo`

	res, err := db.Exec(sqlstr)
	if err != nil {
		return 0, errors.Wrap(err, "deleteall Todo")
	}
	return res.RowsAffected(), nil
}

// Update Updates with the provided records and condition
func Update(ctx context.Context, db gnorm.DB, updates map[string]interface{}, where []sq.Sqlizer) (int64, error) {
	qry := gnorm.Qry().Update("").Table("public.todo")

	qry = qry.SetMap(updates)
	for _, w := range where {
		qry = qry.Where(w)
	}

	sqlstr, args, err := qry.ToSql()
	if err != nil {
		return 0, err
	}

	res, err := db.Exec(sqlstr, args...)
	if err != nil {
		return 0, errors.Wrap(err, "update Todo")
	}
	return res.RowsAffected(), nil
}

// prepareCreate Prepares some fields for a new row if they haven't been provided already.  For example, primary key UUID values, created, etc
func prepareCreate(ctx context.Context, o *Row) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prepareCreate Todo")
	defer span.Finish()

	if o == nil {
		return nil
	}

	return nil
}