CREATE EXTENSION pgcrypto;

CREATE TABLE jobs(
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   errors json NOT NULL DEFAULT '[]'::jsonb,
   error_uri TEXT,
   execute_at timestamp NOT NULL,
   payload json NOT NULL DEFAULT '{}'::jsonb,
   sent BOOLEAN NOT NULL DEFAULT false,
   try INTEGER NOT NULL DEFAULT 0,
   uri TEXT NOT NULL,
   created_at timestamp DEFAULT now(),
   updated_at timestamp DEFAULT now()
);
