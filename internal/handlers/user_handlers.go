package handlers

import (
	"encoding/json"
	"net/http"
	"time"


	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo *repository.UserRepository
}
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}


// request user struct 
type createUserRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

//response user struct 
type userResponse struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Email string 	`json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// create user handler
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request){

	// check if the request method is POST
	if r.Method != http.MethodPost{
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// decode the request body
	var req createUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//hash the password 
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// create a new user model
	user := &models.User{
		Name: req.Name,
		Email: req.Email,
		PasswordHash: string(hash),
	}

	// create the user in the database
	err = h.repo.CreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	//response 
	resp := userResponse{
		ID: user.Id,
		Name: user.Name,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
	}
	
	// return the created user as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}



