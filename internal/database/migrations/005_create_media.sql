CREATE TABLE IF NOT EXISTS media (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    filename    TEXT     NOT NULL,
    original    TEXT     NOT NULL,
    mime_type   TEXT     NOT NULL,
    size_bytes  INTEGER  NOT NULL DEFAULT 0,
    url         TEXT     NOT NULL,
    uploaded_by INTEGER  NOT NULL REFERENCES admin_users(id),
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
