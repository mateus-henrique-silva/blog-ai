CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT     PRIMARY KEY,
    user_id    INTEGER  NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    data       TEXT     NOT NULL DEFAULT '{}',
    ip_hash    TEXT     NOT NULL DEFAULT '',
    user_agent TEXT     NOT NULL DEFAULT '',
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_user    ON sessions(user_id);
