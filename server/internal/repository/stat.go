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
WITH bounds AS (
	SELECT date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') AS today
),
days AS (
	SELECT
    	generate_series(
        	today - ($1::int - 1) * interval '1 day',
            today,
            interval '1 day'
        ) AS day
    FROM bounds
),
agg AS (
	SELECT
    	date_trunc('day', clicked_at AT TIME ZONE 'Asia/Shanghai') AS day,
    	count(*) AS count
    FROM click_log
    CROSS JOIN bounds
    WHERE
    	clicked_at >= (today - ($1::int - 1) * interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    	AND clicked_at < (today + interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	days.day,
	COALESCE(agg.count, 0) AS count
FROM days
LEFT JOIN agg
	ON agg.day = days.day
ORDER BY days.day;
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
WITH bounds AS (
	SELECT
		date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') AS today
),
days AS (
	SELECT
		generate_series(
			today - ($1::int - 1) * interval '1 day',
			today,
			interval '1 day'
		) AS day
	FROM bounds
),
agg AS (
	SELECT
		date_trunc('day', created_at AT TIME ZONE 'Asia/Shanghai') AS day,
		count(*) AS count
	FROM links
	CROSS JOIN bounds
	WHERE
		created_at >= (today - ($1::int - 1) * interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    	AND created_at < (today + interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	days.day,
	COALESCE(agg.count, 0) AS count
FROM days
LEFT JOIN agg
	ON agg.day = days.day
ORDER BY days.day;
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
WITH bounds AS (
	SELECT
		date_trunc('day', now() AT TIME ZONE 'Asia/Shanghai') AS today
),
days AS (
	SELECT
		generate_series(
			today - ($1::int - 1) * interval '1 day',
			today,
            interval '1 day'
		) AS day
    FROM bounds
),
agg AS (
    SELECT
    	date_trunc('day', created_at AT TIME ZONE 'Asia/Shanghai') AS day,
    	count(*) AS count
    FROM users
    CROSS JOIN bounds
    WHERE
    	created_at >= (today - ($1::int - 1) * interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    	AND created_at < (today + interval '1 day') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	days.day,
	COALESCE(agg.count, 0) AS count
FROM days
LEFT JOIN agg
	ON agg.day = days.day
ORDER BY days.day;
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
WITH bounds AS (
	SELECT date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') AS currentHour
),
hours AS (
	SELECT
    	generate_series(
        	currentHour - ($1::int - 1) * interval '1 hour',
            currentHour,
            interval '1 hour'
        ) AS hour
    FROM bounds
),
agg AS (
	SELECT
    	date_trunc('hour', clicked_at AT TIME ZONE 'Asia/Shanghai') AS hour,
    	count(*) AS count
    FROM click_log
    CROSS JOIN bounds
    WHERE
    	clicked_at >= (currentHour - ($1::int - 1) * interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    	AND clicked_at < (currentHour + interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	hours.hour,
	COALESCE(agg.count, 0) AS count
FROM hours
LEFT JOIN agg
	ON agg.hour = hours.hour
ORDER BY hours.hour;
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
WITH bounds AS (
	SELECT
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') AS currentHour
),
hours AS (
	SELECT
		generate_series(
			currentHour - ($1::int - 1) * interval '1 hour',
			currentHour,
			interval '1 hour'
		) AS hour
	FROM bounds
),
agg AS (
	SELECT
		date_trunc('hour', created_at AT TIME ZONE 'Asia/Shanghai') AS hour,
		count(*) AS count
	FROM links
	CROSS JOIN bounds
	WHERE
		created_at >= (currentHour - ($1::int - 1) * interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    	AND created_at < (currentHour + interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	hours.hour,
	COALESCE(agg.count, 0) AS count
FROM hours
LEFT JOIN agg
	ON agg.hour = hours.hour
ORDER BY hours.hour;
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
WITH bounds AS (
	SELECT
		date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') AS currentHour
),
hours AS (
	SELECT
		generate_series(
			currentHour - ($1::int - 1) * interval '1 hour',
			currentHour,
            interval '1 hour'
		) AS hour
    FROM bounds
),
agg AS (
    SELECT
    	date_trunc('hour', created_at AT TIME ZONE 'Asia/Shanghai') AS hour,
    	count(*) AS count
    FROM users
    CROSS JOIN bounds
    WHERE
    	created_at >= (currentHour - ($1::int - 1) * interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    	AND created_at < (currentHour + interval '1 hour') AT TIME ZONE 'Asia/Shanghai'
    GROUP BY 1
)
SELECT
	hours.hour,
	COALESCE(agg.count, 0) AS count
FROM hours
LEFT JOIN agg
	ON agg.hour = hours.hour
ORDER BY hours.hour;
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
