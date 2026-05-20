CREATE TABLE IF NOT EXISTS click_log (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL,
    IP inet NOT NULL,
    user_agent TEXT NOT NULL,
    referer TEXT,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_click_log_code ON click_log(code);
CREATE UNIQUE INDEX idx_click_log_event_id ON click_log(event_id);

WITH hours AS (
    SELECT generate_series(
                   date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai') - (($1::int - 1) * interval '1 hour'),
                   date_trunc('hour', now() AT TIME ZONE 'Asia/Shanghai'),
                   interval '1 hour'
           ) AS hour
)
SELECT
    hours.hour,
SELECT(*)
FROM hours
         LEFT JOIN links
                   ON links.created_at >= hours.hour
                       AND links.created_at < hours.hour + interval '1 hour'
GROUP BY hours.hour
ORDER BY hours.hour