-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtstemplate.html
https://www.postgresql.org/docs/17/sql-altertstemplate.html
https://www.postgresql.org/docs/17/sql-droptstemplate.html
*/

ALTER TEXT SEARCH TEMPLATE old_name RENAME TO new_name ;
ALTER TEXT SEARCH TEMPLATE old_name SET SCHEMA new_schema ;

DROP TEXT SEARCH TEMPLATE thesaurus;

/*

CREATE TEXT SEARCH TEMPLATE name (
    [ INIT = init_function , ]
    LEXIZE = lexize_function
)

ALTER TEXT SEARCH TEMPLATE name RENAME TO new_name
ALTER TEXT SEARCH TEMPLATE name SET SCHEMA new_schema

DROP TEXT SEARCH TEMPLATE [ IF EXISTS ] name [ CASCADE | RESTRICT ]


*/
