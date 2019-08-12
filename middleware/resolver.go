package middleware

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/99designs/gqlgen/graphql"
	"github.com/episub/spawn/opa"
	"github.com/episub/spawn/store"
	"github.com/episub/spawn/vars"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
)

// DefaultPayloadFunc Called to fetch default payload
type DefaultPayloadFunc func(context.Context, string, string, map[string]interface{}) error

// RequestPayloadFunc Returns data specific to particular payloads
type RequestPayloadFunc func(context.Context, string, string, map[string]interface{}) error

// CheckAccess Used to determine a first-pass access to a query.  Can check basic things like, is this user logged in?
func CheckAccess(ctx context.Context, prefix string, object string, input map[string]interface{}) (bool, error) {
	user := ctx.Value("user")
	if user != nil {
		input["user"] = user
	}
	return opa.Allow(ctx, getAuthString(prefix, object, "access"), input)
}

// CheckAllowed Checks whether user is allowed to perform the requested mutation
func CheckAllowed(ctx context.Context, prefix string, objectName string, input map[string]interface{}) error {
	allowed, err := opa.Allow(ctx, getAuthString(prefix, objectName, "allow"), input)

	if err != nil {
		return err
	}

	if !allowed {
		return permissionDeniedError(fmt.Sprintf("%s.%s", prefix, objectName))
	}

	return nil
}

// mergeMap Merges the values in b into a
func mergeMap(a map[string]interface{}, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

func permissionDeniedError(objectName string) error {
	return fmt.Errorf("Not authorised to access %s", objectName)
}

// ResolverMiddleware Customise resolver middleware and include opentracing
// Performs multiple authorisation checks
func ResolverMiddleware(
	defaultPayloadFunc DefaultPayloadFunc,
	requestPayloadFunc RequestPayloadFunc,
) graphql.FieldMiddleware {
	opentracingMiddleware := OpentracingResolverMiddleware()
	defaultPayload := defaultPayloadFunc
	requestPayload := requestPayloadFunc
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		rctx := graphql.GetResolverContext(ctx)

		// Check if access is allowed, for parent queries.  Usually rejecting unauthorised access, except for specific queries that are public:
		if rctx.Object == "Query" || rctx.Object == "Mutation" {
			input := map[string]interface{}{"arguments": rctx.Args}
			name := rctx.Field.Name
			allowed, err := CheckAccess(ctx, strings.ToLower(rctx.Object), name, input)

			if !allowed {
				log.WithFields(logrus.Fields{"error": err, "query": name}).Info("Failed to authorise access to root level query")
				return nil, fmt.Errorf("Authorisation rejected based on 'access' policy for %s", rctx.Object)
			}
		}

		// Mutations are checked before completing action, while queries are checked after
		// This is because mutations can change state, and we only want to do that
		// if permitted.  Queries wait until after so that we're not fetching
		// objects from database multiple times.
		// PRE-FUNCTION CHECK: If this is a root level mutation, run the core 'allow' check:
		var err error
		if rctx.Object == "Mutation" {
			err := runAllowCheck(ctx, requestPayload, rctx, rctx.Args)
			if err != nil {
				return nil, err
			}
		}

		// Run the resolvers
		res, err := opentracingMiddleware(ctx, next)

		if err != nil {
			return nil, err
		}

		// If we have nothing, forward it on.  Nothing left to do.
		if res == nil {
			return res, err
		}

		// POST-FUNCTION CHECK: If this is a root level mutation, run the core 'allow' check:
		if rctx.Object == "Query" {
			err = runAllowCheck(ctx, requestPayload, rctx, res)
			// err = requestPayload(ctx, strings.ToLower(rctx.Object), rctx.Field.Name, res)
			if err != nil {
				return nil, err
			}
		}

		// Store the value we've had returned here so it can be used by OPA
		fields := fullFieldList(*rctx, false)
		fieldsStr := strings.Join(fields, ".")
		addData(ctx, fieldsStr, res)

		// hasFieldAccess disabled in lieu of field directives for now
		// Authorisation check: If this is a field, that the current user has access to view that field
		// Ignore __ prefix, since this corresponds to queries about the schema
		allowed, err := hasFieldAccess(ctx, res, defaultPayload)

		if err != nil || !allowed {
			field := fmt.Sprintf("%s.%s", rctx.Object, rctx.Field.Alias)
			msg := fmt.Sprintf("Not authorised to view field '%s'", field)
			graphql.AddError(ctx, &gqlerror.Error{
				Message: msg,
				Extensions: map[string]interface{}{
					"code":  "5",
					"field": field,
				},
			})
			if err != nil {
				log.Printf("Error authorising view field for '%s.%s': %s", rctx.Object, rctx.Field.Alias, err)
			}
			return nil, nil
		}

		return res, err
	}
}

