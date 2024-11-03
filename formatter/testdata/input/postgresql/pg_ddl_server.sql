-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createserver.html
https://www.postgresql.org/docs/17/sql-alterserver.html
https://www.postgresql.org/docs/17/sql-dropserver.html
*/

CREATE SERVER myserver FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host 'foo', dbname 'foodb', port '5432');

ALTER SERVER foo OPTIONS (host 'foo', dbname 'foodb');

ALTER SERVER foo VERSION '8.4' OPTIONS (SET host 'baz');

DROP SERVER IF EXISTS foo;

/*

CREATE SERVER [ IF NOT EXISTS ] server_name [ TYPE 'server_type' ] [ VERSION 'server_version' ]
    FOREIGN DATA WRAPPER fdw_name
    [ OPTIONS ( option 'value' [, ... ] ) ]

ALTER SERVER name [ VERSION 'new_version' ]
    [ OPTIONS ( [ ADD | SET | DROP ] option ['value'] [, ... ] ) ]
ALTER SERVER name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER SERVER name RENAME TO new_name

DROP SERVER [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]

*/
