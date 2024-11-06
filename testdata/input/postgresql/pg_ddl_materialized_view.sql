-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-creatematerializedview.html
https://www.postgresql.org/docs/17/sql-altermaterializedview.html
https://www.postgresql.org/docs/17/sql-dropmaterializedview.html
https://www.postgresql.org/docs/17/sql-refreshmaterializedview.html
*/

CREATE MATERIALIZED VIEW annual_statistics_basis
TABLESPACE reporting
WITH (fillfactor=70)
AS
SELECT *
    FROM v_annual_statistics_basis
    WITH NO DATA ;


ALTER MATERIALIZED VIEW foo RENAME TO bar;

DROP MATERIALIZED VIEW order_summary;

REFRESH MATERIALIZED VIEW order_summary;

REFRESH MATERIALIZED VIEW CONCURRENTLY annual_statistics_basis WITH NO DATA;

REFRESH MATERIALIZED VIEW annual_statistics_basis WITH NO DATA;

REFRESH MATERIALIZED VIEW annual_statistics_basis WITH DATA;

/*
CREATE MATERIALIZED VIEW [ IF NOT EXISTS ] table_name
    [ (column_name [, ...] ) ]
    [ USING method ]
    [ WITH ( storage_parameter [= value] [, ... ] ) ]
    [ TABLESPACE tablespace_name ]
    AS query
    [ WITH [ NO ] DATA ]

ALTER MATERIALIZED VIEW [ IF EXISTS ] name
    action [, ... ]
ALTER MATERIALIZED VIEW name
    [ NO ] DEPENDS ON EXTENSION extension_name
ALTER MATERIALIZED VIEW [ IF EXISTS ] name
    RENAME [ COLUMN ] column_name TO new_column_name
ALTER MATERIALIZED VIEW [ IF EXISTS ] name
    RENAME TO new_name
ALTER MATERIALIZED VIEW [ IF EXISTS ] name
    SET SCHEMA new_schema
ALTER MATERIALIZED VIEW ALL IN TABLESPACE name [ OWNED BY role_name [, ... ] ]
    SET TABLESPACE new_tablespace [ NOWAIT ]

where action is one of:

    ALTER [ COLUMN ] column_name SET STATISTICS integer
    ALTER [ COLUMN ] column_name SET ( attribute_option = value [, ... ] )
    ALTER [ COLUMN ] column_name RESET ( attribute_option [, ... ] )
    ALTER [ COLUMN ] column_name SET STORAGE { PLAIN | EXTERNAL | EXTENDED | MAIN | DEFAULT }
    ALTER [ COLUMN ] column_name SET COMPRESSION compression_method
    CLUSTER ON index_name
    SET WITHOUT CLUSTER
    SET ACCESS METHOD new_access_method
    SET TABLESPACE new_tablespace
    SET ( storage_parameter [= value] [, ... ] )
    RESET ( storage_parameter [, ... ] )
    OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }

DROP MATERIALIZED VIEW [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]

REFRESH MATERIALIZED VIEW [ CONCURRENTLY ] name
    [ WITH [ NO ] DATA ]


*/
