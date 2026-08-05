CREATE TABLE credentials (
    id TEXT PRIMARY KEY, 
    prospect_id TEXT NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_credentials_prospect_id ON credentials (prospect_id);
CREATE UNIQUE INDEX idx_credentials_username ON credentials (username);