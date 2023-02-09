BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(40) NOT NULL, -- unique
    password_hash VARCHAR(72) NOT NULL, -- bcrypt hash

    first_name VARCHAR(20),
    last_name VARCHAR(20),

    bio VARCHAR(500),
    birthdate DATE,
    jointime TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    avatar varchar
);

-- create unique index on email
CREATE UNIQUE INDEX IF NOT EXISTS users_unique_idx_email ON users (email);

COMMIT;