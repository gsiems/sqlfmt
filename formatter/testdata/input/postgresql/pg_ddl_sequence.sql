-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createsequence.html
https://www.postgresql.org/docs/17/sql-altersequence.html
https://www.postgresql.org/docs/17/sql-dropsequence.html
*/

CREATE SEQUENCE serial START 101;

ALTER SEQUENCE serial RESTART WITH 105;

DROP SEQUENCE serial;

SELECT pg_catalog.setval('app_data.dt_stuff_id_seq', 15, true);


/*

CREATE [ { TEMPORARY | TEMP } | UNLOGGED ] SEQUENCE [ IF NOT EXISTS ] name
    [ AS data_type ]
    [ INCREMENT [ BY ] increment ]
    [ MINVALUE minvalue | NO MINVALUE ] [ MAXVALUE maxvalue | NO MAXVALUE ]
    [ START [ WITH ] start ] [ CACHE cache ] [ [ NO ] CYCLE ]
    [ OWNED BY { table_name.column_name | NONE } ]

ALTER SEQUENCE [ IF EXISTS ] name
    [ AS data_type ]
    [ INCREMENT [ BY ] increment ]
    [ MINVALUE minvalue | NO MINVALUE ] [ MAXVALUE maxvalue | NO MAXVALUE ]
    [ START [ WITH ] start ]
    [ RESTART [ [ WITH ] restart ] ]
    [ CACHE cache ] [ [ NO ] CYCLE ]
    [ OWNED BY { table_name.column_name | NONE } ]
ALTER SEQUENCE [ IF EXISTS ] name SET { LOGGED | UNLOGGED }
ALTER SEQUENCE [ IF EXISTS ] name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER SEQUENCE [ IF EXISTS ] name RENAME TO new_name
ALTER SEQUENCE [ IF EXISTS ] name SET SCHEMA new_schema

DROP SEQUENCE [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]


*/
