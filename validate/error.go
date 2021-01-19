package validate

import (
	"context"
	"fmt"
	"strings"

	"github.com/episub/pqt"
)

// Error Used for passing back validation errors to the resolvers, without having to know anything about graphQL ourselves
type Error struct {
	Field   string
	Message string
}

const validationValue = "validationErrors"

// AddError Adds a validation error to the context, including adding a graphql
func AddError(errors map[string][]string, field string, message string) {
	errors[field] = append(errors[field], message)
}

// DateError Checks a date value and its error, returning nil as the error and adding as a validation error if appropriate
func DateError(errors map[string][]string, field string, d pqt.Date, err error) (pqt.Date, error) {
	switch {
	case (strings.Contains(err.Error(), "parsing time") && strings.Contains(err.Error(), "cannot parse")):
		AddError(errors, field, "Invalid value for date")
		return d, nil
	default:
		return d, err
	}
}

// SetContext Adds an array pointer to the context which we use for adding validation errors to
func SetContext(ctx context.Context) context.Context {
	ve := []Error{}
	return context.WithValue(ctx, validationValue, &ve)
}

// GetErrorsFromContext Returns any validation errors stored in context
func GetErrorsFromContext(ctx context.Context) *[]Error {
	ve, ok := ctx.Value(validationValue).(*[]Error)
	if !ok {
		panic("No validation errors found in context.  Use SetContext function first to create value in context")
	}
	return ve
}

// ErrorsString Prints a string concatenating all the validation errors.
func ErrorsString(ctx context.Context) string {
	ve := GetErrorsFromContext(ctx)
	var str []string

	for _, e := range *ve {
		str = append(str, fmt.Sprintf("Field '%s': %s", e.Field, e.Message))
	}

	return strings.Join(str, ". ")
}

// HasErrors Returns true if there are any validation erros
func HasErrors(ctx context.Context) bool {
	ve := GetErrorsFromContext(ctx)
	if len(*ve) > 0 {
		return true
	}
	return false
}
