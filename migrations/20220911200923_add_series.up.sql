BEGIN;

-- create serieses table
CREATE TABLE IF NOT EXISTS serieses (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    descriptions VARCHAR(500),

    date_started DATE NOT NULL,
    date_ended DATE
	-- date_started TIMESTAMPTZ NOT NULL,
    -- date_ended TIMESTAMPTZ
    
    -- contributed_by INT NOT NULL,
    -- contributed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	-- invalidation VARCHAR(100)
);

call add_contribution_and_invalidation_columns(
	p_table => 'serieses',
	p_contributer_table => 'users',
	p_contributer_table_pk => 'id',
	
	p_column_contributed_by_name => 'contributed_by',
	p_column_contributed_by_type => 'INT NOT NULL',
	p_column_contributed_by_fk_constraint_name => 'serieses_contributed_by_fk_users',
	
	p_column_contributed_at_name => 'contributed_at',
	p_column_contributed_at_type => 'TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP',
	
	p_column_invalidation_name => 'invalidation',
	p_column_invalidation_type => 'VARCHAR(100)'
);

call create_audit_table(
	p_table => 'serieses',
	p_audit_table_name => 'serieses_audit',
	p_audit_table_pk_columns_order_sep_by_comma => 'id, contributed_by, contributed_at'
);

call build_trigger_audit_on_update(
	p_table => 'serieses',
	p_table_contributed_at_column => 'contributed_at',
	p_audit_table_name => 'serieses_audit',
	p_trigger_name => 'serieses_trigger_audit_on_update',
	p_trigger_function_name => 'serieses_function_triggers_on_update'
);

-- -- create a rule to not delete records on delete
-- CREATE OR REPLACE RULE serieses_rule_on_delete_do_nothing_instead
-- AS ON DELETE TO serieses
-- DO INSTEAD NOTHING;

-- -- create a rule to also insert old records into audit table on update
-- CREATE OR REPLACE RULE serieses_rule_audit_on_update
-- AS ON UPDATE TO serieses
-- DO ALSO
--     INSERT INTO serieses_audit SELECT OLD.*;

-- create index on title and descriptions
CREATE INDEX serieses_idx_title ON serieses (title);
CREATE INDEX serieses_idx_descriptions ON serieses (descriptions);

-- create index on date_started and date_ended
CREATE INDEX serieses_idx_date_started ON serieses (date_started);
CREATE INDEX serieses_idx_date_ended ON serieses (date_ended);

COMMIT;