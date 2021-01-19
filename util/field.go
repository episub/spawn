package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/episub/pqt"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
)

func GetNullString(val *string) (ns sql.NullString) {
	if val == nil {
		return
	}

	ns.String = *val
	ns.Valid = true

	return
}

func GetNullUUID(val *uuid.UUID) (ns uuid.NullUUID) {
	if val == nil {
		return
	}

	if *val == uuid.Nil {
		return
	}

	ns.UUID = *val
	ns.Valid = true

	return
}

func GetPointerString(val sql.NullString) *string {
	if !val.Valid {
		return nil
	}

	return &val.String
}

func GetPointerStringFromUUID(val uuid.NullUUID) *string {
	if !val.Valid {
		return nil
	}

	s := val.UUID.String()

	return &s
}

func GetNullTime(val *time.Time) (ns pq.NullTime) {
	if val == nil {
		return
	}

	ns.Time = *val
	ns.Valid = true

	return
}

func GetNullDate(val *string) (n pqt.NullDate, err error) {
	if val != nil {
		n.Date, err = civil.ParseDate(*val)
		if err != nil {
			return
		}
		n.Valid = true
	}

	return
}

func GetPointerDate(val pqt.NullDate) *civil.Date {
	if !val.Valid {
		return nil
	}

	a := val.Date

	return &a
}

func GetPointerDateString(val pqt.NullDate) *string {
	if !val.Valid {
		return nil
	}

	str := val.Date.String()
	return &str
}

// Functions to help quickly convert from map to required values:

func MustBool(v interface{}) (o bool, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	o, ok := v.(bool)

	if !ok {
		err = fmt.Errorf("Expected bool for value, but had %s", reflect.TypeOf(v))
		return
	}

	return
}

func MustFloat64(v interface{}) (o float64, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	switch reflect.TypeOf(v).String() {
	case "float64":
		o, _ = v.(float64)
	case "json.Number":
		j, _ := v.(json.Number)
		o, err = j.Float64()
	default:
		err = fmt.Errorf("Expected float64 for value, but had %s", reflect.TypeOf(v))
	}

	return
}

func MustInt(v interface{}) (o int, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	switch reflect.TypeOf(v).String() {
	case "int":
		o, _ = v.(int)
	case "int64":
		o64, _ := v.(int64)
		o = int(o64)
	case "json.Number":
		j, _ := v.(json.Number)
		var i int64
		i, err = j.Int64()
		o = int(i)
	default:
		err = fmt.Errorf("Expected int for value, but had %s", reflect.TypeOf(v))
	}

	return
}

func MustNullInt(v interface{}) (o sql.NullInt64, err error) {
	if v == nil {
		return
	}

	switch reflect.TypeOf(v).String() {
	case "int":
		n, _ := v.(int)
		o.Int64 = int64(n)
	case "int64":
		o.Int64, _ = v.(int64)
	default:
		err = fmt.Errorf("Expected int or int64 for value, but had %s", reflect.TypeOf(v))
	}

	return
}

func MustNullTime(v interface{}) (o pq.NullTime, err error) {
	if v == nil {
		return
	}

	str, ok := v.(string)
	if !ok {
		return o, fmt.Errorf("Expected string for time")
	}

	if len(str) == 0 {
		return
	}

	var t time.Time
	t, err = MustTime(v)

	o.Time = t
	o.Valid = true
	return
}

func MustString(v interface{}, trim bool) (o string, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	o, ok := v.(string)

	if !ok {
		err = fmt.Errorf("Expected string for value, but had %s", reflect.TypeOf(v))
		return
	}

	if trim {
		o = strings.TrimSpace(o)
	}

	return
}

func MustUUID(v interface{}) (o uuid.UUID, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	switch reflect.TypeOf(v).String() {
	case "string":
		var s string
		s, err = MustString(v, false)

		if err != nil {
			return
		}

		return uuid.FromString(s)
	case "uuid.UUID":
		return v.(uuid.UUID), nil
	default:
		err = fmt.Errorf("Cannot convert field to uuid")
	}

	return
}

func MustTime(v interface{}) (o time.Time, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	str, ok := v.(string)

	if !ok {
		err = fmt.Errorf("Expected time for value, but had %s", reflect.TypeOf(v))
		return
	}

	o, err = time.Parse(time.RFC3339, str)

	return
}

func MustDate(v interface{}) (o pqt.Date, err error) {
	if v == nil {
		err = fmt.Errorf("Cannot be nil")
		return
	}

	str, ok := v.(string)
	if !ok {
		err = fmt.Errorf("Expected string date for value, but had %s", reflect.TypeOf(v))
		return
	}

	o.Date, err = civil.ParseDate(str)
	if err != nil {
		return
	}
	return
}

func MustNullBool(v interface{}) (o sql.NullBool, err error) {
	if v == nil {
		return
	}

	t, ok := v.(bool)
	if !ok {
		err = fmt.Errorf("Expected null or bool for value, but had %s", reflect.TypeOf(v))
		return
	}

	o.Valid = true
	o.Bool = t
	return
}
func MustNullDate(v interface{}) (o pqt.NullDate, err error) {
	if v == nil {
		return
	}

	str, ok := v.(string)
	if !ok {
		err = fmt.Errorf("Expected null or date for value, but had %s", reflect.TypeOf(v))
		return
	}

	if len(str) == 0 {
		o.Valid = false
		return
	}

	o.Date, err = civil.ParseDate(str)
	if err != nil {
		return
	}
	o.Valid = true
	return
}

func MustNullString(v interface{}, trim bool) (o sql.NullString, err error) {
	if v == nil {
		return
	}

	str, ok := v.(string)
	if !ok {
		err = fmt.Errorf("Expected null or string for value, but had %s", reflect.TypeOf(v))
		return
	}

	if trim {
		o.String = strings.TrimSpace(str)
	}
	o.Valid = len(str) > 0
	return
}

func MustNullUUID(v interface{}) (o uuid.NullUUID, err error) {
	if v == nil {
		return
	}

	o.UUID, err = MustUUID(v)
	if err != nil {
		return o, err
	}

	o.Valid = true
	return
}

// func MustValidateDate(ctx context.Context, path string, v interface{}) (pqt.Date, error) {
// 	d, err := MustDate(v)
// 	if err != nil {
// 		d, err = validate.DateError(ctx, path, d, err)
// 	}
//
// 	return d, err
// }
