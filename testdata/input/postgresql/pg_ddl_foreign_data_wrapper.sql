-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createforeigndatawrapper.html
https://www.postgresql.org/docs/17/sql-alterforeigndatawrapper.html
https://www.postgresql.org/docs/17/sql-dropforeigndatawrapper.html
*/

CREATE FOREIGN DATA WRAPPER dummy;

CREATE FOREIGN DATA WRAPPER file HANDLER file_fdw_handler;

CREATE FOREIGN DATA WRAPPER mywrapper
    OPTIONS (debug 'true');


ALTER FOREIGN DATA WRAPPER dbi OPTIONS (ADD foo '1', DROP bar);

ALTER FOREIGN DATA WRAPPER dbi VALIDATOR bob.myvalidator;

DROP FOREIGN DATA WRAPPER dbi;

/*
CREATE FOREIGN DATA WRAPPER name
    [ HANDLER handler_function | NO HANDLER ]
    [ VALIDATOR validator_function | NO VALIDATOR ]
    [ OPTIONS ( option 'value' [, ... ] ) ]


ALTER FOREIGN DATA WRAPPER name
    [ HANDLER handler_function | NO HANDLER ]
    [ VALIDATOR validator_function | NO VALIDATOR ]
    [ OPTIONS ( [ ADD | SET | DROP ] option ['value'] [, ... ]) ]
ALTER FOREIGN DATA WRAPPER name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER FOREIGN DATA WRAPPER name RENAME TO new_name

DROP FOREIGN DATA WRAPPER [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]

*/
