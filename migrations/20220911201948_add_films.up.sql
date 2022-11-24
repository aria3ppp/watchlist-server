BEGIN;

-- create films table
CREATE TABLE IF NOT EXISTS films (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    descriptions VARCHAR(500),

    date_released DATE NOT NULL,
	-- date_released TIMESTAMPTZ NOT NULL,
    duration INT,

    series_id INT,
    season_number INT,
    episode_number INT

    -- contributed_by INT NOT NULL,
    -- contributed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	-- invalidation VARCHAR(100)
);

call add_contribution_and_invalidation_columns(
	p_table => 'films',
	p_contributer_table => 'users',
	p_contributer_table_pk => 'id',
	
	p_column_contributed_by_name => 'contributed_by',
	p_column_contributed_by_type => 'INT NOT NULL',
	p_column_contributed_by_fk_constraint_name => 'films_contributed_by_fk_users',
	
	p_column_contributed_at_name => 'contributed_at',
	p_column_contributed_at_type => 'TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP',
	
	p_column_invalidation_name => 'invalidation',
	p_column_invalidation_type => 'VARCHAR(100)'
);

call create_audit_table(
	p_table => 'films',
	p_audit_table_name => 'films_audit',
	p_audit_table_pk_columns_order_sep_by_comma => 'id, contributed_by, contributed_at'
);

call build_trigger_audit_on_update(
	p_table => 'films',
	p_table_contributed_at_column => 'contributed_at',
	p_audit_table_name => 'films_audit',
	p_trigger_name => 'films_trigger_audit_on_update',
	p_trigger_function_name => 'films_function_triggers_on_update'
);

-- -- create a rule to not delete records on delete
-- CREATE OR REPLACE RULE films_rule_on_delete_do_nothing_instead
-- AS ON DELETE TO films
-- DO INSTEAD NOTHING;

-- -- create a rule to also insert old records into audit table on update
-- CREATE OR REPLACE RULE films_rule_audit_on_update
-- AS ON UPDATE TO films
-- DO ALSO
--     INSERT INTO films_audit SELECT OLD.*;

-- unique episodes
-- create a unique constraint on (series_id, season_number, episode_number)
ALTER TABLE films
ADD CONSTRAINT films_unique_episode_cnst
UNIQUE (series_id, season_number, episode_number);

-- -- unique movies
-- -- it is so wrong!
-- -- imagine a movie can replace an episode with the same title!
-- -- so commenting out for more discussion in the future
-- ALTER TABLE films
-- ADD CONSTRAINT films_unique_cnst
-- UNIQUE (title, date_released);

-- create index on title and descriptions
CREATE INDEX films_idx_title ON films (title);
CREATE INDEX films_idx_descriptions ON films (descriptions);

-- create index on date_released
CREATE INDEX films_idx_date_released ON films (date_released);

-- create index on duration
CREATE INDEX films_idx_duration ON films (duration);

-- add series_id foreign key constraint
ALTER TABLE IF EXISTS films
    ADD CONSTRAINT films_fk_serieses
    FOREIGN KEY (series_id)
    REFERENCES serieses(id);

COMMIT;