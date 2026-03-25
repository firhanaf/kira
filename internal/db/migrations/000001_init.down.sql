-- ─────────────────────────────────────────────────────────────────────────────
-- 001_init.down.sql
-- ─────────────────────────────────────────────────────────────────────────────

DROP TABLE IF EXISTS refresh_tokens       CASCADE;
DROP TABLE IF EXISTS notifications        CASCADE;
DROP TABLE IF EXISTS invoice_line_items   CASCADE;
DROP TABLE IF EXISTS invoices             CASCADE;
DROP TABLE IF EXISTS time_entries         CASCADE;
DROP TABLE IF EXISTS features             CASCADE;
DROP TABLE IF EXISTS projects             CASCADE;
DROP TABLE IF EXISTS users                CASCADE;

DROP FUNCTION IF EXISTS set_updated_at CASCADE;

DROP EXTENSION IF EXISTS "pgcrypto";
