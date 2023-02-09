BEGIN;

-- create watchfilms table
CREATE TABLE IF NOT EXISTS watchfilms (
    id SERIAL PRIMARY KEY,
    
    user_id INT NOT NULL,
    film_id INT NOT NULL,

    time_added TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    time_watched TIMESTAMPTZ
);

-- add user_id foreign key constraint
ALTER TABLE IF EXISTS watchfilms
    ADD CONSTRAINT watchfilms_fk_users
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- create index on user_id fk
CREATE INDEX IF NOT EXISTS watchfilms_idx_user_id ON watchfilms (user_id);

-- add film_id foreign key constraint
ALTER TABLE IF EXISTS watchfilms
    ADD CONSTRAINT watchfilms_fk_films
    FOREIGN KEY (film_id)
    REFERENCES films(id)
    ON DELETE CASCADE;

-- create index on film_id fk
CREATE INDEX IF NOT EXISTS watchfilms_idx_film_id ON watchfilms (film_id);

-- create index on time_added and time_watched
CREATE INDEX IF NOT EXISTS watchfilms_idx_time_added ON watchfilms (time_added);
CREATE INDEX IF NOT EXISTS watchfilms_idx_time_watched ON watchfilms (time_watched);

COMMIT;