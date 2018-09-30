CREATE EXTENSION pgcrypto;

CREATE TABLE jobs(
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   error_uri TEXT,
   execute_at timestamp NOT NULL,
   payload BYTEA NOT NULL,
   uri TEXT NOT NULL,
   created_at timestamp DEFAULT now(),
   updated_at timestamp DEFAULT now()
);

CREATE TABLE done(
  id SERIAL PRIMARY KEY,
  job_id UUID NOT NULL,
  created_at timestamp DEFAULT now(),
  updated_at timestamp DEFAULT now()
)
