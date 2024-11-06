-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtsparser.html
https://www.postgresql.org/docs/17/sql-altertsparser.html
https://www.postgresql.org/docs/17/sql-droptsparser.html
*/

ALTER TEXT SEARCH PARSER my_parser RENAME TO my_new_parser ;
ALTER TEXT SEARCH PARSER my_parser SET SCHEMA new_schema ;

DROP TEXT SEARCH PARSER my_parser;

/*

CREATE TEXT SEARCH PARSER name (
    START = start_function ,
    GETTOKEN = gettoken_function ,
    END = end_function ,
    LEXTYPES = lextypes_function
    [, HEADLINE = headline_function ]
)

ALTER TEXT SEARCH PARSER name RENAME TO new_name
ALTER TEXT SEARCH PARSER name SET SCHEMA new_schema

DROP TEXT SEARCH PARSER [ IF EXISTS ] name [ CASCADE | RESTRICT ]


*/
