package integration_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConstraints(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		code       string
		constraint string
		countQuery string
		count      int
	}{
		{
			name: "duplicate login",
			query: `insert into users (id, name, age, login, password, created_at, city)
				values (2, 'duplicate', 22, 'admin', 'hash', now(), 'Moscow')`,
			code:       "23505",
			constraint: "users_login_unique",
			countQuery: `select count(*) from users`,
			count:      1,
		},
		{
			name:       "missing location",
			query:      `update event set location_id = 9999`,
			code:       "23503",
			constraint: "event_location_id_fkey",
			countQuery: `select count(*) from event e join location l on l.id = e.location_id`,
			count:      1,
		},
		{
			name: "invalid rating",
			query: `insert into review (event_id, author_id, rating, text)
				select id, 1, 0, 'Review' from event`,
			code:       "23514",
			constraint: "review_rating_check",
			countQuery: `select count(*) from review`,
			count:      0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := prepareTestDB(t)
			seedEvent(t, ctx)

			_, err := testApp.pool.Exec(ctx, tc.query)
			var pgErr *pgconn.PgError
			require.ErrorAs(t, err, &pgErr)
			assert.Equal(t, tc.code, pgErr.Code)
			assert.Equal(t, tc.constraint, pgErr.ConstraintName)

			var count int
			require.NoError(t, testApp.pool.QueryRow(ctx, tc.countQuery).Scan(&count))
			assert.Equal(t, tc.count, count)
		})
	}
}
