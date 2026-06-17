package repository

import (
	"context"
	"fmt"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type DLQRepository interface {
	RecordDLQMessage(context.Context, *dtov1.DLQMessage) error
	GetDLQMessages(context.Context, *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error)
	SetDLQMessageRetrying(context.Context, int64) (*dtov1.RetryDLQMessageData, error)
	MarkAsResolved(ctx context.Context, id int64) (string, error)
	SetSafetyStatusUnknown(ctx context.Context, code string) error
}

type dlqRepository struct {
	db *pgxpool.Pool
}

func NewDLQRepository(db *pgxpool.Pool) DLQRepository {
	return &dlqRepository{db: db}
}

func (r *dlqRepository) RecordDLQMessage(c context.Context, req *dtov1.DLQMessage) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
INSERT INTO dlq_messages(source_stream, source_consumer, stream_seq, subject, payload, reason, failed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

	_, err := r.db.Exec(ctx, query, req.SourceStream, req.SourceConsumer, req.StreamSeq, req.Subject, req.Payload, req.Reason, req.FailedAt)
	return err
}

func (r *dlqRepository) GetDLQMessages(c context.Context, req *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 构造 where 语句
	whereSQL := ""
	clauses := make([]string, 0)
	args := make([]any, 0)

	if req.Status != nil {
		args = append(args, *req.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}

	if req.IsResolved != nil {
		if *req.IsResolved {
			clauses = append(clauses, "resolved_at IS NOT NULL")
		} else {
			clauses = append(clauses, "resolved_at IS NULL")
		}
	}

	if len(clauses) != 0 {
		whereSQL = " WHERE " + strings.Join(clauses, " AND ")
	}

	// 查询具体 DLQ Messages
	queryItem := `
SELECT id, source_stream, source_consumer, stream_seq, subject, payload, reason, status, failed_at, resolved_at
FROM dlq_messages
` + whereSQL + fmt.Sprintf(`
ORDER BY failed_at DESC, id DESC
OFFSET $%d LIMIT $%d`, len(args)+1, len(args)+2)

	rows, err := tx.Query(ctx, queryItem, append(args, (req.PageNum-1)*req.PageSize, req.PageSize)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.DLQMessageItem, error) {
		var item dtov1.DLQMessageItem
		err := row.Scan(
			&item.ID,
			&item.SourceStream,
			&item.SourceConsumer,
			&item.StreamSeq,
			&item.Subject,
			&item.Payload,
			&item.Reason,
			&item.Status,
			&item.FailedAt,
			&item.ResolvedAt,
		)
		return &item, err
	})
	if err != nil {
		return nil, err
	}

	// 查询总数
	queryTotal := `
SELECT count(*)
FROM dlq_messages
` + whereSQL

	var total int64
	err = tx.QueryRow(ctx, queryTotal, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	resp := &dtov1.GetDLQMessagesResp{
		Total:    total,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Items:    items,
	}

	return resp, nil
}

func (r *dlqRepository) SetDLQMessageRetrying(c context.Context, id int64) (*dtov1.RetryDLQMessageData, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
UPDATE dlq_messages
SET status = 'retrying'
WHERE id = $1
RETURNING subject, payload
`

	var data dtov1.RetryDLQMessageData
	err := r.db.QueryRow(ctx, query, id).Scan(
		&data.Subject,
		&data.Payload,
	)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *dlqRepository) MarkAsResolved(c context.Context, id int64) (string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
UPDATE dlq_messages
SET status = 'resolved', resolved_at = NOW()
WHERE id = $1
RETURNING subject
`

	var subject string
	err := r.db.QueryRow(ctx, query, id).Scan(&subject)
	if err != nil {
		return "", err
	}

	return subject, nil
}

func (r *dlqRepository) SetSafetyStatusUnknown(c context.Context, code string) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
UPDATE links
SET safety_status = 'unknown'
WHERE code = $1
`
	_, err := r.db.Exec(ctx, query, code)
	return err
}
