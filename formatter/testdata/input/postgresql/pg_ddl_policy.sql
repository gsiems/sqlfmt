-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createpolicy.html
https://www.postgresql.org/docs/17/sql-alterpolicy.html
https://www.postgresql.org/docs/17/sql-droppolicy.html
*/

DROP POLICY p1 ON my_table;

/*

CREATE POLICY name ON table_name
    [ AS { PERMISSIVE | RESTRICTIVE } ]
    [ FOR { ALL | SELECT | INSERT | UPDATE | DELETE } ]
    [ TO { role_name | PUBLIC | CURRENT_ROLE | CURRENT_USER | SESSION_USER } [, ...] ]
    [ USING ( using_expression ) ]
    [ WITH CHECK ( check_expression ) ]

ALTER POLICY name ON table_name RENAME TO new_name

ALTER POLICY name ON table_name
    [ TO { role_name | PUBLIC | CURRENT_ROLE | CURRENT_USER | SESSION_USER } [, ...] ]
    [ USING ( using_expression ) ]
    [ WITH CHECK ( check_expression ) ]

DROP POLICY [ IF EXISTS ] name ON table_name [ CASCADE | RESTRICT ]



*/
