package resolvers

import (
	"context"
	"fmt"

	"github.com/example/todo/gnorm/public/todo"
	"github.com/example/todo/gnorm/public/user"
	"github.com/example/todo/graph"
	"github.com/example/todo/loader"
	"github.com/example/todo/models"
)

type Resolver struct{}

func (r *Resolver) Mutation() graph.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() graph.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateTodo(ctx context.Context, input models.NewTodo) (*todo.Row, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Todos(ctx context.Context) ([]*todo.Row, error) {
	all, _, _, err := loader.Loader.GetAllTodo(ctx, models.Filter{})
	// Return as pointers:
	res := make([]*todo.Row, len(all))
	for i, t := range all {
		res[i] = &t
	}
	return res, err
}

func (r *Resolver) Todo() graph.TodoResolver {
	return &todoResolver{r}
}

type todoResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

func (t *todoResolver) User(ctx context.Context, obj *todo.Row) (*user.Row, error) {
	user, err := loader.Loader.GetUser(ctx, obj.UserID)
	return &user, err
}

func (t *todoResolver) ID(ctx context.Context, obj *todo.Row) (string, error) {
	return fmt.Sprintf("%d", obj.TodoID), nil
}

func (r *Resolver) User() graph.UserResolver {
	return &userResolver{r}
}

func (u *userResolver) ID(ctx context.Context, obj *user.Row) (string, error) {
	return fmt.Sprintf("%d", obj.UserID), nil
}
