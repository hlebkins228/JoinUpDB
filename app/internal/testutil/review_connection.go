//go:build allure

package testutil

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"

	"JoinUp/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type reviewStore struct {
	reviews map[int]models.Review
	nextID  int
}

type ReviewConnection struct {
	store     reviewStore
	closed    bool
	QueryText string
	Args      []any
}

func NewReviewConnection() *ReviewConnection {
	c := new(ReviewConnection)
	c.Reset()
	return c
}

func (c *ReviewConnection) Reset() {
	c.store = reviewStore{reviews: make(map[int]models.Review)}
	c.QueryText = ""
	c.Args = nil
}

func (c *ReviewConnection) Close() {
	c.store = reviewStore{}
	c.QueryText = ""
	c.Args = nil
	c.closed = true
}

func (c *ReviewConnection) Closed() bool {
	return c.closed
}

func (c *ReviewConnection) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	if c.closed {
		return nil, fmt.Errorf("review connection is closed")
	}
	return &reviewTx{connection: c, store: reviewStore{
		reviews: maps.Clone(c.store.reviews),
		nextID:  c.store.nextID,
	}}, nil
}

func (c *ReviewConnection) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	c.QueryText, c.Args = query, args
	if c.closed {
		return reviewRow{err: fmt.Errorf("review connection is closed")}
	}
	return c.store.queryRow(query, args...)
}

func (c *ReviewConnection) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	if c.closed {
		return nil, fmt.Errorf("review connection is closed")
	}
	return c.store.query(query, args...)
}

func (c *ReviewConnection) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, fmt.Errorf("Exec is not used by ReviewRepo")
}

type reviewTx struct {
	pgx.Tx
	connection *ReviewConnection
	store      reviewStore
	closed     bool
}

func (tx *reviewTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if tx.closed {
		return reviewRow{err: pgx.ErrTxClosed}
	}
	return tx.store.queryRow(query, args...)
}

func (tx *reviewTx) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx.closed {
		return nil, pgx.ErrTxClosed
	}
	return tx.store.query(query, args...)
}

func (tx *reviewTx) Commit(context.Context) error {
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.connection.store = tx.store
	tx.store = reviewStore{}
	tx.closed = true
	return nil
}

func (tx *reviewTx) Rollback(context.Context) error {
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.store = reviewStore{}
	tx.closed = true
	return nil
}

func (s *reviewStore) queryRow(query string, args ...any) pgx.Row {
	query = strings.ToLower(strings.Join(strings.Fields(query), " "))
	switch {
	case strings.HasPrefix(query, "insert into review "):
		eventID, authorID := args[0].(int), args[1].(int)
		if s.exists(eventID, authorID) {
			return reviewRow{err: &pgconn.PgError{Code: "23505"}}
		}
		s.nextID++
		review := models.Review{
			ID: s.nextID, EventID: eventID, AuthorID: authorID,
			Rating: args[2].(int), Text: args[3].(string), CreatedAt: args[4].(time.Time),
		}
		s.reviews[review.ID] = review
		return reviewRow{values: []any{review.ID, review.CreatedAt}}
	case strings.HasPrefix(query, "select exists(select 1 from review "):
		return reviewRow{values: []any{s.exists(args[0].(int), args[1].(int))}}
	default:
		return reviewRow{err: fmt.Errorf("unsupported review query: %s", query)}
	}
}

func (s *reviewStore) exists(eventID, authorID int) bool {
	for _, review := range s.reviews {
		if review.EventID == eventID && review.AuthorID == authorID {
			return true
		}
	}
	return false
}

func (s *reviewStore) query(query string, args ...any) (pgx.Rows, error) {
	if !strings.Contains(strings.ToLower(query), "from review r") {
		return nil, fmt.Errorf("unsupported review query: %s", query)
	}
	reviews := make([]models.Review, 0)
	for _, review := range s.reviews {
		if review.EventID == args[0].(int) {
			reviews = append(reviews, review)
		}
	}
	slices.SortFunc(reviews, func(a, b models.Review) int {
		if cmp := b.CreatedAt.Compare(a.CreatedAt); cmp != 0 {
			return cmp
		}
		return b.ID - a.ID
	})
	rows := &reviewRows{index: -1}
	for _, review := range reviews {
		rows.values = append(rows.values, []any{
			review.ID, review.EventID, review.AuthorID, review.Rating, review.Text, review.CreatedAt,
			review.AuthorID, "", (*int)(nil),
		})
	}
	return rows, nil
}

type reviewRow struct {
	values []any
	err    error
}

func (r reviewRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("expected %d scan destinations, got %d", len(r.values), len(dest))
	}
	for i, value := range r.values {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(value))
	}
	return nil
}

type reviewRows struct {
	pgx.Rows
	values [][]any
	index  int
}

func (r *reviewRows) Next() bool {
	r.index++
	return r.index < len(r.values)
}

func (r *reviewRows) Scan(dest ...any) error {
	return (reviewRow{values: r.values[r.index]}).Scan(dest...)
}

func (r *reviewRows) Err() error {
	return nil
}

func (r *reviewRows) Close() {
	r.values = nil
}
