BEGIN;

-- create an audit table with a primary key index on specified columns
create or replace procedure create_audit_table(
	p_table text,
	p_audit_table_name text,
	p_audit_table_pk_columns_order_sep_by_comma text
)
language plpgsql
as $$
declare
	v_row RECORD;
    v_CREATE_AUDIT_TABLE_BODY TEXT;
	v_CREATE_AUDIT_TABLE_CMD TEXT;
begin
	perform from information_schema.tables
	where table_name = p_table and table_type = 'BASE TABLE';
	
	if not found then
		raise exception 'table name "%" not found', p_table;
	end if;

    v_CREATE_AUDIT_TABLE_BODY = '';
	
	for v_row in
		select column_name, data_type, is_nullable
		from information_schema.columns
		where table_name = p_table
		order by ordinal_position
	loop
	
		v_CREATE_AUDIT_TABLE_BODY = v_CREATE_AUDIT_TABLE_BODY || quote_ident(v_row.column_name) || ' ' || v_row.data_type;
		
		if v_row.is_nullable = 'NO' then
			v_CREATE_AUDIT_TABLE_BODY = v_CREATE_AUDIT_TABLE_BODY || ' NOT NULL';
		end if;

		v_CREATE_AUDIT_TABLE_BODY = v_CREATE_AUDIT_TABLE_BODY || ', ';
		
	end loop;

	-- set primary key
	v_CREATE_AUDIT_TABLE_BODY = v_CREATE_AUDIT_TABLE_BODY || 'PRIMARY KEY (' || p_audit_table_pk_columns_order_sep_by_comma || ')';
    
	-- build create audit table command
	v_CREATE_AUDIT_TABLE_CMD = 'CREATE TABLE IF NOT EXISTS ' || quote_ident(p_audit_table_name) || ' (' || v_CREATE_AUDIT_TABLE_BODY || ')';
		
	-- create the audit table
	execute v_CREATE_AUDIT_TABLE_CMD;
	
end;
$$;

COMMIT;