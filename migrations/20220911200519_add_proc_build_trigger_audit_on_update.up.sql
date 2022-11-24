BEGIN;

-- build a trigger that audit old records on update
create or replace procedure build_trigger_audit_on_update(
	p_table text,
	p_table_contributed_at_column text,
	p_audit_table_name text,
	p_trigger_name text,
	p_trigger_function_name text
)
language plpgsql
as $body$
declare
	v_trigger_func_body text;
	v_trigger_func_cmd text;
	v_create_trigger_on_table_cmd text;
begin
	-- build trigger function
	v_trigger_func_body = 'BEGIN '
			|| 'INSERT INTO ' || quote_ident(p_audit_table_name) || ' SELECT OLD.*; '
			|| 'NEW.' || p_table_contributed_at_column || ' = CURRENT_TIMESTAMP; '
			|| 'RETURN NEW; '
			|| 'END;';
	
	v_trigger_func_cmd = 'CREATE OR REPLACE FUNCTION ' || p_trigger_function_name || '() RETURNS TRIGGER ' 
						|| 'LANGUAGE plpgsql AS $$ ' || v_trigger_func_body || ' $$';
	
	-- raise notice 'trigger func cmd: %', v_trigger_func_cmd;
	
	-- create trigger function
	execute v_trigger_func_cmd;
	
	-- build trigger on table
	v_create_trigger_on_table_cmd = 'CREATE TRIGGER ' || p_trigger_name || ' '
									|| 'BEFORE UPDATE ON ' || p_table || ' '
									|| 'FOR EACH ROW EXECUTE FUNCTION ' || p_trigger_function_name || '()';
									
	-- raise notice 'create trigger cmd: %', v_create_trigger_on_table_cmd;
	
	-- create trigger on table
	execute v_create_trigger_on_table_cmd;
	
end;
$body$;

COMMIT;