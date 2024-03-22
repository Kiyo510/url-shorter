CREATE TABLE IF NOT EXISTS short_url_mappings
(
    hash         varchar(255) PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE
);
