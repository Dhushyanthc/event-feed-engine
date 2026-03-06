package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
)


type UserRepository struct {
	db *pgxpool.Pool
}
func NewUserRepository(db*pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}


// create a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context,user *models.User) error {

	// query 
	query := `INSERT INTO users (name, email, password_hash) Values ($1, $2, $3 ) Returning id, created_at`

	//Queryrow 
	return r.db.QueryRow(
		ctx, 
		query, 
		user.Name, 
		user.Email, 
		user.PasswordHash).Scan(
			&user.Id,
			&user.CreatedAt,
		)
}