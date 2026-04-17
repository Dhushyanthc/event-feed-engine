package feed

import (
	"context"

	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"github.com/Dhushyanthc/event-feed-engine/internal/models"
)

type FeedFanout struct {
	followRepo *repository.FollowRepository
	feedRepo *repository.FeedRepository
}
func NewFeedFanout(followRepo *repository.FollowRepository, feedRepo *repository.FeedRepository) *FeedFanout {
	return &FeedFanout{followRepo: followRepo, feedRepo: feedRepo}
}

func (s *FeedFanout) FanoutPost(ctx context.Context, event *PostCreatedEvent) error{



	// Get followers of the user
	followers, err := s.followRepo.GetFollowers(ctx, event.UserID, 50, 0)
	if err != nil {
		return err
	}

	authorFeed := &models.FeedItem{
	UserID:    event.UserID,
	PostID:    event.PostID,
	CreatedAt: event.CreatedAt,
}

	err = s.feedRepo.InsertFeedItem(ctx, authorFeed)
	if err != nil {
		return err
	}

	// Create a new feed item for each follower
	for  _, followerID := range followers {
			feedItem := &models.FeedItem	{
			UserID:    followerID,
			PostID:    event.PostID,
			CreatedAt: event.CreatedAt,
	}
		err := s.feedRepo.InsertFeedItem(ctx, feedItem)
		if err != nil {
			return err
		}
}
	return nil	
}