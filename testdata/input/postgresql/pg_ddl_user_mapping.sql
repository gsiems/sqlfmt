-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createusermapping.html
https://www.postgresql.org/docs/17/sql-alterusermapping.html
https://www.postgresql.org/docs/17/sql-dropusermapping.html
*/

CREATE USER MAPPING FOR bob SERVER foo OPTIONS (user 'bob', password 'secret');

ALTER USER MAPPING FOR bob SERVER foo OPTIONS (SET password 'public');

DROP USER MAPPING IF EXISTS FOR bob SERVER foo;

/*
CREATE USER MAPPING [ IF NOT EXISTS ] FOR { user_name | USER | CURRENT_ROLE | CURRENT_USER | PUBLIC }
    SERVER server_name
    [ OPTIONS ( option 'value' [ , ... ] ) ]

ALTER USER MAPPING FOR { user_name | USER | CURRENT_ROLE | CURRENT_USER | SESSION_USER | PUBLIC }
    SERVER server_name
    OPTIONS ( [ ADD | SET | DROP ] option ['value'] [, ... ] )

DROP USER MAPPING [ IF EXISTS ] FOR { user_name | USER | CURRENT_ROLE | CURRENT_USER | PUBLIC } SERVER server_name
*/
