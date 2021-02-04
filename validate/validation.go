package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/episub/spawn/form"
	"github.com/gofrs/uuid"
)

var ErrMustBeString = fmt.Errorf("Field must be a string")
var ErrMustBeBool = fmt.Errorf("Field must be a boolean")
var ErrMustBeInt = fmt.Errorf("Field must be an integer")

var numberRx = regexp.MustCompile(`^-?\d*$`)
var emailRx = regexp.MustCompile(`^\S+@\S+$`)
var lettersWithSpacesRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ]+$`)
var lettersWithNumbersRx = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var usernameRx = regexp.MustCompile(`^[A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*$`)
var lettersSpacesAndNumbersRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ0-9]+$`)
var urlRx = regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
var aTrueRx = regexp.MustCompile(`^true$`)

// OrChain Chains two or more validations in an or arrangement
func OrChain(validators ...form.Validator) form.Validator {
	return func(s interface{}) error {
		var errors []string
		for _, vd := range validators {
			err := vd(s)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				// Just need one in chain to pass for this to be fine
				return nil
			}
		}

		// Happens if there are no validators
		if len(errors) == 0 {
			return nil
		}

		return fmt.Errorf(strings.Join(errors, " or "))
	}
}

// Regex Confirms that value matches the provided regex
func Regex(rx *regexp.Regexp, message string) form.Validator {
	return func(s interface{}) error {
		value, ok := s.(string)
		if !ok {
			return fmt.Errorf("Field must be string")
		}
		if !rx.MatchString(value) {
			return fmt.Errorf(message)
		}

		return nil
	}
}

// Email Checks that email address is valid
func Email() form.Validator {
	return Regex(emailRx, "Email address is invalid")
}

// Fail Used for testing, always fails and adds the message
func Fail(message string) form.Validator {
	return func(v interface{}) error {
		return fmt.Errorf(message)
	}
}

// NoError used when there's an error.  If there's an error, then it fails
func NoError(message string) form.Validator {
	return func(v interface{}) error {
		err, ok := v.(error)
		if !ok {
			return fmt.Errorf("Value must be an error")
		}
		if err != nil {

			return fmt.Errorf("message")
		}

		return nil
	}
}

// False Checks that a value is true
func False(msg string) form.Validator {
	return func(v interface{}) error {
		b, ok := v.(bool)
		if !ok {
			return fmt.Errorf("Value must be a bool")
		}

		if b {
			return fmt.Errorf(msg)
		}

		return nil
	}
}

// Length Requires the string to be a length between m and n inclusive
func Length(m int, n int) form.Validator {
	return func(v interface{}) error {
		field, ok := v.(string)
		if !ok {
			return ErrMustBeString
		}

		var msg string
		if m == n {
			msg = fmt.Sprintf("Must be exactly %d characters long", m)
		} else {
			msg = fmt.Sprintf("Must be between %d and %d characters long", m, n)
		}

		if len(field) < m || len(field) > n {
			return fmt.Errorf(msg)
		}

		return nil
	}
}

// MinimumLength Requires the minimum length for a string to be n characters
func MinimumLength(n int) form.Validator {
	return func(s interface{}) error {
		v, ok := s.(string)
		if !ok {
			return ErrMustBeString
		}
		if len(v) < n {
			return fmt.Errorf("Must be at least %d characters long", n)
		}
		return nil
	}
}

//UUID Verifies that provided string is a UUID
func UUID() form.Validator {
	return func(s interface{}) error {
		v, ok := s.(string)
		if !ok {
			return ErrMustBeString
		}
		_, err := uuid.FromString(v)

		if err != nil {
			return fmt.Errorf("Expected UUID for field")
		}

		return nil
	}
}

// Numbers Must be a string containing only numbers
func Numbers() form.Validator {
	return Regex(numberRx, "Must be numbers only")
}

// Positive Checks that an integer is positive
func Positive() form.Validator {
	return func(s interface{}) error {
		v, ok := s.(int)
		if !ok {
			return ErrMustBeInt
		}
		if v < 0 {
			return fmt.Errorf("Must be positive")
		}

		return nil
	}
}

// Username Username validation
func Username() form.Validator {
	return Regex(usernameRx, "Usernames can only contain letters, numbers, underscores and hyphens, and must not begin or end with an underscore or hyphen")
}

// LettersWithSpaces Must only contain letters and spaces
func LettersWithSpaces() form.Validator {
	return Regex(lettersWithSpacesRx, "Must only contain letters and spaces")
}

// LettersWithNumbers Must be a string containing only numbers
func LettersWithNumbers() form.Validator {
	return Regex(lettersWithNumbersRx, "Must only contain letters and numbers")
}

// LettersSpacesAndNumbers Must be a string containing only numbers
func LettersSpacesAndNumbers() form.Validator {
	return Regex(lettersSpacesAndNumbersRx, "Must only contain letters, spaces and numbers")
}

// URL Checks that URL is valid
func URL() form.Validator {
	return Regex(urlRx, "Website URL is invalid")
}

// True Checks that a value is true
func True(msg string) form.Validator {
	return func(s interface{}) error {
		v, ok := s.(bool)
		if !ok {
			return ErrMustBeBool
		}

		if !v {
			return fmt.Errorf(msg)
		}

		return nil
	}
}

// // After Checks that a date value is after given date value to be compared with
// func After(field string, errors map[string][]string, v civil.Date, dateToCompareWith civil.Date) form.Validator {
// 	return func(v interface{}) error {
// 	if !v.After(dateToCompareWith) {
// 		AddError(errors, field, "Date should be after "+dateToCompareWith.String())
// 		return false
// 	}
//
// 	return true
// }
//
// // AfterNow Checks that a date value is after now
// func AfterNow(field string, errors map[string][]string, v civil.Date) form.Validator {
// 	return func(v interface{}) error {
// 	now := civil.DateOf(time.Now().Add((time.Hour * -24)))
// 	return After(field, errors, v, now)
// }
//
// // DOB Checks that a date value is before 18 years from now
// func DOB(field string, errors map[string][]string, v civil.Date) form.Validator {
// 	return func(v interface{}) error {
// 	now := time.Now()
// 	minAge := civil.Date{
// 		Day:   (now.Year() - 18),
// 		Month: now.Month(),
// 		Year:  now.Day(),
// 	}
// 	return After(field, errors, v, minAge)
// }
//
// // IN Checks that a value is an item within an array
// func IN(field string, errors map[string][]string, v string, array []string) form.Validator {
// 	return func(v interface{}) error {
// 	for _, element := range array {
// 		if element == v {
// 			return true
// 		}
// 	}
//
// 	AddError(errors, field, "Value not allowed.  Must be one of: "+strings.Join(array, ", "))
// 	return false
// }
//
// // IntsEqual Verifies that the two integers are equal
// func IntsEqual(field string, errors map[string][]string, a, b int, msg string) form.Validator {
// 	return func(v interface{}) error {
// 	if a != b {
// 		AddError(errors, field, msg)
// 		return false
// 	}
//
// 	return true
// }
// }
