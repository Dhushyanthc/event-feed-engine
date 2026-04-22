package repository

import (
	"context"

	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) CreatePost(ctx context.Context, post *models.Post) error {

	query := `INSERT INTO posts (user_id, content, media_url) Values ($1, $2, $3)
	Returning id, created_at`

	err := r.db.QueryRow(
		ctx,
		query,
		post.UserId,
		post.Content,
		post.MediaURL,
	).Scan(
		&post.Id,
		&post.CreatedAt,
	)

	return err

}

func (r *PostRepository) CreatePostTx(ctx context.Context, tx pgx.Tx, post *models.Post) error {
	query := `INSERT INTO posts (user_id, content, media_url) Values ($1, $2, $3)
	Returning id, created_at`

	return tx.QueryRow(
		ctx,
		query,
		post.UserId,
		post.Content,
		post.MediaURL,
	).Scan(
		&post.Id,
		&post.CreatedAt,
	)
}

// //////////////////////////////////////////////////
func (r *PostRepository) GetPostsByIDs(
	ctx context.Context,
	postIDs []int64,
) ([]*models.Post, error) {

	query := `
	SELECT id, user_id, content, media_url, created_at
	FROM posts
	WHERE id = ANY($1)
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, postIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post

	for rows.Next() {

		post := &models.Post{}

		err := rows.Scan(
			&post.Id,
			&post.UserId,
			&post.Content,
			&post.MediaURL,
			&post.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return posts, nil
}
