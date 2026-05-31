CREATE TABLE IF NOT EXISTS dlq_messages (
    id BIGSERIAL PRIMARY KEY,
    source_stream TEXT NOT NULL,
    source_consumer TEXT NOT NULL,
    stream_seq BIGINT NOT NULL,
    subject TEXT NOT NULL,
    payload JSONB NOT NULL,
    reason TEXT NOT NULL,
    -- status 可以为 ‘pending', 'resolved', 'failed', 'retrying' failed 表示 retry 过了但是失败了
    status TEXT NOT NULL DEFAULT 'pending',
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);