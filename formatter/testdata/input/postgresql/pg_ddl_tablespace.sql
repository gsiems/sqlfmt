-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtablespace.html
https://www.postgresql.org/docs/17/sql-altertablespace.html
https://www.postgresql.org/docs/17/sql-droptablespace.html
*/

CREATE TABLESPACE dbspace LOCATION '/data/dbs';

CREATE TABLESPACE indexspace OWNER genevieve LOCATION '/data/indexes';

ALTER TABLESPACE index_space RENAME TO fast_raid;

ALTER TABLESPACE index_space OWNER TO mary;

DROP TABLESPACE mystuff;

/*

CREATE TABLESPACE tablespace_name
    [ OWNER { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER } ]
    LOCATION 'directory'
    [ WITH ( tablespace_option = value [, ... ] ) ]

ALTER TABLESPACE name RENAME TO new_name
ALTER TABLESPACE name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER TABLESPACE name SET ( tablespace_option = value [, ... ] )
ALTER TABLESPACE name RESET ( tablespace_option [, ... ] )

DROP TABLESPACE [ IF EXISTS ] name

*/
