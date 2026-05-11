package repository

import (
	"context"
	"fmt"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type LinkRepository interface {
	CreateLink(context.Context, *dtov1.CreateLinkReq) (int64, error)
	GetLinks(context.Context, *dtov1.GetLinksReq) ([]*dtov1.LinkItem, error)
	UpdateLink(context.Context, *dtov1.UpdateLinkReq) error
	GetLinkByID(context.Context, int64, int64) (*model.Link, error)
	DeleteLink(context.Context, *dtov1.DeleteLinkReq) error
	GetOriginalURLByCode(context.Context, string) (string, error)
	CheckCodeConflict(context.Context, string) (bool, error)
	RecordClick(context.Context, *dtov1.ClickLinkReq) error
	GetViewCountByUserID(context.Context, int64) (int64, error)
	GetLinkCountByUserID(context.Context, int64) (int64, error)
}
type linkRepository struct {
	db *pgxpool.Pool
}

func NewLinkRepository(db *pgxpool.Pool) LinkRepository {
	return &linkRepository{db: db}
}

func (r *linkRepository) CreateLink(c context.Context, req *dtov1.CreateLinkReq) (int64, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO links(user_id, original_url, code, expires_at, created_at, updated_at, view_count, status)
		VALUES ($1, $2, $3, $4, now(), now(), 0, 'active')
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, req.UserID, req.OriginalURL, req.Code, req.ExpiresAt).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (r *linkRepository) GetLinks(c context.Context, req *dtov1.GetLinksReq) ([]*dtov1.LinkItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	whereClauses := []string{"user_id = $1"}
	args := []any{req.UserID}
	argIndex := 2

	if req.OriginalURL != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("original_url LIKE $%d", argIndex))
		searchArgs := fmt.Sprintf("%%%s%%", *req.OriginalURL)
		args = append(args, searchArgs)
		argIndex++
	}
	if req.Code != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("code LIKE $%d", argIndex))
		searchArgs := fmt.Sprintf("%%%s%%", *req.Code)
		args = append(args, searchArgs)
		argIndex++
	}
	if req.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	if req.ExpiresBegin != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("expires_at >= $%d", argIndex))
		args = append(args, *req.ExpiresBegin)
		argIndex++
	}
	if req.ExpiresEnd != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("expires_at <= $%d", argIndex))
		args = append(args, *req.ExpiresEnd)
		argIndex++
	}

	query := fmt.Sprintf(`
SELECT id, original_url, code, view_count, status, created_at, updated_at, expires_at
FROM links
WHERE %s
AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $%d OFFSET $%d
	`,
		strings.Join(whereClauses, " AND "),
		argIndex,
		argIndex+1)

	args = append(args, req.PageSize, (req.PageNum-1)*req.PageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//links := make([]*dtov1.LinkItem, 0)
	//
	//for rows.Next() {
	//	item := &dtov1.LinkItem{}
	//
	//	err := rows.Scan(
	//		&item.OriginalURL,
	//		&item.Code,
	//		&item.ViewCount,
	//		&item.Status,
	//		&item.CreatedAt,
	//		&item.UpdatedAt,
	//		&item.ExpiresAt,
	//	)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	links = append(links, item)
	//}
	//
	//if err := rows.Err(); err != nil {
	//	return nil, err
	//}

	links, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.LinkItem, error) {
		item := &dtov1.LinkItem{}

		err := row.Scan(
			&item.ID,
			&item.OriginalURL,
			&item.Code,
			&item.ViewCount,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		return item, nil
	})
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (r *linkRepository) UpdateLink(c context.Context, req *dtov1.UpdateLinkReq) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()
	// UPDATE links SET xxx WHERE ID = xx AND user_id = xx
	setClauses := []string{"expires_at = $1", "updated_at = NOW()"}
	args := []any{req.ExpiresAt}
	argIndex := 2
	if req.OriginalURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("original_url = $%d", argIndex))
		args = append(args, *req.OriginalURL)
		argIndex++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	query := fmt.Sprintf(
		"UPDATE links SET %s WHERE id = $%d AND user_id = $%d",
		strings.Join(setClauses, ", "),
		argIndex,
		argIndex+1,
	)
	args = append(args, req.ID, req.UserID)

	rows, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if rows.RowsAffected() == 0 {
		return consts.ErrNoRowsAffected
	}

	return nil
}

func (r *linkRepository) GetLinkByID(c context.Context, id int64, userID int64) (*model.Link, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
SELECT id, user_id, code, original_url, view_count, status, created_at, updated_at, expires_at
FROM links
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	resp := &model.Link{}

	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&resp.ID,
		&resp.UserID,
		&resp.Code,
		&resp.OriginalURL,
		&resp.ViewCount,
		&resp.Status,
		&resp.CreatedAt,
		&resp.UpdatedAt,
		&resp.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *linkRepository) DeleteLink(c context.Context, req *dtov1.DeleteLinkReq) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
UPDATE links SET deleted_at = NOW() WHERE id = $1 AND user_id = $2`

	rows, err := r.db.Exec(ctx, query, req.ID, req.UserID)
	if err != nil {
		return err
	}
	if rows.RowsAffected() == 0 {
		return consts.ErrNoRowsAffected
	}

	return nil
}

func (r *linkRepository) GetOriginalURLByCode(c context.Context, code string) (string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
SELECT original_url FROM links WHERE code = $1 AND deleted_at IS NULL AND status = 'active' AND (expires_at IS NULL OR expires_at > NOW())`

	var originalURL string
	err := r.db.QueryRow(ctx, query, code).Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (r *linkRepository) CheckCodeConflict(c context.Context, code string) (bool, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
SELECT EXISTS (
	SELECT 1
	FROM links
	WHERE code = $1 AND deleted_at IS NULL
)
`
	var exists bool
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *linkRepository) RecordClick(c context.Context, req *dtov1.ClickLinkReq) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateQuery := `
UPDATE links
SET view_count = view_count + 1, updated_at = NOW()
WHERE code = $1
	AND deleted_at IS NULL
	AND status = 'active'
`
	_, err = tx.Exec(ctx, updateQuery, req.Code)
	if err != nil {
		return err
	}

	logQuery := `
INSERT INTO click_log(code, ip, user_agent, referer, clicked_at)
VALUES ($1, $2, $3, $4, $5)
`
	_, err = tx.Exec(ctx, logQuery, req.Code, req.IP, req.UserAgent, req.Referer, req.ClickedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *linkRepository) GetViewCountByUserID(c context.Context, userID int64) (int64, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
SELECT COALESCE(SUM(view_count), 0)
FROM links
WHERE user_id = $1 AND deleted_at IS NULL
`

	var total int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *linkRepository) GetLinkCountByUserID(c context.Context, userID int64) (int64, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
SELECT COALESCE(COUNT(*), 0)
FROM links
WHERE user_id = $1 AND deleted_at IS NULL
`

	var total int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
