package validate

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/gofrs/uuid"

	"cloud.google.com/go/civil"
)

var numberRx = regexp.MustCompile(`^[0-9]+$`)
var emailRx = regexp.MustCompile(`^\S+@\S+$`)
var lettersWithSpacesRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ]+$`)
var lettersWithNumbersRx = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var usernameRx = regexp.MustCompile(`^[A-Za-z0-9]+(?:[_-][A-Za-z0-9]+)*$`)
var lettersSpacesAndNumbersRx = regexp.MustCompile(`^[- 'a-zA-ZÀ-ÖØ-öø-ÿ0-9]+$`)
var urlRx = regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
var aTrueRx = regexp.MustCompile(`^true$`)

// Regex Confirms that value matches the provided regex
func Regex(ctx context.Context, rx *regexp.Regexp, field string, value string, message string) bool {
	if !rx.MatchString(value) {
		AddError(ctx, field, message)
		return false
	}

	return true
}

// Email Checks that email address is valid
func Email(ctx context.Context, field string, email string) bool {
	return Regex(ctx, emailRx, field, email, fmt.Sprintf("Email address '%s' is invalid", email))
}

// Fail Used for testing, always fails and adds the message
func Fail(ctx context.Context, field string, message string) bool {
	AddError(ctx, field, message)
	return false
}

// NoError used when there's an error.  If there's an error, then it fails
func NoError(ctx context.Context, err error, field string, message string) bool {
	if err != nil {
		AddError(ctx, field, message)
		return false
	}

	return true
}

// False Checks that a value is true
func False(ctx context.Context, field string, v bool, msg string) bool {
	if v {
		AddError(ctx, field, msg)
	}

	return !v
}

// Length Requires the string to be a length between m and n inclusive
func Length(ctx context.Context, field, v string, m int, n int) bool {
	var msg string
	if m == n {
		msg = fmt.Sprintf("Must be exactly %d characters long", m)
	} else {
		msg = fmt.Sprintf("Must be between %d and %d characters long", m, n)
	}

	if len(v) < m || len(v) > n {
		AddError(ctx, field, msg)
		return false
	}

	return true
}

// MinimumLength Requires the minimum length for a string to be n characters
func MinimumLength(ctx context.Context, n int, v string, field string) bool {
	if len(v) < n {
		AddError(ctx, field, fmt.Sprintf("Must be at least %d characters long", n))
		return false
	}
	return true
}

// StringsNotEqual Verifies that the two provided values are not equal
func StringsNotEqual(ctx context.Context, field, a, b, msg string) bool {
	if a == b {
		AddError(ctx, field, msg)
		return false
	}

	return true
}

//UUID Verifies that provided string is a UUID
func UUID(ctx context.Context, field, id string, msg string) bool {
	_, err := uuid.FromString(id)

	if err != nil {
		AddError(ctx, field, msg)
		return false
	}

	return true
}

// Numbers Must be a string containing only numbers
func Numbers(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, numberRx, field, v, "Must be a number")
}

// Phone Must be a string containing only numbers
func Phone(ctx context.Context, field string, v string) bool {
	return Numbers(ctx, field, v) && Length(ctx, field, v, 7, 13)
}

// Positive Checks that an integer is positive
func Positive(ctx context.Context, field string, v int) bool {
	if v < 0 {
		AddError(ctx, field, "Must be positive")
		return false
	}

	return true
}

// Username Username validation
func Username(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, usernameRx, field, v, "Usernames can only contain letters, numbers, underscores and hyphens, and must not begin or end with an underscore or hyphen")
}

// LettersWithSpaces Must only contain letters and spaces
func LettersWithSpaces(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, lettersWithSpacesRx, field, v, "Must only contain letters and spaces")
}

// LettersWithNumbers Must be a string containing only numbers
func LettersWithNumbers(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, lettersWithNumbersRx, field, v, "Must only contain letters and numbers")
}

// LettersSpacesAndNumbers Must be a string containing only numbers
func LettersSpacesAndNumbers(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, lettersSpacesAndNumbersRx, field, v, "Must only contain letters, spaces and numbers")
}

// URL Checks that URL is valid
func URL(ctx context.Context, field string, v string) bool {
	return !Regex(ctx, urlRx, field, v, "Website URL is invalid")
}

// True Checks that a value is true
func True(ctx context.Context, field string, v bool) bool {
	if !v {
		AddError(ctx, field, fmt.Sprintf("Must be true"))
	}

	return v
}

// After Checks that a date value is after given date value to be compared with
func After(ctx context.Context, field string, v civil.Date, dateToCompareWith civil.Date) bool {
	if !v.After(dateToCompareWith) {
		AddError(ctx, field, "Date should be after "+dateToCompareWith.String())
		return false
	}

	return true
}

// AfterNow Checks that a date value is after now
func AfterNow(ctx context.Context, field string, v civil.Date) bool {
	now := civil.DateOf(time.Now().Add((time.Hour * -24)))
	return After(ctx, field, v, now)
}

// DOB Checks that a date value is before 18 years from now
func DOB(ctx context.Context, field string, v civil.Date) bool {
	now := time.Now()
	minAge := civil.Date{
		Day:   (now.Year() - 18),
		Month: now.Month(),
		Year:  now.Day(),
	}
	return After(ctx, field, v, minAge)
}

// IN Checks that a value is an item within an array
func IN(ctx context.Context, field string, v string, array []string) bool {
	for _, element := range array {
		if element == v {
			return true
		}
	}

	AddError(ctx, field, "Value not allowed")
	return false
}

// IntsEqual Verifies that the two integers are equal
func IntsEqual(ctx context.Context, field string, a, b int, msg string) bool {
	if a != b {
		AddError(ctx, field, msg)
		return false
	}

	return true
}
