package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
)


type UserRepository struct {
	db *pgxpool.Pool
}
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
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



//new method 
///////////////////////////////////////////
//retrieve user by email
func (r *UserRepository) GetUserByEmail (ctx context.Context, email string) (User *models.User, err error){

	if email == ""{
		return nil, errors.New("email is required")
	}


	// query
	query:= `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`

	User = &models.User{}

	Err := r.db.QueryRow(ctx, query, email).Scan(
		&User.Id,
		&User.Name,
		&User.Email,
		&User.PasswordHash,
		&User.CreatedAt,
	)

	if Err != nil {
		return nil, Err
	}


	return User, nil
}