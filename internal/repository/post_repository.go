package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
)

type PostRepository struct {
	db *pgxpool.Pool
}
func NewPostRepository (db *pgxpool.Pool) *PostRepository{
	return &PostRepository{db:db}
}

func (r *PostRepository) CreatePost(ctx context.Context, post *models.Post) error{

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