package feed

import (
	"context"
	"time"

	"github.com/Dhushyanthc/event-feed-engine/internal/repository"
	"go.uber.org/zap"
)

type DBWorker struct {
	eventRepo *repository.EventRepository
	fanoutSvc *FeedFanout
	logger    *zap.Logger
}

func NewDBWorker(eventRepo *repository.EventRepository, fanout *FeedFanout, logger *zap.Logger) *DBWorker {
	return &DBWorker{
		eventRepo: eventRepo,
		fanoutSvc: fanout,
		logger:    logger,
	}
}

func (w *DBWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.processBatch(ctx)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (w *DBWorker) processBatch(ctx context.Context) {

	tx, err := w.eventRepo.BeginTx(ctx)
	if err != nil {
		w.logger.Error("failed to begin tx", zap.Error(err))
		return
	}

	events, err := w.eventRepo.GetUnprocessedEventsTx(ctx, tx, 10)
	if err != nil {
		tx.Rollback(ctx)
		w.logger.Error("failed to fetch events", zap.Error(err))
		return
	}

	for _, event := range events {
		err := w.eventRepo.MarkProcessingTx(ctx, tx, event.ID)
		if err != nil {
			tx.Rollback(ctx)
			w.logger.Error("failed to mark processing", zap.Error(err))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		w.logger.Error("failed to commit claim tx", zap.Error(err))
		return
	}

	var successfulEvents []int64

	for _, event := range events {

		e := &PostCreatedEvent{
			PostID:    event.PostID,
			UserID:    event.UserID,
			CreatedAt: event.CreatedAt,
		}

		err := w.fanoutSvc.FanoutPost(ctx, e)
		if err != nil {
			w.logger.Error("fanout failed", zap.Error(err))
			continue
		}

		successfulEvents = append(successfulEvents, event.ID)
	}

	tx2, err := w.eventRepo.BeginTx(ctx)
	if err != nil {
		w.logger.Error("failed to begin finalize tx", zap.Error(err))
		return
	}

	for _, id := range successfulEvents {
		err := w.eventRepo.MarkProcessedTx(ctx, tx2, id)
		if err != nil {
			tx2.Rollback(ctx)
			w.logger.Error("failed to mark processed", zap.Error(err))
			return
		}
	}

	if err := tx2.Commit(ctx); err != nil {
		w.logger.Error("failed to commit finalize tx", zap.Error(err))
	}
}
