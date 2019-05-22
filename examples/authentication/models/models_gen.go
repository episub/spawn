// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"fmt"
	"io"
	"strconv"

	"github.com/example/todo/gnorm/public/todo"
)

type NewTodo struct {
	Text   string `json:"text"`
	UserID string `json:"userId"`
}

type PageInfo struct {
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

type TodoEdge struct {
	Cursor string    `json:"cursor"`
	Node   *todo.Row `json:"node"`
}

type TodoFilter struct {
	Done *bool `json:"done"`
}

type TodosConnection struct {
	TotalCount int         `json:"totalCount"`
	Edges      []*TodoEdge `json:"edges"`
	PageInfo   *PageInfo   `json:"pageInfo"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

var AllSortDirection = []SortDirection{
	SortDirectionAsc,
	SortDirectionDesc,
}

func (e SortDirection) IsValid() bool {
	switch e {
	case SortDirectionAsc, SortDirectionDesc:
		return true
	}
	return false
}

func (e SortDirection) String() string {
	return string(e)
}

func (e *SortDirection) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SortDirection(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SortDirection", str)
	}
	return nil
}

func (e SortDirection) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TodoSort string

const (
	TodoSortContent TodoSort = "CONTENT"
)

var AllTodoSort = []TodoSort{
	TodoSortContent,
}

func (e TodoSort) IsValid() bool {
	switch e {
	case TodoSortContent:
		return true
	}
	return false
}

func (e TodoSort) String() string {
	return string(e)
}

func (e *TodoSort) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TodoSort(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TodoSort", str)
	}
	return nil
}

func (e TodoSort) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
