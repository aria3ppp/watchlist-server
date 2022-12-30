BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(40) UNIQUE NOT NULL,
    hashed_password VARCHAR(72) NOT NULL,

    first_name VARCHAR(20),
    last_name VARCHAR(20),

    bio VARCHAR(500),
    birthdate DATE,
    jointime TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    avatar varchar
);

COMMIT;