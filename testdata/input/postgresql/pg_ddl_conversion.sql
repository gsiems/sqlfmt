-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createconversion.html
https://www.postgresql.org/docs/17/sql-alterconversion.html
https://www.postgresql.org/docs/17/sql-dropconversion.html
*/

CREATE CONVERSION myconv FOR 'UTF8' TO 'LATIN1' FROM myfunc;

ALTER CONVERSION iso_8859_1_to_utf8 RENAME TO latin1_to_unicode;
ALTER CONVERSION iso_8859_1_to_utf8 OWNER TO joe;

DROP CONVERSION myname;

/*
CREATE [ DEFAULT ] CONVERSION name
    FOR source_encoding TO dest_encoding FROM function_name

ALTER CONVERSION name RENAME TO new_name
ALTER CONVERSION name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER CONVERSION name SET SCHEMA new_schema

DROP CONVERSION [ IF EXISTS ] name [ CASCADE | RESTRICT ]
*/
