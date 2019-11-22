package gnorm

import (
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
)

// InString Returns an in clause for an array of strings
func InString(field string, values []string) In {
	newVals := make([]interface{}, len(values))

	for i, v := range values {
		newVals[i] = v
	}

	return In{Field: field, Values: newVals}
}

// InInt Returns an in clause for an array of integers
func InInt(field string, values []int) In {
	newVals := make([]interface{}, len(values))

	for i, v := range values {
		newVals[i] = v
	}

	return In{Field: field, Values: newVals}
}

// In Return a clause for values in an array
type In struct {
	Field  string
	Values []interface{}
}

// InUUIDUUID Returns an in clause for an array of uuid's
func InUUIDUUID(field string, values []uuid.UUID) In {
	newVals := make([]interface{}, len(values))

	for i, v := range values {
		newVals[i] = v
	}

	return In{Field: field, Values: newVals}
}

// ToSql Returns the sql related objects expected by squirrel
func (in In) ToSql() (sql string, args []interface{}, err error) {
	questions := make([]string, len(in.Values))

	for i := range in.Values {
		questions[i] = "?"
	}

	// No entries:
	if len(in.Values) == 0 {
		sql = fmt.Sprintf("false")
		return
	}

	// Some values, so build it out:
	sql = fmt.Sprintf("%s in (%s)", in.Field, strings.Join(questions, ", "))

	return sql, in.Values, nil
}
