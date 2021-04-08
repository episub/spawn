package loader

import (
	"context"

	"github.com/example/todo/gnorm/public/user"
)

func hydrateModelUser(ctx context.Context, i user.Row) (o user.Row) {
	return i
}
