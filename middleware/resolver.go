package middleware

import (
	"context"
	"errors"
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
	reflections "gopkg.in/oleiade/reflections.v1"
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

// CheckAllowed Checks whether user is allowed to perform the requested
// mutation.  Returns the reason (which may be empty) and error (which may
// also contain the reason text)
func CheckAllowed(
	ctx context.Context,
	prefix string,
	objectName string,
	input map[string]interface{},
) (string, interface{}, error) {
	// Attempt the authz policy first.  If we have an error, then we try allow
	allowed, reason, data, err := opa.Authorised(ctx, getAuthString(prefix, objectName, "authz"), input)

	if err != nil {
		// Trouble with authz, so fall back to allow check:
		log.WithField("error", err).Debug("Failed authz, so will attempt allow policy")
		allowed, err = opa.Allow(ctx, getAuthString(prefix, objectName, "allow"), input)
	}

	if err != nil {
		return reason, data, err
	}

	if !allowed {
		var msg string
		if len(reason) > 0 {
			msg = ": " + reason
		}
		return reason, data, permissionDeniedError(fmt.Sprintf("%s.%s%s", prefix, objectName, msg))
	}

	return reason, data, nil
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
			reason, data, err := runAllowCheck(ctx, requestPayload, rctx, rctx.Args)
			if err != nil {
				var msg string
				if len(reason) > 0 {
					msg = reason
				} else {
					msg = err.Error()
				}
				graphql.AddError(ctx, &gqlerror.Error{
					Message: msg,
					Extensions: map[string]interface{}{
						"code":  "ALLOW_CHECK_FAIL",
						"error": err.Error(),
						"data":  data,
					},
				})
				return nil, nil
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
			reason, data, err := runAllowCheck(ctx, requestPayload, rctx, res)
			// err = requestPayload(ctx, strings.ToLower(rctx.Object), rctx.Field.Name, res)
			if err != nil {
				var msg string
				if len(reason) > 0 {
					msg = reason
				} else {
					msg = err.Error()
				}
				graphql.AddError(ctx, &gqlerror.Error{
					Message: msg,
					Extensions: map[string]interface{}{
						"code":  "ALLOW_CHECK_FAIL",
						"error": err.Error(),
						"data":  data,
					},
				})
				return nil, nil
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

// runAllowCheck Returns reason, if given one, and error.  Error may contain
// the reason as well.
func runAllowCheck(
	ctx context.Context,
	requestPayload func(context.Context, string, string, map[string]interface{}) error,
	rctx *graphql.ResolverContext,
	data interface{},
) (string, interface{}, error) {
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
		return "", nil, err
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

func getField(object interface{}, fieldName string) (field string, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("Failed to get field on object")
		}
	}()

	f, err := reflections.GetField(object, fieldName)
	field = f.(string)
	return field, err
}

func arrayContains(list []string, v string) bool {
	for _, f := range list {
		if v == f {
			return true
		}
	}

	return false
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
	// panic("Is it safe to cache these unnamed values like I am?")

	if rctx.Object != "Mutation" && !strings.HasPrefix(rctx.Object, "__") {
		var parent interface{}
		if rctx.Parent != nil {
			parent = rctx.Parent.Result
		}

		field := rctx.Field.Alias
		policy := fmt.Sprintf("data.api.entity.%s", lowerFirst(rctx.Object))
		cacheName := policy

		// We want to remember if this policy check was for the same part of the
		// tree or somewhere very different.  We shorten fl by one so that, e.g.,
		// suppliersConnection.edges.node.name becomes
		// suppliersConnection.edges.node.  That way we can cache the field list
		// for all of the node fields, rather than none being cached
		fl := fullFieldList(*rctx, false)
		if len(fl) > 0 {
			fl = fl[:len(fl)-1]
		}

		id, err := getField(parent, "ID")

		// TODO IMPORTANT: Eliminate the dependency on viewField.  All fields
		// should be explicitly listed.  This will allow us to eliminate some
		// extra code here that falls back to viewField policy
		// For now, only objects with an ID field are being cached
		if err == nil && len(id) > 0 {
			cacheName = fmt.Sprintf("%s:%s:%s", strings.Join(fl, "."), id, cacheName)
			// cacheName = fmt.Sprintf("%s:%s", id, cacheName)
			// log.Printf("Cache search key: %s", cacheName)

			v, err := store.ContextReadValue(ctx, vars.SharedData, cacheName)
			if err != nil {
				return false, err
			}

			vb, ok := v.([]string)

			// If len == 0, then this might be because the policy allows all fields
			// and we need to fall bacak on the viewField policy
			if ok && len(vb) > 0 {
				// log.Printf("RETURNING CACHED VALUE object: %s, field: %s", rctx.Object, rctx.Field.Alias)
				return arrayContains(vb, field), nil
			}
		}
		// log.Printf("CAN'T USE CACHED VALUE FOR object: %s, field: %s", rctx.Object, rctx.Field.Alias)

		//log.Printf("Object: %s, field: %s", rctx.Object, rctx.Field.Alias)
		// Not found in cache, so let's check the policy:
		allowed = false

		input := make(map[string]interface{})
		input["field"] = field

		// Add some default data to the check
		input["fieldValue"] = object
		input["entity"] = parent

		err = defaultPayload(ctx, rctx.Object, rctx.Field.Alias, input)
		if err != nil {
			log.Printf("WARNING: Could not add default payload: %s", err)
		}

		allFields, err := opa.AuthorisedStrings(ctx, fmt.Sprintf("%s.allowedFields", policy), input)
		if err == nil && len(allFields) > 0 {
			// Save in cache:
			if len(cacheName) > 0 {
				cErr := store.ContextAddValue(ctx, vars.SharedData, cacheName, allFields)
				if cErr != nil {
					log.Error(cErr)
				}
			}
			return arrayContains(allFields, field), nil
		}

		// Failing finding the specific field, there might be a general policy
		// that allows any field to be viewed:
		log.Warningf("viewField will soon be deprecated.  If you are not already, please specify allowedFields permission to explicitly list all fields allowed.  Falling back to viewField for object: %s, field: %s", rctx.Object, field)
		allowed, err = opa.Allow(ctx, policy+".viewField", input)
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
