CREATE TABLE prospects (
    id TEXT PRIMARY KEY, 
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended')),
    verification_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), 
    expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_prospect_username ON prospects (username); 
CREATE UNIQUE INDEX idx_prospect_email ON prospects (email); 