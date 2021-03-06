package models

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/example/todo/gnorm"
)

// Filter Details on order and filters for a particular search request
type Filter struct {
	Cursor *string      // If provided, returns results from the cursor
	Count  int64        // If 0, returns all
	Before bool         // If true, returns count results before cursor, otherwise count results after cursor
	Where  []sq.Sqlizer // Filters to apply
	Order  gnorm.Order  // Ordering of fields
}

// NewFilter Returns new filter based on graphql values passed into it
func NewFilter(first *int, after *string, last *int, before *string, direction *SortDirection) Filter {
	var f Filter

	if first != nil {
		f.Count = int64(*first)
	} else if last != nil {
		f.Count = int64(*last)
		f.Before = true
	}

	if before != nil {
		f.Cursor = before
		f.Before = true
	}

	if after != nil {
		f.Cursor = after
	}

	if direction != nil && *direction == SortDirectionDesc {
		f.Order.Descending = true
	}

	return f
}
