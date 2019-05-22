package resolvers

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/example/todo/gnorm"
	"github.com/example/todo/gnorm/public/todo"
	"github.com/example/todo/models"
)

func editableUpdateTodoFields(ctx context.Context, id string) ([]string, error) {
	return []string{}, nil
}

func sortTodo(ctx context.Context, sortField models.TodoSort, order gnorm.Order) (gnorm.Order, error) {
	var err error
	switch sortField {
	case models.TodoSortContent:
		err = order.AddField("content")
	default:
		err = fmt.Errorf("Unsupported sort field")
	}

	if err != nil {
		err = fmt.Errorf("Cannot sort by field %s: %s", sortField, err)
		return order, err
	}

	return order, nil
}

func filterTodo(ctx context.Context, f models.TodoFilter) (where []sq.Sqlizer, err error) {
	if f.Done != nil {
		where = append(where, sq.Eq{todo.DoneCol: *f.Done})
	}
	return
}

func (r *queryResolver) TodosConnection(
	ctx context.Context,
	first *int,
	after *string,
	last *int,
	before *string,
	filters *models.TodoFilter,
	sortField *models.TodoSort,
	sortDirection *models.SortDirection,
) (
	*models.TodosConnection,
	error,
) {
	con, err := queryTodos(ctx, first, after, last, before, filters, sortField, sortDirection, []sq.Sqlizer{})
	return &con, err
}
