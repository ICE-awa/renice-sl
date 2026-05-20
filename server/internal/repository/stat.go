package repository

import (
	"context"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type StatRepository interface {
	GetClickDayStat(context.Context, int) ([]*dtov1.ClickStatItem, error)
	GetLinkDayStat(context.Context, int) ([]*dtov1.LinkStatItem, error)
	GetUserDayStat(context.Context, int) ([]*dtov1.UserStatItem, error)
	GetClickHourStat(context.Context, int) ([]*dtov1.ClickStatItem, error)
	GetLinkHourStat(context.Context, int) ([]*dtov1.LinkStatItem, error)
	GetUserHourStat(context.Context, int) ([]*dtov1.UserStatItem, error)
}

type statRepository struct {
	db *pgxpool.Pool
}

func NewStatRepository(db *pgxpool.Pool) StatRepository {
	return &statRepository{db: db}
}

func (r *statRepository) GetClickDayStat(c context.Context, day int) ([]*dtov1.ClickStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH days AS (
	SELECT generate_series(
		date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 day'),
		date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai'),
		interval '1 day'
	) AS day
)
SELECT
    days.day,
    count(click_log.id)
FROM days
LEFT JOIN click_log
    ON click_log.clicked_at >= days.day
	AND click_log.clicked_at < days.day + interval '1 day'
GROUP BY days.day
ORDER BY days.day
`

	rows, err := r.db.Query(ctx, query, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.ClickStatItem, error) {
		var item dtov1.ClickStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *statRepository) GetLinkDayStat(c context.Context, day int) ([]*dtov1.LinkStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH days AS (
	SELECT generate_series(
		date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 day'),
        date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai'),
		interval '1 day'
	) AS day
)
SELECT
    days.day,
    count(links.id)
FROM days
LEFT JOIN links
    ON links.created_at >= days.day
    AND links.created_at < days.day + interval '1 day'
GROUP BY days.day
ORDER BY days.day
`

	rows, err := r.db.Query(ctx, query, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.LinkStatItem, error) {
		var item dtov1.LinkStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *statRepository) GetUserDayStat(c context.Context, day int) ([]*dtov1.UserStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH days AS (
    SELECT generate_series(
        date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 day'),
        date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai'),
        interval '1 day'
    ) AS day
)
SELECT
	days.day,
	count(users.id)
FROM days
LEFT JOIN users
    ON users.created_at >= days.day
    AND users.created_at < days.day + interval '1 day'
GROUP BY days.day
ORDER BY days.day
`

	rows, err := r.db.Query(ctx, query, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.UserStatItem, error) {
		var item dtov1.UserStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *statRepository) GetClickHourStat(c context.Context, hour int) ([]*dtov1.ClickStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH hours AS (
	SELECT generate_series(
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 hour'),
        date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai'),
        interval '1 hour'
	) AS hour
)
SELECT
	hours.hour,
	count(click_log.id)
FROM hours
LEFT JOIN click_log
ON click_log.clicked_at >= hours.hour
AND click_log.clicked_at < hours.hour + interval '1 hour'
GROUP BY hours.hour
ORDER BY hours.hour
`

	rows, err := r.db.Query(ctx, query, hour)
	if err != nil {
		return nil, err
	}

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.ClickStatItem, error) {
		var item dtov1.ClickStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *statRepository) GetLinkHourStat(c context.Context, hour int) ([]*dtov1.LinkStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH hours AS (
	SELECT generate_series(
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 hour'),
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai'),
		interval '1 hour'
	) AS hour
)
SELECT
	hours.hour,
	count(links.id)
FROM hours
LEFT JOIN links
ON links.created_at >= hours.hour
AND links.created_at < hours.hour + interval '1 hour'
GROUP BY hours.hour
ORDER BY hours.hour
`

	rows, err := r.db.Query(ctx, query, hour)
	if err != nil {
		return nil, err
	}

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.LinkStatItem, error) {
		var item dtov1.LinkStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *statRepository) GetUserHourStat(c context.Context, hour int) ([]*dtov1.UserStatItem, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
WITH hours AS (
	SELECT generate_series(
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 hour'),
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai'),
		interval '1 hour'
	) AS hour
)
SELECT
	hours.hour,
	count(users.id)
FROM hours
LEFT JOIN users
ON users.created_at >= hours.hour
AND users.created_at < hours.hour + interval '1 hour'
GROUP BY hours.hour
ORDER BY hours.hour
`

	rows, err := r.db.Query(ctx, query, hour)
	if err != nil {
		return nil, err
	}

	data, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*dtov1.UserStatItem, error) {
		var item dtov1.UserStatItem

		err := row.Scan(&item.Time, &item.Count)
		if err != nil {
			return nil, err
		}

		return &item, nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}
