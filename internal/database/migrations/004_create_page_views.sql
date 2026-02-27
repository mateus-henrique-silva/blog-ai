CREATE TABLE IF NOT EXISTS page_views (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    post_id    INTEGER  NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    ip_hash    TEXT     NOT NULL DEFAULT '',
    user_agent TEXT     NOT NULL DEFAULT '',
    viewed_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_pv_post_id ON page_views(post_id);
CREATE INDEX IF NOT EXISTS idx_pv_viewed  ON page_views(viewed_at DESC);
CREATE INDEX IF NOT EXISTS idx_pv_post_ip ON page_views(post_id, ip_hash);
