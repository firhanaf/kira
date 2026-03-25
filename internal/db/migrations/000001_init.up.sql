-- ─────────────────────────────────────────────────────────────────────────────
-- 001_init.up.sql
-- Initial schema: users, projects, features, time_entries,
--                 invoices, invoice_line_items, notifications, refresh_tokens
-- ─────────────────────────────────────────────────────────────────────────────

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─────────────────────────────────────────
-- UTILITY: auto updated_at trigger function
-- ─────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ─────────────────────────────────────────
-- USERS
-- ─────────────────────────────────────────
CREATE TABLE users (
                       id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                       email            TEXT        UNIQUE NOT NULL,
                       name             TEXT        NOT NULL,
                       avatar_url       TEXT,
                       password_hash    TEXT        NOT NULL,
                       default_currency TEXT        NOT NULL DEFAULT 'IDR',
                       default_rate     INTEGER     NOT NULL DEFAULT 0,      -- cents/sen per hour
                       timezone         TEXT        NOT NULL DEFAULT 'Asia/Jakarta',
                       created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       deleted_at       TIMESTAMPTZ
);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────
-- PROJECTS
-- ─────────────────────────────────────────
CREATE TABLE projects (
                          id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                          user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          name         TEXT        NOT NULL,
                          description  TEXT,
                          client_name  TEXT,
                          client_email TEXT,
                          status       TEXT        NOT NULL DEFAULT 'active',
                          CONSTRAINT projects_status_check
                              CHECK (status IN ('active', 'completed', 'on_hold', 'archived')),
                          currency     TEXT        NOT NULL DEFAULT 'IDR',
                          CONSTRAINT projects_currency_check
                              CHECK (currency IN ('IDR', 'USD')),
                          hourly_rate  INTEGER     NOT NULL DEFAULT 0,
                          color        TEXT        NOT NULL DEFAULT '#6366f1',
                          created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_projects_user_id ON projects(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_status  ON projects(status)  WHERE deleted_at IS NULL;

CREATE TRIGGER trg_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────
-- FEATURES
-- ─────────────────────────────────────────
CREATE TABLE features (
                          id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                          project_id       UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                          name             TEXT        NOT NULL,
                          description      TEXT,
                          notes            TEXT,
                          severity         TEXT        NOT NULL DEFAULT 'medium',
                          CONSTRAINT features_severity_check
                              CHECK (severity IN ('critical', 'high', 'medium', 'low')),
                          status           TEXT        NOT NULL DEFAULT 'idle',
                          CONSTRAINT features_status_check
                              CHECK (status IN ('idle', 'in_progress', 'paused', 'done')),
                          git_branch       TEXT,
                          git_repo_url     TEXT,
                          elapsed_seconds  INTEGER     NOT NULL DEFAULT 0,
                          position         INTEGER     NOT NULL DEFAULT 0,
                          created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_features_project_id ON features(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_features_status     ON features(status)     WHERE deleted_at IS NULL;

CREATE TRIGGER trg_features_updated_at
    BEFORE UPDATE ON features
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────
-- TIME ENTRIES
-- One row per start→pause cycle
-- ─────────────────────────────────────────
CREATE TABLE time_entries (
                              id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                              feature_id       UUID        NOT NULL REFERENCES features(id) ON DELETE CASCADE,
                              user_id          UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
                              started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              ended_at         TIMESTAMPTZ,                   -- NULL = currently running
                              duration_seconds INTEGER,                       -- set when ended_at is set
                              note             TEXT,
                              created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- DB-level enforcement: only one active (running) entry per user
CREATE UNIQUE INDEX idx_time_entries_one_active_per_user
    ON time_entries(user_id)
    WHERE ended_at IS NULL;

CREATE INDEX idx_time_entries_feature_id  ON time_entries(feature_id);
CREATE INDEX idx_time_entries_user_id     ON time_entries(user_id);
CREATE INDEX idx_time_entries_started_at  ON time_entries(started_at DESC);

-- ─────────────────────────────────────────
-- INVOICES
-- ─────────────────────────────────────────
CREATE TABLE invoices (
                          id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
                          project_id       UUID           NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                          user_id          UUID           NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
                          invoice_number   TEXT           NOT NULL,
                          status           TEXT           NOT NULL DEFAULT 'draft',
                          CONSTRAINT invoices_status_check
                              CHECK (status IN ('draft', 'sent', 'paid', 'overdue')),
                          subtotal         INTEGER        NOT NULL DEFAULT 0,   -- cents/sen
                          margin_pct       NUMERIC(5,2)   NOT NULL DEFAULT 0,
                          tax_pct          NUMERIC(5,2)   NOT NULL DEFAULT 11,
                          total            INTEGER        NOT NULL DEFAULT 0,
                          currency         TEXT           NOT NULL DEFAULT 'IDR',
                          due_date         DATE,
                          paid_at          TIMESTAMPTZ,
                          share_token      TEXT           UNIQUE,
                          share_expires_at TIMESTAMPTZ,
                          pdf_url          TEXT,
                          notes            TEXT,
                          created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
                          updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_project_id   ON invoices(project_id);
CREATE INDEX idx_invoices_user_id      ON invoices(user_id);
CREATE INDEX idx_invoices_share_token  ON invoices(share_token) WHERE share_token IS NOT NULL;

CREATE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────
-- INVOICE LINE ITEMS
-- Immutable snapshot at invoice creation time
-- ─────────────────────────────────────────
CREATE TABLE invoice_line_items (
                                    id               UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
                                    invoice_id       UUID    NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
                                    feature_id       UUID    REFERENCES features(id) ON DELETE SET NULL,
                                    feature_name     TEXT    NOT NULL,   -- snapshot
                                    severity         TEXT    NOT NULL,   -- snapshot
                                    git_branch       TEXT,               -- snapshot
                                    elapsed_seconds  INTEGER NOT NULL,
                                    hourly_rate      INTEGER NOT NULL,
                                    amount           INTEGER NOT NULL,   -- cents/sen
                                    notes            TEXT
);

CREATE INDEX idx_line_items_invoice_id ON invoice_line_items(invoice_id);

-- ─────────────────────────────────────────
-- NOTIFICATIONS
-- ─────────────────────────────────────────
CREATE TABLE notifications (
                               id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                               project_id      UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                               user_id         UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
                               feature_id      UUID        REFERENCES features(id) ON DELETE SET NULL,
                               channel         TEXT        NOT NULL,
                               CONSTRAINT notifications_channel_check
                                   CHECK (channel IN ('email', 'shareable_link')),
                               recipient_email TEXT,
                               subject         TEXT,
                               message         TEXT        NOT NULL,
                               status          TEXT        NOT NULL DEFAULT 'pending',
                               CONSTRAINT notifications_status_check
                                   CHECK (status IN ('pending', 'sent', 'failed')),
                               share_token     TEXT        UNIQUE,
                               share_expires_at TIMESTAMPTZ,
                               sent_at         TIMESTAMPTZ,
                               error_message   TEXT,
                               created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_project_id ON notifications(project_id);
CREATE INDEX idx_notifications_share_token ON notifications(share_token)
    WHERE share_token IS NOT NULL;

-- ─────────────────────────────────────────
-- REFRESH TOKENS
-- ─────────────────────────────────────────
CREATE TABLE refresh_tokens (
                                id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                token_hash  TEXT        UNIQUE NOT NULL,
                                expires_at  TIMESTAMPTZ NOT NULL,
                                revoked_at  TIMESTAMPTZ,
                                created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
