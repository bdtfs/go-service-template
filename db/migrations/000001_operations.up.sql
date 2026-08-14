CREATE TABLE IF NOT EXISTS operations (
    external_id TEXT PRIMARY KEY,
    type        TEXT        NOT NULL,
    status      TEXT        NOT NULL,
    user_id     BIGINT      NOT NULL,
    amount      BIGINT      NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS operations_user_id_created_at_idx
    ON operations (user_id, created_at DESC);
