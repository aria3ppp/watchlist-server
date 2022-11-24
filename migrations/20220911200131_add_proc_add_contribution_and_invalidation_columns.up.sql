BEGIN;

-- add complement contributed_by, contributed_at and invalidation columns to table
create or replace procedure add_contribution_and_invalidation_columns(
	p_table text,
	p_contributer_table text,
	p_contributer_table_pk text,
	
	p_column_contributed_by_name text,
	p_column_contributed_by_type text,
	p_column_contributed_by_fk_constraint_name text,
	
	p_column_contributed_at_name text,
	p_column_contributed_at_type text,
	
	p_column_invalidation_name text,
	p_column_invalidation_type text
)
language plpgsql
as $$
begin
	-- check table exists
	perform from information_schema.tables
	where table_name = p_table and table_type = 'BASE TABLE';
	
	if not found then
		raise exception 'table name "%" not found', p_table;
	end if;
	
	-- check contributer referenced table exists
	perform from information_schema.tables
	where table_name = p_contributer_table and table_type = 'BASE TABLE';
	
	if not found then
		raise exception 'contributer table name "%" not found', p_contributer_table;
	end if;
	
	-- add columns
	execute 'ALTER TABLE ' || quote_ident(p_table) || ' '
		|| 'ADD COLUMN ' || quote_ident(p_column_contributed_by_name) || ' ' || p_column_contributed_by_type || ', '
		|| 'ADD COLUMN ' || quote_ident(p_column_contributed_at_name) || ' ' || p_column_contributed_at_type || ', '
		|| 'ADD COLUMN ' || quote_ident(p_column_invalidation_name) || ' ' || p_column_invalidation_type;
				   
	-- add contributed_by foreign key
	execute 'ALTER TABLE ' || quote_ident(p_table) || ' '
				|| 'ADD CONSTRAINT ' || quote_ident(p_column_contributed_by_fk_constraint_name) || ' '
				|| 'FOREIGN KEY (' || quote_ident(p_column_contributed_by_name) || ') '
				|| 'REFERENCES ' || quote_ident(p_contributer_table) || ' (' || quote_ident(p_contributer_table_pk) || ')';
				
end;
$$;

COMMIT;