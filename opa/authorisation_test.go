package opa

import (
	"context"
	"log"
	"testing"
)

func init() {
	err := LoadBundle("../policies/bundle")
	if err != nil {
		panic(err)
	}
}

var authorisedCases = []struct {
	Name    string
	Policy  string
	Allowed bool
}{
	{
		Name:    "testAuthorisedPass",
		Policy:  "data.api.tests.pass.allow",
		Allowed: true,
	},
	{
		Name:    "testAuthorisedFail",
		Policy:  "data.api.tests.fail.allow",
		Allowed: false,
	},
}

func TestAuthorised(t *testing.T) {
	// We test example policies, and their expected boolean reply

	for _, c := range authorisedCases {
		allow, err := Authorised(context.Background(), c.Policy, map[string]interface{}{})

		if err != nil {
			t.Error(err)
		}

		if allow != c.Allowed {
			t.Errorf("Expected allow=%t for policy %s, but had allow=%t", c.Allowed, c.Policy, allow)
		}
	}
}

func TestAuthorisedIDs(t *testing.T) {
	expectedIDs := []string{
		"id1",
		"id2",
		"id3",
	}

	ids, err := AuthorisedStrings(context.Background(), "data.api.tests.set.allowed", map[string]interface{}{})

	if err != nil {
		t.Error(err)
		return
	}

	if len(ids) != len(expectedIDs) {
		t.Errorf("Expected %d ids but had %d", len(expectedIDs), len(ids))
	}

	found := 0

	for _, e := range expectedIDs {
		for _, a := range ids {
			if e == a {
				found++
				continue
			}
		}
	}

	if found != len(expectedIDs) {
		t.Errorf("Could not find all expected ID's")
	}
}

var permissionCases = []struct {
	Permission string
	Allowed    bool
}{
	{
		Permission: "fail",
		Allowed:    false,
	},
	{
		Permission: "pass",
		Allowed:    true,
	},
	{
		// There is no policy for this, so should return false
		Permission: "empty",
		Allowed:    false,
	},
}

func TestAuthorisedPermissions(t *testing.T) {
	var includePermissions []string

	for _, p := range permissionCases {
		includePermissions = append(includePermissions, p.Permission)
	}

	permissions, err := AuthorisedPermissions(context.Background(), includePermissions, "data.api.tests", nil, map[string]interface{}{})

	if err != nil {
		t.Error(err)
		return
	}

	for _, e := range permissionCases {
		var found bool

		for _, p := range permissions {
			if p.Name == e.Permission && p.Value == e.Allowed {
				found = true
			}
		}

		if !found {
			log.Printf("Could not find permission %s", e.Permission)
		}
	}
}

func TestGetInt(t *testing.T) {
	// We test example policies, and their expected boolean reply
	policy := "data.api.tests.int.limit"
	var expected int64 = 5
	v, err := GetInt(context.Background(), policy, map[string]interface{}{})

	if err != nil {
		t.Error(err)
	}

	if v != expected {
		t.Errorf("Expected int=%d for policy %s, but had %d", expected, policy, v)
	}
}
