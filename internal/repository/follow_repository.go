package repository

import (
	"context"
	"fmt"

	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowRepository struct {
	db *pgxpool.Pool
}
func NewFollowRepository(db *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) CreateFollow(ctx context.Context, follow *models.Follow) error{

	query := `INSERT INTO followers (follower_id, following_id) VALUES ($1, $2) RETURNING created_at`

	err := r.db.QueryRow(
		ctx,
		query,
		follow.FollowerID,
		follow.FollowingID,
	).Scan(
		&follow.CreatedAt,	
	)
	return err
}

func (r *FollowRepository) DeleteFollow(ctx context.Context, follow *models.Follow) error{
	query := `DELETE FROM followers WHERE follower_id = $1 and following_id = $2`

	cmd, err := r.db.Exec(
		ctx,
		query,
		follow.FollowerID,
		follow.FollowingID,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("follow relationship not found")
	}
	return nil
}

////////////////////////////////////////////
func (r *FollowRepository) GetFollowers( ctx context.Context, userID int64, limit, offset int) ([]int64, error){

	query := `SELECT follower_id FROM followers WHERE following_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows,err := r.db.Query(ctx, query, userID, limit, offset) 
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	


	var followers []int64
	for rows.Next(){
		var follow int64
		err := rows.Scan(&follow)
		if err != nil {
			return nil, err
		}
		followers = append(followers, follow)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return followers, nil
}


func (r *FollowRepository) GetFollowing(ctx context.Context, userID int64, limit, offset int) ([]int64, error){

	query := `SELECT following_id FROM followers WHERE follower_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()	

	var following []int64
	for rows.Next(){
		var follow int64
		err := rows.Scan(&follow)
		if err != nil {
			return nil, err
		}
		following = append(following, follow)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	
	return following, nil
}

////////////////////////////////////////////////