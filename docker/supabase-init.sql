-- ultrawhale Supabase schema
-- Run: psql -U postgres -d ultrawhale -f supabase-init.sql

-- Create authenticator role for PostgREST
DO $$ BEGIN
  CREATE ROLE authenticator WITH LOGIN PASSWORD 'ultrawhale' NOINHERIT;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE anon WITH NOLOGIN;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

GRANT anon TO authenticator;

-- Workflows table
CREATE TABLE IF NOT EXISTS workflows (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT,
  script      TEXT NOT NULL,
  version     TEXT DEFAULT 'v1',
  status      TEXT DEFAULT 'draft',  -- draft, active, completed, failed
  created_at  TIMESTAMPTZ DEFAULT now(),
  updated_at  TIMESTAMPTZ DEFAULT now(),
  run_count   INTEGER DEFAULT 0,
  last_run    TIMESTAMPTZ
);

-- Sessions table (ultrawhale session tracking)
CREATE TABLE IF NOT EXISTS sessions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  TEXT NOT NULL UNIQUE,
  agent       TEXT DEFAULT 'ultrawhale',
  version     TEXT,
  machine     TEXT,
  arch        TEXT,
  tier        TEXT,
  started_at  TIMESTAMPTZ DEFAULT now(),
  ended_at    TIMESTAMPTZ,
  status      TEXT DEFAULT 'active'
);

-- Grant API access
GRANT USAGE ON SCHEMA public TO anon;
GRANT SELECT ON workflows TO anon;
GRANT INSERT ON workflows TO anon;
GRANT UPDATE ON workflows TO anon;
GRANT SELECT ON sessions TO anon;
GRANT INSERT ON sessions TO anon;
GRANT UPDATE ON sessions TO anon;

-- Notify PostgREST to reload schema
NOTIFY pgrst, 'reload schema';
