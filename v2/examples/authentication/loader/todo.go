package loader

import (
	"context"

	"github.com/example/todo/gnorm/public/todo"
)

func hydrateModelTodo(ctx context.Context, i todo.Row) todo.Row {
	return i
}
