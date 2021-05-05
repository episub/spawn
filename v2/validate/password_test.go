package validate

import (
	"context"
	"testing"
)

type passwordTest struct {
	Validator PasswordValidator
	Password  string
	Accept    bool
}

var PasswordUpperLowerNumberTests = []passwordTest{
	{
		Password: "oijo",
		Accept:   false,
	},
	{
		Password: "oiJo",
		Accept:   false,
	},
	{
		Password: "2ijo",
		Accept:   false,
	},
	{
		Password: "2iJo",
		Accept:   true,
	},
	{
		Password: "2i Jo",
		Accept:   true,
	},
	{
		Password: "2iJo$",
		Accept:   true,
	},
	{
		Password: "OIJoijoij123",
		Accept:   true,
	},
}

var passwordTests = []struct {
	Tests     []passwordTest
	Validator PasswordValidator
	Name      string
	Length    uint
}{
	{
		PasswordUpperLowerNumberTests,
		PasswordUpperLowerNumber,
		"PasswordUpperLowerNumber",
		4,
	},
}

func TestPassword(t *testing.T) {
	for _, tc := range passwordTests {
		for _, tp := range tc.Tests {
			t.Run(tc.Name, func(t *testing.T) {
				ctx := context.Background()
				ctx = SetContext(ctx)
				res := Password(ctx, tp.Password, tc.Length, tc.Validator, "password")
				if res != tp.Accept {
					t.Errorf("Expected password '%s' match %t, but had %t: %s", tp.Password, tp.Accept, res, ErrorsString(ctx))
				}
			})
		}
	}
}
