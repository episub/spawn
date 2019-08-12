package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"github.com/episub/spawn/store"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/rego"
	opentracing "github.com/opentracing/opentracing-go"
)

// Permission Permission name and value
type Permission struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

// ErrNoPolicy Returns when no such policy
var ErrNoPolicy = fmt.Errorf("No such policy found")

// AuthorisedStrings Returns a string list of strings that are authorised by the policy.  Expects to get from policy an array of strings
func AuthorisedStrings(ctx context.Context, policy string, data map[string]interface{}) ([]string, error) {
	//func AuthorisedStrings(ctx context.Context, policy string, store *store.DataStore, data map[string]interface{}) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AuthorisedStrings")
	defer span.Finish()

	// allowed ID's
	var allowed []string

	// Call the policy, and get our response
	rs, err := runRego(ctx, policy, data)

	if err != nil {
		return allowed, err
	}

	// Explicitly convert to array of interfaces, and all of those interfaces should be strings though we cannot cast directly to []string
	allowedInterface, ok := rs[0].Expressions[0].Value.([]interface{})

	if !ok {
		return allowed, fmt.Errorf("Could not authorise action")
	}

	// Iterate through the results to convert to strings
	for _, a := range allowedInterface {
		allowed = append(allowed, a.(string))
	}

	return allowed, nil
}

// AuthorisedPermissions Returns an array []Permission for the provided list of permissions, and their relevant values
// Expects rootPolicy+permission+'allow' to be the name of the policy.  E.g.:
// rootPolicy:  data.api.repositories
// permission:  edit
// full policy: data.api.repositories.edit.allow
func AuthorisedPermissions(ctx context.Context, permissions []string, rootPolicy string, store *store.DataStore, data map[string]interface{}) ([]Permission, error) {
	var parsedPermissions []Permission

	// Iterate over each specified permission, checking if the user has it or not
	for _, p := range permissions {
		policy := fmt.Sprintf("%s.%s.allow", rootPolicy, p)
		allowed, err := Allow(ctx, policy, data)

		if err != nil {
			graphql.AddErrorf(ctx, fmt.Sprintf("Error verifying permission to %s for %s: %s", p, rootPolicy, err))
			allowed = false
		}

		parsedPermissions = append(parsedPermissions, Permission{Name: p, Value: allowed})
	}

	return parsedPermissions, nil
}

// Allow Returns a simple true/false answer for a true/false policy.  If policy
// does not exist, it returns false and no error, but logs it
func Allow(ctx context.Context, policy string, data map[string]interface{}) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Authorised")
	defer span.Finish()

	var allowed bool

	// Call the policy, and get our response
	rs, err := runRego(ctx, policy, data)

	if err != nil {
		return allowed, err
	}

	// Explicitly convert to array of interfaces, and all of those interfaces should be strings though we cannot cast directly to []string
	var ok bool

	// No such policy, but we just count that as false?
	if (len(rs) < 1) || (len(rs[0].Expressions) < 1) {
		log.Printf("No such policy %s", policy)
		return false, nil
	}
	allowed, ok = rs[0].Expressions[0].Value.(bool)

	if !ok {
		return false, fmt.Errorf("Could not authorise action.  Return type: %s", reflect.TypeOf(rs[0].Expressions[0].Value))
	}

	return allowed, nil
}

// Authorised Returns a true/false answer with a reason for a policy that
// returns an object with both 'reason' and 'value' attributes.  If policy
// does not exist, it returns ErrNoPolicy
func Authorised(
	ctx context.Context,
	policy string,
	data map[string]interface{},
) (bool, string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Authorised")
	defer span.Finish()

	var allowed bool
	var reason string

	// Call the policy, and get our response
	rs, err := runRego(ctx, policy, data)

	if err != nil {
		return allowed, reason, err
	}

	// Explicitly convert to array of interfaces, and all of those interfaces should be strings though we cannot cast directly to []string
	var ok bool

	// No such policy, but we just count that as false?
	if (len(rs) < 1) || (len(rs[0].Expressions) < 1) {
		return false, reason, ErrNoPolicy
	}
	reply, ok := rs[0].Expressions[0].Value.(map[string]interface{})

	if !ok {
		return false, reason, fmt.Errorf("Could not authorise action.  Return type: %s", reflect.TypeOf(rs[0].Expressions[0].Value))
	}

	allowed, ok = reply["value"].(bool)
	if !ok {
		return false, reason, fmt.Errorf("Expected value to be boolean, but was not")
	}

	// Don't care if this is missing, since reason should be optional
	reason, _ = reply["reason"].(string)

	return allowed, reason, nil
}

// GetInt Returns an integer given by the named policy
func GetInt(ctx context.Context, policy string, data map[string]interface{}) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetInt")
	defer span.Finish()

	// Call the policy, and get our response
	rs, err := runRego(ctx, policy, data)

	if err != nil {
		return 0, err
	}

	var tv json.Number
	var ok bool
	// No such policy, but we just count that as false?
	if (len(rs) < 1) || (len(rs[0].Expressions) < 1) {
		return 0, fmt.Errorf("No such policy %s", policy)
	}
	tv, ok = rs[0].Expressions[0].Value.(json.Number)

	if !ok {
		return 0, fmt.Errorf("Could not authorise action.  Return type: %s", reflect.TypeOf(rs[0].Expressions[0].Value))
	}

	return tv.Int64()
}

func runRego(ctx context.Context, query string, input map[string]interface{}) (rego.ResultSet, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "runRego")
	defer span.Finish()
	debug := len(os.Getenv("DEBUG_OPA")) > 0
	// Fetch user from context, if exists.  If not, we don't mind -- some actions will be publicly possible:

	if debug {
		jsonString, err := prettyjson.Marshal(input)

		if err != nil {
			panic(err)
		}

		fmt.Printf("rego input json for %s:\n", query)
		fmt.Println(string(jsonString))
	}

	compiler := GetCompiler(ctx)
	store := GetStore(ctx)

	compiled := getCompiledQuery(query)

	m := metrics.New()

	rego := rego.New(
		rego.ParsedQuery(compiled),
		rego.Compiler(compiler),
		rego.Input(input),
		rego.Store(store),
		rego.Metrics(m),
	)

	rs, err := rego.Eval(ctx)
	if debug {
		fmt.Println("Dumping rego.Eval metrics:", m.All())
	}
	return rs, err
}
