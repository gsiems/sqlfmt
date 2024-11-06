-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createeventtrigger.html
https://www.postgresql.org/docs/17/sql-altereventtrigger.html
https://www.postgresql.org/docs/17/sql-dropeventtrigger.html
*/

CREATE OR REPLACE FUNCTION abort_any_command()
  RETURNS event_trigger
 LANGUAGE plpgsql
  AS $$
BEGIN
  RAISE EXCEPTION 'command % is disabled', tg_tag;
END;
$$;

CREATE EVENT TRIGGER abort_ddl ON ddl_command_start
   EXECUTE FUNCTION abort_any_command();

DROP EVENT TRIGGER snitch;

/*
CREATE EVENT TRIGGER name
    ON event
    [ WHEN filter_variable IN (filter_value [, ... ]) [ AND ... ] ]
    EXECUTE { FUNCTION | PROCEDURE } function_name()

ALTER EVENT TRIGGER name DISABLE
ALTER EVENT TRIGGER name ENABLE [ REPLICA | ALWAYS ]
ALTER EVENT TRIGGER name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER EVENT TRIGGER name RENAME TO new_name

DROP EVENT TRIGGER [ IF EXISTS ] name [ CASCADE | RESTRICT ]

*/
