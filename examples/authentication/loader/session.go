package loader

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/example/todo/gnorm/public/session"
	"github.com/gofrs/uuid"
)

// CreateSession Creates a new session
func (l *PostgresLoader) CreateSession(ctx context.Context, userID int, expiry time.Time) (uuid.UUID, error) {
	var err error
	var i session.Row
	i.Expires = expiry
	i.UserID = userID

	i, err = session.Upsert(ctx, l.pool, i)

	return i.SessionID, err
}

// DeleteSession Marks a session as expired as of now
func (l *PostgresLoader) DeleteSession(ctx context.Context, sessionID string) (bool, error) {
	_, err := session.Update(
		ctx,
		l.pool,
		map[string]interface{}{"expires": time.Now()},
		[]sq.Sqlizer{sq.Eq{session.SessionIDCol: sessionID}},
	)

	return (err == nil), sanitiseError(err)
}

func hydrateModelSession(ctx context.Context, i session.Row) (o session.Row) {
	return i
}
