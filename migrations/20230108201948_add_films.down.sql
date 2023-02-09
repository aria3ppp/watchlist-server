BEGIN;

DROP FUNCTION IF EXISTS films_function_triggers_on_update;
DROP TABLE IF EXISTS films_audit;
DROP TABLE IF EXISTS films;

COMMIT;