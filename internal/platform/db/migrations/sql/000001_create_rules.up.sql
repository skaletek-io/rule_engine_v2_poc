CREATE TABLE templates (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    description TEXT,
    schema      JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id  UUID        NOT NULL REFERENCES templates(id) ON DELETE RESTRICT,
    external_ref TEXT,
    payload      JSONB       NOT NULL DEFAULT '{}',
    occurred_at  TIMESTAMPTZ NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rules (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID        NOT NULL REFERENCES templates(id) ON DELETE RESTRICT,
    name        TEXT        NOT NULL,
    expression  TEXT        NOT NULL,
    severity    TEXT        NOT NULL DEFAULT 'medium',
    message     TEXT        NOT NULL DEFAULT '',
    priority    INTEGER     NOT NULL DEFAULT 10,
    status      TEXT        NOT NULL DEFAULT 'draft',
    mode        TEXT        NOT NULL DEFAULT 'live',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
