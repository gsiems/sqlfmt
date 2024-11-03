-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createschema.html
https://www.postgresql.org/docs/17/sql-alterschema.html
https://www.postgresql.org/docs/17/sql-dropschema.html
*/

CREATE SCHEMA myschema;

CREATE SCHEMA AUTHORIZATION joe;

CREATE SCHEMA IF NOT EXISTS test AUTHORIZATION joe;

DROP SCHEMA mystuff CASCADE;


/*
CREATE SCHEMA schema_name [ AUTHORIZATION role_specification ] [ schema_element [ ... ] ]
CREATE SCHEMA AUTHORIZATION role_specification [ schema_element [ ... ] ]
CREATE SCHEMA IF NOT EXISTS schema_name [ AUTHORIZATION role_specification ]
CREATE SCHEMA IF NOT EXISTS AUTHORIZATION role_specification

where role_specification can be:

    user_name
  | CURRENT_ROLE
  | CURRENT_USER
  | SESSION_USER

ALTER SCHEMA name RENAME TO new_name
ALTER SCHEMA name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }

DROP SCHEMA [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]

*/
