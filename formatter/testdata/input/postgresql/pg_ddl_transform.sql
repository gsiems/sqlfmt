-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtransform.html
https://www.postgresql.org/docs/17/sql-droptransform.html
*/

CREATE TRANSFORM FOR hstore LANGUAGE plpython3u (
    FROM SQL WITH FUNCTION hstore_to_plpython(internal),
    TO SQL WITH FUNCTION plpython_to_hstore(internal)
);

DROP TRANSFORM FOR hstore LANGUAGE plpython3u;

/*

CREATE [ OR REPLACE ] TRANSFORM FOR type_name LANGUAGE lang_name (
    FROM SQL WITH FUNCTION from_sql_function_name [ (argument_type [, ...]) ],
    TO SQL WITH FUNCTION to_sql_function_name [ (argument_type [, ...]) ]
);

DROP TRANSFORM [ IF EXISTS ] FOR type_name LANGUAGE lang_name [ CASCADE | RESTRICT ]

*/
