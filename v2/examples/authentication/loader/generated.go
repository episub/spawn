package loader

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

func runBatchLoaders() {
	log.Printf("Initialising batch loaders...")

	go Loader.runTodoBatcher()
	go Loader.runSessionBatcher()
	go Loader.runUserBatcher()

	log.Printf("...done")
}

// updatePath Used to keep track of nested field name in create or update actions.  E.g., address in a client update should be something like, client.person.address.address1.  This allows us to send back informative errors to the client so they can track which field exactly an error relates to
const updatePath = "updatePath"

// getPath Returns the path for the given field, so that we can nest validation error fields.  E.g., return client.person.email instead of just email
func getPath(ctx context.Context, field string) string {
	var path string
	p, ok := ctx.Value(updatePath).([]string)

	if ok {
		p = append(p, field)
		path = strings.Join(p, ".")
	}

	return path
}

// isTopPath Returns true if this is the top path
func isTopPath(ctx context.Context) bool {
	p, _ := ctx.Value(updatePath).([]string)

	return !(len(p) > 0)
}

func addFieldGQLError(ctx context.Context, message string, field string) {
	path := getPath(ctx, field)

	graphql.AddError(ctx, &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"field": path,
		},
	})
}

func addPathToContext(ctx context.Context, path string) context.Context {
	p, ok := ctx.Value(updatePath).([]string)

	if !ok {
		p = []string{}
	}

	return context.WithValue(ctx, updatePath, append(p, path))
}

// Returns true if there's any graphql errors.  If 'top' is set to true, it only returns such errors if this is the top path
func hasGQLErrors(ctx context.Context, top bool) bool {
	// We don't care about 'ok' value because if it's not set, we can assume path is top level
	p, _ := ctx.Value(updatePath).([]string)

	if top && len(p) > 0 {
		return false
	}

	rctx := graphql.GetRequestContext(ctx)

	if rctx != nil && len(rctx.Errors) > 0 {
		return true
	}

	return false
}
