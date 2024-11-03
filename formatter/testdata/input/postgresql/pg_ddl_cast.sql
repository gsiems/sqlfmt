-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createcast.html
https://www.postgresql.org/docs/17/sql-dropcast.html
*/

CREATE CAST (bigint AS int4) WITH FUNCTION int4(bigint) AS ASSIGNMENT;

DROP CAST (text AS int);

/*
CREATE CAST (source_type AS target_type)
    WITH FUNCTION function_name [ (argument_type [, ...]) ]
    [ AS ASSIGNMENT | AS IMPLICIT ]

CREATE CAST (source_type AS target_type)
    WITHOUT FUNCTION
    [ AS ASSIGNMENT | AS IMPLICIT ]

CREATE CAST (source_type AS target_type)
    WITH INOUT
    [ AS ASSIGNMENT | AS IMPLICIT ]

DROP CAST [ IF EXISTS ] (source_type AS target_type) [ CASCADE | RESTRICT ]
*/
