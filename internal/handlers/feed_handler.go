package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dhushyanthc/event-feed-engine/internal/repository"

	"github.com/Dhushyanthc/event-feed-engine/internal/middleware"
)

type FeedHandler struct {
	feedRepo *repository.FeedRepository
	postRepo *repository.PostRepository
}
func NewFeedHandler(feedRepo *repository.FeedRepository, postRepo *repository.PostRepository) *FeedHandler {
	return &FeedHandler{feedRepo: feedRepo, postRepo: postRepo}
}



func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request){

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0

	limitstr := r.URL.Query().Get("limit")
	offsetstr := r.URL.Query().Get("offset")

	
	if limitstr != "" {
		l, err := strconv.Atoi(limitstr)
		if err == nil {
			limit = l
		}
	}

	if offsetstr != "" {
		o, err := strconv.Atoi(offsetstr)
		if err == nil {
			offset = o
		}
	}

	if limit > 50 {
		limit = 50
	}


	feed, err := h.feedRepo.GetFeed(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get feed", http.StatusInternalServerError)
		return
	}

	posts, err := h.postRepo.GetPostsByIDs(r.Context(), feed)
	if err != nil {
		http.Error(w, "Failed to get posts, reload ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
	})

}