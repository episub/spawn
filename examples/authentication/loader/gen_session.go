// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots
package loader

import (
	"context"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/example/todo/gnorm"
	"github.com/example/todo/gnorm/public/session"
	"github.com/example/todo/models"
	"github.com/gofrs/uuid"
	opentracing "github.com/opentracing/opentracing-go"
)

// SessionFetchRequest A request for a session object, to be batched
type SessionFetchRequest struct {
	SessionID uuid.UUID
	Reply     chan SessionFetchReply
}

// SessionFetchReply A reply with the requested object or an error
type SessionFetchReply struct {
	Row   session.Row
	Error error
}

var sessionInitialised bool
var sessionFRs []SessionFetchRequest
var sessionMX sync.RWMutex

// OneSession Returns a single Session with the given where clauses and order
func (l *PostgresLoader) OneSession(ctx context.Context, where []sq.Sqlizer, order *gnorm.Order) (o session.Row, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "OneSession")
	defer span.Finish()

	r, err := session.One(ctx, l.pool, where, order)

	o = hydrateModelSession(ctx, r)

	return o, err
}

// GetSession Returns Session with given ID
func (l *PostgresLoader) GetSession(ctx context.Context, id uuid.UUID) (o session.Row, err error) {
	return l.getSession(ctx, id, l.pool)
}

// getSession Returns Session with given ID, using provided DB connection
func (l *PostgresLoader) getSession(ctx context.Context, id uuid.UUID, db gnorm.DB) (o session.Row, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetSession")
	defer span.Finish()

	r, err := l.batchedGetSession(id, l.pool)

	if err != nil {
		err = sanitiseError(err)
		return
	}

	o = hydrateModelSession(ctx, r)

	return
}

func (l *PostgresLoader) batchedGetSession(id uuid.UUID, db gnorm.DB) (o session.Row, err error) {
	sessionMX.RLock()
	if !sessionInitialised {
		err = fmt.Errorf("batchedGetSession not initialised.  Add 'go loader.runSessionBatcher()' to init")
	}
	sessionMX.RUnlock()
	if err != nil {
		return
	}

	rchan := make(chan SessionFetchReply)
	r := SessionFetchRequest{
		SessionID: id,
		Reply:     rchan,
	}

	sessionMX.Lock()
	sessionFRs = append(sessionFRs, r)
	sessionMX.Unlock()

	reply := <-rchan

	return reply.Row, reply.Error
}

func (l *PostgresLoader) runSessionBatcher() {
	sessionMX.Lock()
	sessionInitialised = true
	sessionMX.Unlock()
	for {
		time.Sleep(time.Millisecond * 20)

		sessionMX.Lock()
		if len(sessionFRs) > 0 {
			var sessions []session.Row
			var err error
			var ids []uuid.UUID

			for _, r := range sessionFRs {
				ids = append(ids, r.SessionID)
			}

			sessions, err = session.Query(context.Background(), l.pool, []sq.Sqlizer{gnorm.InUUIDUUID(session.SessionIDCol, ids)})

		OUTER:
			for _, r := range sessionFRs {
				for _, c := range sessions {
					if c.SessionID == r.SessionID {
						r.Reply <- SessionFetchReply{Row: c, Error: nil}
						continue OUTER
					}
				}

				err2 := err

				if err2 == nil {
					err2 = fmt.Errorf("Not found")
				}
				r.Reply <- SessionFetchReply{Error: err2}
			}

			sessionFRs = []SessionFetchRequest{}
		}

		sessionMX.Unlock()
	}
}

// GetAllSession Returns an array of all Session entries, using the provided filter
// For an explanation on how the query works, and reversing orders, etc:
// https://use-the-index-luke.com/sql/partial-results/fetch-next-page
func (l *PostgresLoader) GetAllSession(ctx context.Context, filter models.Filter) (all []session.Row, pi models.PageInfo, count int, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetAllSession")
	defer span.Finish()

	descending := filter.Order.Descending
	// If filter.Before, we reverse the order of the results now:
	if filter.Before {
		filter.Order.Descending = !descending
	}

	r, hasMore, count, err := session.QueryPaginated(ctx, l.pool, filter.Cursor, filter.Where, filter.Order, filter.Count)

	if err != nil {
		return
	}

	// We may need to reverse the order back again if we swapped it:
	if descending != filter.Order.Descending {
		// Restore the order
		for i := len(r)/2 - 1; i >= 0; i-- {
			opp := len(r) - 1 - i
			r[i], r[opp] = r[opp], r[i]
		}
	}

	if filter.Before {
		pi.HasPreviousPage = hasMore
		if filter.Cursor != nil {
			pi.HasNextPage = true
		}
	} else {
		pi.HasNextPage = hasMore
		if filter.Cursor != nil {
			pi.HasPreviousPage = true
		}
	}

	all = make([]session.Row, len(r))
	for i, b := range r {
		all[i] = hydrateModelSession(ctx, b)
	}

	return
}
