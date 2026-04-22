package feed

import (
	"context"

	"github.com/Dhushyanthc/event-feed-engine/internal/models"
	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
)

const followerFanoutBatchSize = 500

type FeedFanout struct {
	followRepo *repository.FollowRepository
	feedRepo   *repository.FeedRepository
}

func NewFeedFanout(followRepo *repository.FollowRepository, feedRepo *repository.FeedRepository) *FeedFanout {
	return &FeedFanout{followRepo: followRepo, feedRepo: feedRepo}
}

func (s *FeedFanout) FanoutPost(ctx context.Context, event *PostCreatedEvent) error {
	authorFeed := &models.FeedItem{
		UserID:    event.UserID,
		PostID:    event.PostID,
		CreatedAt: event.CreatedAt,
	}

	err := s.feedRepo.InsertFeedItem(ctx, authorFeed)
	if err != nil {
		return err
	}

	for offset := 0; ; offset += followerFanoutBatchSize {
		followers, err := s.followRepo.GetFollowers(ctx, event.UserID, followerFanoutBatchSize, offset)
		if err != nil {
			return err
		}

		if len(followers) == 0 {
			break
		}

		for _, followerID := range followers {
			feedItem := &models.FeedItem{
				UserID:    followerID,
				PostID:    event.PostID,
				CreatedAt: event.CreatedAt,
			}

			err := s.feedRepo.InsertFeedItem(ctx, feedItem)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
