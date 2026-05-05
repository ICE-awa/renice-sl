CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    view_count BIGINT NOT NULL,
    -- status 可以为 'active', 'inactive'
    status VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_links_code ON links(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_user_id ON links(user_id) WHERE deleted_at IS NULL;