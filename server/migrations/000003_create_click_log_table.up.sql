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
CREATE INDEX idx_click_log_clicked_at ON click_log(clicked_at);