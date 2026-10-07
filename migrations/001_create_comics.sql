CREATE TABLE IF NOT EXISTS comics (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(150) NOT NULL,
    type        VARCHAR(20)  NOT NULL, 
    author      VARCHAR(100) NOT NULL,
    stock       INTEGER      NOT NULL DEFAULT 0,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS comics_title_lower_key
    ON comics (LOWER(title));

CREATE INDEX IF NOT EXISTS comics_type_idx
    ON comics (type);