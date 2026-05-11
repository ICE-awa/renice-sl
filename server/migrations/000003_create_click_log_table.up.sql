CREATE TABLE IF NOT EXISTS click_log (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    IP inet NOT NULL,
    user_agent TEXT NOT NULL,
    referer TEXT,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_click_log_code ON click_log(code);