package form

import (
	"fmt"
)

type TypeName string

var (
	TypeBool   TypeName = "bool"
	TypeString TypeName = "string"
)

type Validator func(interface{}) error

type FieldDefinition struct {
	Name        string
	Group       string // Group is used when we want to break fields out into separate groups.  E.g., if creating or updating involves modifying multiple tables
	FieldType   TypeName
	Validations []Validator
}

type Definition struct {
	Fields map[string]FieldDefinition
}

func ApplyDefinition(def Definition, m map[string]interface{}) (map[string]interface{}, map[string][]error) {
	applied := make(map[string]interface{})
	// Loop over this form's definitions, enforcing fields match the type that
	// we need.  This is because sometimes a html form will submit values, like
	// a checkbox, in a format other than a bool.  We convert that value into
	// the type we expect
	errors := make(map[string][]error)
	for k, v := range m {
		var err error
		// Check if this is in our definitions
		fieldDef, ok := def.Fields[k]
		if !ok {
			continue
		}

		switch fieldDef.FieldType {
		case TypeBool:
			applied[k], err = ensureBool(v)
		case TypeString:
			applied[k], err = ensureString(v)
		default:
			err = fmt.Errorf("Unknown field type %s when converting field %s", fieldDef.FieldType, k)
		}

		if err != nil {
			// Continue, since if it's the wrong type, no point doing validations
			errors[k] = append(errors[k], err)
			continue
		}

		// Check validations:
		for _, validator := range fieldDef.Validations {
			err = validator(applied[k])
			if err != nil {
				errors[k] = append(errors[k], err)
			} else {
			}
		}
	}

	return applied, errors
}

func ensureBool(s interface{}) (interface{}, error) {
	switch v := s.(type) {
	case bool:
		return v, nil
	case string:
		var b bool
		str := s.(string)
		switch str {
		case "true", "on":
			b = true
			return b, nil
		case "false", "off":
			b = false
			return b, nil
		default:
			return nil, fmt.Errorf("field must be true or false")
		}
	default:
		return nil, fmt.Errorf("Cannot convert type %T to bool", v)
	}
}

func ensureString(s interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", s)
	return str, nil
}
