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
    REFERENCES users(id);

-- add film_id foreign key constraint
ALTER TABLE IF EXISTS watchfilms
    ADD CONSTRAINT watchfilms_fk_films
    FOREIGN KEY (film_id)
    REFERENCES films(id);

-- create index on time_added and time_watched
CREATE INDEX watchfilms_idx_time_added ON watchfilms (time_added);
CREATE INDEX watchfilms_idx_time_watched ON watchfilms (time_watched);

COMMIT;