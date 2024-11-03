-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createrole.html
https://www.postgresql.org/docs/17/sql-alterrole.html
https://www.postgresql.org/docs/17/sql-droprole.html
https://www.postgresql.org/docs/17/sql-set-role.html

https://www.postgresql.org/docs/17/sql-drop-owned.html
https://www.postgresql.org/docs/17/sql-reassign-owned.html
https://www.postgresql.org/docs/17/sql-creategroup.html
*/

CREATE ROLE jonathan LOGIN;

CREATE USER davide WITH PASSWORD 'jw8s0F4';

CREATE ROLE miriam WITH LOGIN PASSWORD 'jw8s0F4' VALID UNTIL '2005-01-01';

CREATE ROLE admin WITH CREATEDB CREATEROLE;

ALTER ROLE davide WITH PASSWORD 'hu8jmn3';

ALTER ROLE davide WITH PASSWORD NULL;

ALTER ROLE chris VALID UNTIL 'May 4 12:00:00 2015 +1';

ALTER ROLE fred VALID UNTIL 'infinity';

ALTER ROLE miriam CREATEROLE CREATEDB;

ALTER ROLE worker_bee SET maintenance_work_mem = 100000;

ALTER ROLE fred IN DATABASE devel SET client_min_messages = DEBUG;

DROP ROLE jonathan;


/*

CREATE ROLE name [ [ WITH ] option [ ... ] ]

where option can be:

      SUPERUSER | NOSUPERUSER
    | CREATEDB | NOCREATEDB
    | CREATEROLE | NOCREATEROLE
    | INHERIT | NOINHERIT
    | LOGIN | NOLOGIN
    | REPLICATION | NOREPLICATION
    | BYPASSRLS | NOBYPASSRLS
    | CONNECTION LIMIT connlimit
    | [ ENCRYPTED ] PASSWORD 'password' | PASSWORD NULL
    | VALID UNTIL 'timestamp'
    | IN ROLE role_name [, ...]
    | ROLE role_name [, ...]
    | ADMIN role_name [, ...]
    | SYSID uid

ALTER ROLE role_specification [ WITH ] option [ ... ]

where option can be:

      SUPERUSER | NOSUPERUSER
    | CREATEDB | NOCREATEDB
    | CREATEROLE | NOCREATEROLE
    | INHERIT | NOINHERIT
    | LOGIN | NOLOGIN
    | REPLICATION | NOREPLICATION
    | BYPASSRLS | NOBYPASSRLS
    | CONNECTION LIMIT connlimit
    | [ ENCRYPTED ] PASSWORD 'password' | PASSWORD NULL
    | VALID UNTIL 'timestamp'

ALTER ROLE name RENAME TO new_name

ALTER ROLE { role_specification | ALL } [ IN DATABASE database_name ] SET configuration_parameter { TO | = } { value | DEFAULT }
ALTER ROLE { role_specification | ALL } [ IN DATABASE database_name ] SET configuration_parameter FROM CURRENT
ALTER ROLE { role_specification | ALL } [ IN DATABASE database_name ] RESET configuration_parameter
ALTER ROLE { role_specification | ALL } [ IN DATABASE database_name ] RESET ALL

where role_specification can be:

    role_name
  | CURRENT_ROLE
  | CURRENT_USER
  | SESSION_USER

DROP ROLE [ IF EXISTS ] name [, ...]


SET [ SESSION | LOCAL ] ROLE role_name
SET [ SESSION | LOCAL ] ROLE NONE
RESET ROLE

DROP OWNED BY { name | CURRENT_ROLE | CURRENT_USER | SESSION_USER } [, ...] [ CASCADE | RESTRICT ]

REASSIGN OWNED BY { old_role | CURRENT_ROLE | CURRENT_USER | SESSION_USER } [, ...]
               TO { new_role | CURRENT_ROLE | CURRENT_USER | SESSION_USER }


*/
