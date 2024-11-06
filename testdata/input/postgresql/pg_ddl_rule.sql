-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createrule.html
https://www.postgresql.org/docs/17/sql-alterrule.html
https://www.postgresql.org/docs/17/sql-droprule.html
*/

CREATE RULE "_RETURN" AS
    ON SELECT TO t1
    DO INSTEAD
        SELECT * FROM t2;

CREATE RULE "_RETURN" AS
    ON SELECT TO t2
    DO INSTEAD
        SELECT * FROM t1;


CREATE RULE notify_me AS ON UPDATE TO mytable DO ALSO NOTIFY mytable;

ALTER RULE notify_all ON emp RENAME TO notify_me;

DROP RULE newrule ON mytable;

/*

CREATE [ OR REPLACE ] RULE name AS ON event
    TO table_name [ WHERE condition ]
    DO [ ALSO | INSTEAD ] { NOTHING | command | ( command ; command ... ) }

where event can be one of:

    SELECT | INSERT | UPDATE | DELETE

ALTER RULE name ON table_name RENAME TO new_name

DROP RULE [ IF EXISTS ] name ON table_name [ CASCADE | RESTRICT ]



*/