func runAllowCheck(
	ctx context.Context,
	requestPayload func(context.Context, string, string, map[string]interface{}) error,
	rctx *graphql.ResolverContext,
	data interface{},
) error {
	input := make(map[string]interface{})

	dataMap, ok := data.(map[string]interface{})
	if ok {
		mergeMap(input, dataMap)
	} else {
		input["entity"] = data
	}
	mergeMap(input, rctx.Args)

	err := requestPayload(ctx, strings.ToLower(rctx.Object), rctx.Field.Name, input)
	if err != nil {
		return err
	}

	return CheckAllowed(ctx, strings.ToLower(rctx.Object), rctx.Field.Name, input)
}

// OpentracingResolverMiddleware Taken from an older version of gqlgen
func OpentracingResolverMiddleware() graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		rctx := graphql.GetResolverContext(ctx)
		span, ctx := opentracing.StartSpanFromContext(ctx, rctx.Object+"_"+rctx.Field.Name,
			opentracing.Tag{Key: "resolver.object", Value: rctx.Object},
			opentracing.Tag{Key: "resolver.field", Value: rctx.Field.Name},
		)
		defer span.Finish()
		ext.SpanKind.Set(span, "server")
		ext.Component.Set(span, "gqlgen")

		res, err := next(ctx)

		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(
				otlog.String("event", "error"),
				otlog.String("message", err.Error()),
				otlog.String("error.kind", fmt.Sprintf("%T", err)),
			)
		}

		return res, err
	}
}

// hasFieldAccess Verify access to the requested query field
func hasFieldAccess(ctx context.Context, object interface{}, defaultPayload func(context.Context, string, string, map[string]interface{}) error) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "hasFieldAccess")
	defer span.Finish()

	var err error
	rctx := graphql.GetResolverContext(ctx)
	span.LogFields(
		otlog.String("object", rctx.Object),
		otlog.String("field", rctx.Field.Alias),
	)

	allowed := true

	if rctx.Object != "Mutation" && !strings.HasPrefix(rctx.Object, "__") {
		field := rctx.Field.Alias
		policy := fmt.Sprintf("data.api.entity.%s.viewField", lowerFirst(rctx.Object))
		fl := fullFieldList(*rctx, false)
		cacheName := strings.Join(fl, ".") + ":" + policy

		// Check for cached answer:
		v, err := store.ContextReadValue(ctx, vars.SharedData, cacheName)
		if err != nil {
			return false, err
		}

		vb, ok := v.(bool)

		if ok {
			return vb, nil
		}

		// Not found in cache, so let's check the policy:
		allowed = false

		input := make(map[string]interface{})
		var parent interface{}
		input["field"] = field
		if rctx.Parent != nil {
			parent = rctx.Parent.Result
		}

		// Add some default data to the check
		input["fieldValue"] = object
		input["entity"] = parent

		err = defaultPayload(ctx, rctx.Object, rctx.Field.Alias, input)
		if err != nil {
			log.Printf("WARNING: Could not add default payload: %s", err)
		}

		allowed, err = opa.Allow(ctx, policy, input)

		// Caching is disabled for now: Implementing caching is tricky, because a cached answer for one record (e.g., a client viewing their own client field) might be used for another (e.g., a client viewing another client's field).

		//	// Store the result in cache
		//	if err == nil {
		//		//log.WithField("cacheName", cacheName).Infof("Adding field resolver value to store")
		//		cErr := store.ContextAddValue(ctx, sharedData, cacheName, allowed)
		//		if cErr != nil {
		//			log.WithField("error", cErr).Errorf("Error storing field value in cache")
		//		}
		//	}
	}

	return allowed, err
}

// fullFieldList Returns a list of the fields/names in the query.  E.g., getting the username of a user for a client would return 'client user username' as the three elements in order.  Does not count array position counts
func fullFieldList(rctx graphql.ResolverContext, useIndex bool) []string {
	var fields []string

	path := rctx.Path()

	for i, p := range path {
		t := reflect.TypeOf(p).String()
		switch t {
		case "string":
			s, _ := p.(string)
			fields = append(fields, s)
		case "int":
			if useIndex {
				v, _ := p.(int)
				fields = append(fields, fmt.Sprintf("%d", v))
			}
		default:
			log.Printf("Unknown type %s", t)
			fields[i] = "unknown"
		}
	}

	return fields
}

// addData Stores data in the current context
func addData(ctx context.Context, name string, value interface{}) {
	data := ctx.Value(vars.SharedData)

	if data == nil {
		return
	}

	dataMap := data.(store.DataStore)

	dataMap.AddValue(name, value)
}
func getAuthString(prefix string, objectName string, value string) string {
	return fmt.Sprintf("data.api.%s.%s.%s", prefix, objectName, value)
}

// https://groups.google.com/forum/#!topic/golang-nuts/WfpmVDQFecU
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
