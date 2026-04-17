package feed

import (
	"context"
	"time"
)

type Worker struct {
	queue     *EventQueue
	fanoutSvc *FeedFanout
	dlq *DeadLetterQueue
}

func NewWorker(queue *EventQueue,dlq *DeadLetterQueue, fanout *FeedFanout) *Worker {
	return &Worker{
		queue:     queue,
		dlq: dlq,
		fanoutSvc: fanout,
	}
}

func (w *Worker) Start(ctx context.Context) {
	var err error
	for {
		select {
		case event := <-w.queue.Subscribe():
			for i := 0; i < 3; i++ {
				err = w.fanoutSvc.FanoutPost(ctx, event)
				if err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				w.dlq.Push(event)
			}

		case <-ctx.Done():
			return
		}
	}
}
