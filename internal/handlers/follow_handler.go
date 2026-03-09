package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Dhushyanthc/event-feed-engine/internal/middleware"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
)

type FollowHandler struct {
	followRepo *repository.FollowRepository
}
func NewFollowHandler(followRepo *repository.FollowRepository) *FollowHandler{
	return &FollowHandler{followRepo: followRepo}
}

type FollowRequest struct {
	FollowingID int64 `json:"following_id"`
}

type FollowResponse struct {		
	FollowerID int64 `json:"follower_id"`
	FollowingID int64 `json:"following_id"`
	CreatedAt string `json:"created_at"`
}

func (h *FollowHandler) FollowUser(w http.ResponseWriter, r *http.Request){

	var req FollowRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FollowingID <= 0 {
	http.Error(w, "invalid following_id", http.StatusBadRequest)
	return
	}

	followerID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// self follow check
	if followerID == req.FollowingID {
	http.Error(w, "cannot follow yourself", http.StatusBadRequest)
	return
	}

	follow := &models.Follow{
		FollowerID: followerID,
		FollowingID: req.FollowingID,
	}

	err = h.followRepo.CreateFollow(r.Context(), follow)
	if err != nil {
		http.Error(w, "failed to follow user", http.StatusInternalServerError)
		return
	}

	resp := FollowResponse{
		FollowerID: follow.FollowerID,
		FollowingID: follow.FollowingID,
		CreatedAt: follow.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}


func (h *FollowHandler) UnfollowUser(w http.ResponseWriter, r *http.Request){
	var req FollowRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	followerID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	follow := &models.Follow{
		FollowerID: followerID,
		FollowingID: req.FollowingID,
	}


	if req.FollowingID <= 0 {
	http.Error(w, "invalid following_id", http.StatusBadRequest)
	return
}

	err = h.followRepo.DeleteFollow(r.Context(), follow)
	if err != nil {
		http.Error(w, "failed to unfollow	user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//////////////////////////////////////////////////////
func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request){
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}


	limit := 20
	offset := 0

	limitstr := r.URL.Query().Get("limit")
	offsetstr := r.URL.Query().Get("offset")

	if limit > 100 {
	limit = 100
}

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

	followers, err := h.followRepo.GetFollowers(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get followers", http.StatusInternalServerError)
		return
	}				

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
	"followers": followers,
})

}


func (h *FollowHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	limit := 20
	offset := 0

	limitstr := r.URL.Query().Get("limit")
	offsetstr := r.URL.Query().Get("offset")

	if limit > 100 {
	limit = 100
}

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

	following, err := h.followRepo.GetFollowing(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get following", http.StatusInternalServerError)
		return 
	}	

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(following)
}