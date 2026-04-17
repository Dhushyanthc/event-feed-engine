package feed

import (
	"context"
	"go.uber.org/zap"
)

type DLQWorker struct {
	dlq    *DeadLetterQueue
	logger *zap.Logger
	fanoutSvc *FeedFanout
}

func NewDLQWorker(dlq *DeadLetterQueue, logger *zap.Logger, fanout *FeedFanout) *DLQWorker {
	return &DLQWorker{
		dlq:    dlq,
		logger: logger,
		fanoutSvc: fanout,

	}
}

func (w *DLQWorker) Start(ctx context.Context) {
	for {
		select {
		case event := <-w.dlq.Subscribe():

			var err error

			for i := 0; i < 3; i++ {
				err = w.fanoutSvc.FanoutPost(ctx, event)
				if err == nil {
					break
				}
			}

			if err != nil {
				w.logger.Error(
					"DLQ retry failed permanently",
					zap.Int64("post_id", event.PostID),
					zap.Int64("user_id", event.UserID),
				)
			}

		case <-ctx.Done():
			return
		}
	}
}