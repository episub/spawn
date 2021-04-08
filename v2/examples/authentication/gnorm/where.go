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
func (i In) ToSql() (sql string, args []interface{}, err error) {
	questions := make([]string, len(i.Values))

	for i := range i.Values {
		questions[i] = "?"
	}

	sql = fmt.Sprintf("%s in (%s)", i.Field, strings.Join(questions, ", "))

	return sql, i.Values, nil
}
