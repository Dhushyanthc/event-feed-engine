package handlers

import (
	"net/http"
	"encoding/json"
	

	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"github.com/Dhushyanthc/event-feed-engine/internal/auth"
)


type LoginHandler struct {
	repo      *repository.UserRepository
	jwtSecret string
}
func NewLoginHandler(repo *repository.UserRepository, jwtSecret string) *LoginHandler {
	return &LoginHandler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}


func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// parse the request body to get email and password
	var req struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return 
	}
	if req.Email == "" || req.Password == ""{
		http.Error(w,"Email and password are required", http.StatusBadRequest	)
		return
	}

	// retrieve the user from the database using the email
	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return 
	}

	// compare the provided password with the stored password hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	// if the password is incorrect, return an error
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return 
	}

	//generate a JWT token
	token, err := auth.GenerateJWT(user.Id, h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// if the password is correct, generate a JWT token and return it in the response
	resp := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{
		ID:   user.Id,
		Name: user.Name,
		Email: user.Email,
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	
	

}
