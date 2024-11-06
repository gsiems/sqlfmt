-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-importforeignschema.html
*/

IMPORT FOREIGN SCHEMA foreign_films
    FROM SERVER film_server INTO films;

IMPORT FOREIGN SCHEMA foreign_films LIMIT TO (actors, directors)
    FROM SERVER film_server INTO films;
