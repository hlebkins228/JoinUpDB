package postgres

import (
	"JoinUp/internal/exceptions"
	"JoinUp/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolationCode = "23505"

type ReviewRepo struct {
	Pool
}

func NewReviewRepo(pool *pgxpool.Pool) ReviewRepo {
	return ReviewRepo{Pool: Pool{pool: pool}}
}

func (r *ReviewRepo) Create(ctx context.Context, rv *models.Review) (*models.Review, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		insert into review (event_id, author_id, rating, text, created_at)
		values ($1, $2, $3, $4, $5)
		returning id, created_at at time zone 'utc'
	`

	err = run.QueryRow(ctx, query,
		rv.EventID,
		rv.AuthorID,
		rv.Rating,
		rv.Text,
		rv.CreatedAt,
	).Scan(&rv.ID, &rv.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, fmt.Errorf("%w: %v", exceptions.ErrAlreadyExists, err)
		}
		return nil, err
	}

	return rv, nil
}

func (r *ReviewRepo) ListByEvent(ctx context.Context, eventID int) ([]*models.Review, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select r.id, r.event_id, r.author_id, r.rating, r.text,
		       r.created_at at time zone 'utc',
		       u.id, u.name, u.avatar_id
		from review r
		join users u on u.id = r.author_id
		where r.event_id = $1
		order by r.created_at desc, r.id desc
	`

	rows, err := run.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]*models.Review, 0)
	for rows.Next() {
		review := models.Review{}
		err = rows.Scan(
			&review.ID,
			&review.EventID,
			&review.AuthorID,
			&review.Rating,
			&review.Text,
			&review.CreatedAt,
			&review.Author.ID,
			&review.Author.Name,
			&review.Author.AvatarID,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return reviews, nil
}

func (r *ReviewRepo) ExistsByEventAndAuthor(ctx context.Context, eventID, authorID int) (bool, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	const query = `select exists(select 1 from review where event_id = $1 and author_id = $2)`

	var ok bool
	err = run.QueryRow(ctx, query, eventID, authorID).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}
