package repository

import (
	"context"

	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedRepository struct {
	db *pgxpool.Pool
}
func NewFeedRepository(db *pgxpool.Pool) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) InsertFeedItem(ctx context.Context, feedItem *models.FeedItem) error{

	query := `INSERT INTO user_feed (user_id, post_id, created_at) VALUES ($1, $2, $3)`

	_, err := r.db.Exec(
		ctx,
		query,
		feedItem.UserID,
		feedItem.PostID,
		feedItem.CreatedAt,
	)
	
	if err != nil {
		return err
	}
	return nil
}


//////////////////////////////////////////////////
func (r *FeedRepository) GetFeed(ctx context.Context, userID int64, limit int, offset int) ([]int64, error) {

	query := `SELECT post_id FROM user_feed WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedItems []int64
	for rows.Next() {
		var feedItem int64
		err := rows.Scan(&feedItem)
		if err != nil {
			return nil, err
		}	
		feedItems = append(feedItems, feedItem)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return feedItems, nil
}