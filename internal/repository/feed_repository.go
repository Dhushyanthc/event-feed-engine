package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type FeedRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewFeedRepository(db *pgxpool.Pool, redis *redis.Client) *FeedRepository {
	return &FeedRepository{
		db:    db,
		redis: redis,
	}
}

func (r *FeedRepository) InsertFeedItem(ctx context.Context, feedItem *models.FeedItem) error {

	query := `INSERT INTO user_feed (user_id, post_id, created_at) VALUES ($1, $2, $3) ON CONFLICT (user_id, post_id) DO NOTHING`

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

	key := "feed:user:" + strconv.FormatInt(feedItem.UserID, 10)

	if err := r.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(feedItem.CreatedAt.Unix()),
		Member: feedItem.PostID,
	}).Err(); err == nil {

		// trim feed
		r.redis.ZRemRangeByRank(ctx, key, 0, -1001)
	}
	return nil
}

// ////////////////////////////////////////////////
func (r *FeedRepository) GetFeed(ctx context.Context, userID int64, limit int, offset int) ([]int64, error) {

	key := "feed:user:" + strconv.FormatInt(userID, 10)

	cachedIDs, err := r.redis.ZRevRange(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, err
	}

	if len(cachedIDs) > 0 {
		//cache hit
		var ids []int64

		for _, idStr := range cachedIDs {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return nil, err
			}

			ids = append(ids, id)
		}

		return ids, nil
	}

	query := `SELECT post_id, created_at FROM user_feed WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedItems []int64
	var redisMembers []redis.Z
	for rows.Next() {
		var postID int64
		var createdAt time.Time
		err := rows.Scan(&postID, &createdAt)
		if err != nil {
			return nil, err
		}

		redisMembers = append(redisMembers, redis.Z{
			Score:  float64(createdAt.Unix()),
			Member: postID,
		})
		feedItems = append(feedItems, postID)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if offset == 0 && len(redisMembers) > 0 {
		err = r.redis.ZAdd(ctx, key, redisMembers...).Err()
		if err == nil {
			r.redis.ZRemRangeByRank(ctx, key, 0, -1001)
		}

	}

	return feedItems, nil

}
