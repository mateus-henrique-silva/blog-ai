CREATE TABLE IF NOT EXISTS posts (
    id           INTEGER  PRIMARY KEY AUTOINCREMENT,
    uuid         TEXT     NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(16)))),
    title        TEXT     NOT NULL,
    slug         TEXT     NOT NULL UNIQUE,
    excerpt      TEXT     NOT NULL DEFAULT '',
    content_md   TEXT     NOT NULL DEFAULT '',
    content_html TEXT     NOT NULL DEFAULT '',
    cover_image  TEXT     NOT NULL DEFAULT '',
    category     TEXT     NOT NULL DEFAULT '',
    tags         TEXT     NOT NULL DEFAULT '',
    status       TEXT     NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published')),
    published_at DATETIME,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_posts_slug      ON posts(slug);
CREATE INDEX IF NOT EXISTS idx_posts_status    ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_category  ON posts(category);
CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(published_at DESC);
