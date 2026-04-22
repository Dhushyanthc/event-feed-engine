package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dhushyanthc/event-feed-engine/internal/middleware"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
)

type PostHandler struct {
	repo      *repository.PostRepository
	eventRepo *repository.EventRepository
}

func NewPostHandler(repo *repository.PostRepository, eventRepo *repository.EventRepository) *PostHandler {
	return &PostHandler{
		repo:      repo,
		eventRepo: eventRepo,
	}
}

type PostRequest struct {
	Content  string `json:"content"`
	MediaURL string `json:"media_url"`
}

type PostResponse struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url"`
	CreatedAt string `json:"created_at"`
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" && req.MediaURL == "" {
		http.Error(w, "Content or MediaURL is required", http.StatusBadRequest)
		return
	}

	//i think we need to write a middle ware to upload the image to the cloudinary and get the url and then it to the post -- later

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	post := &models.Post{
		UserId:   userID,
		Content:  req.Content,
		MediaURL: req.MediaURL,
	}

	tx, err := h.eventRepo.BeginTx(r.Context())
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	err = h.repo.CreatePostTx(r.Context(), tx, post)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	err = h.eventRepo.CreateEventTx(r.Context(), tx, post.Id, post.UserId, post.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to enqueue fanout job", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	resp := PostResponse{
		Id:        post.Id,
		UserId:    post.UserId,
		Content:   post.Content,
		MediaURL:  post.MediaURL,
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
