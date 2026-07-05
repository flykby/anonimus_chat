package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Input struct {
	UserID   *int64
	DialogID *int64
	Type     Type
	Metadata any
}

type Emitter struct {
	log *slog.Logger
}

func NewEmitter(log *slog.Logger) *Emitter {
	return &Emitter{log: log}
}

func (e *Emitter) Emit(ctx context.Context, db DBTX, in Input) error {
	if err := validateType(in.Type); err != nil {
		return err
	}

	meta, err := marshalMetadata(in.Type, in.Metadata)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO events (user_id, dialog_id, event_type, metadata)
		VALUES ($1, $2, $3, $4::jsonb)
	`, in.UserID, in.DialogID, in.Type.String(), string(meta))
	if err != nil {
		return fmt.Errorf("insert event %s: %w", in.Type, err)
	}

	if e.log != nil {
		e.log.Info("event emitted",
			"event_type", in.Type.String(),
			"user_id", in.UserID,
			"dialog_id", in.DialogID,
			"metadata", string(meta),
		)
	}
	return nil
}
