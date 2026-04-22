package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

type FeedEvent struct {
	ID        int64
	PostID    int64
	UserID    int64
	CreatedAt time.Time
	Attempts  int
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *EventRepository) CreateEvent(ctx context.Context, postID, userID int64) error {
	query := `
	INSERT INTO feed_events (post_id, user_id, created_at, processed)
	VALUES ($1, $2, NOW(), FALSE)
	`

	_, err := r.db.Exec(ctx, query, postID, userID)
	return err
}

func (r *EventRepository) CreateEventTx(ctx context.Context, tx pgx.Tx, postID, userID int64, createdAt time.Time) error {
	query := `
	INSERT INTO feed_events (post_id, user_id, created_at, processed)
	VALUES ($1, $2, $3, FALSE)
	`

	_, err := tx.Exec(ctx, query, postID, userID, createdAt)
	return err
}

func (r *EventRepository) GetUnprocessedEvents(ctx context.Context, limit int) ([]*FeedEvent, error) {
	query := `
	SELECT id, post_id, user_id, created_at, attempts
	FROM feed_events
	WHERE processed = FALSE
	AND failed = FALSE
	ORDER BY created_at ASC
	LIMIT $1
	FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*FeedEvent

	for rows.Next() {
		var e FeedEvent
		err := rows.Scan(&e.ID, &e.PostID, &e.UserID, &e.CreatedAt, &e.Attempts)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, nil
}

func (r *EventRepository) MarkProcessed(ctx context.Context, eventID int64) error {
	query := `
	UPDATE feed_events
	SET processed = TRUE
	WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, eventID)
	return err
}

func (r *EventRepository) GetUnprocessedEventsTx(ctx context.Context, tx pgx.Tx, limit int) ([]*FeedEvent, error) {
	query := `
	SELECT id, post_id, user_id, created_at, attempts
	FROM feed_events
	WHERE processed = FALSE
	AND failed = FALSE
  	AND (
    	processing = FALSE
    	OR updated_at < NOW() - INTERVAL '30 seconds'
  	)
	ORDER BY created_at ASC
	LIMIT $1
	FOR UPDATE SKIP LOCKED;
	`

	rows, err := tx.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*FeedEvent

	for rows.Next() {
		var e FeedEvent
		if err := rows.Scan(&e.ID, &e.PostID, &e.UserID, &e.CreatedAt, &e.Attempts); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, nil
}

func (r *EventRepository) MarkProcessingTx(ctx context.Context, tx pgx.Tx, eventID int64) error {
	query := `
	UPDATE feed_events
	SET processing = TRUE,
	    attempts = attempts + 1,
	    updated_at = NOW()
	WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, eventID)
	return err
}

func (r *EventRepository) MarkFailedTx(ctx context.Context, tx pgx.Tx, eventID int64, lastError string, maxAttempts int) error {
	query := `
	UPDATE feed_events
	SET processing = FALSE,
	    failed = attempts >= $2,
	    last_error = LEFT($3, 1000),
	    updated_at = NOW()
	WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, eventID, maxAttempts, lastError)
	return err
}

func (r *EventRepository) MarkProcessedTx(ctx context.Context, tx pgx.Tx, eventID int64) error {
	query := `
	UPDATE feed_events
	SET processed = TRUE,
	    processing = FALSE,
	    updated_at = NOW()
	WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, eventID)
	return err
}
