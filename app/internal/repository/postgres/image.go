package postgres

import (
	"JoinUp/internal/models"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ImageRepo struct {
	Pool
}

func NewImageRepo(pool *pgxpool.Pool) ImageRepo {
	return ImageRepo{Pool: Pool{pool: pool}}
}

func (r *ImageRepo) AddImage(ctx context.Context, img *models.Image) (int, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	const query = `
		insert into image (name, data)
		values ($1, $2)
		returning id
	`

	var id int
	err = run.QueryRow(ctx, query, img.Name, img.Data).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ImageRepo) GetImage(ctx context.Context, id int) (*models.Image, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	const query = `
		select name, data
		from image
		where id = $1
	`

	var img models.Image
	err = run.QueryRow(ctx, query, id).Scan(&img.Name, &img.Data)
	if err != nil {
		return nil, err
	}

	return &img, nil
}

func (r *ImageRepo) CheckExists(ctx context.Context, id int) (bool, error) {
	run, err := r.runFromCtx(ctx)
	if err != nil {
		return false, err
	}

	const query = `select exists(select 1 from image where id = $1)`

	var ok bool
	err = run.QueryRow(ctx, query, id).Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}
