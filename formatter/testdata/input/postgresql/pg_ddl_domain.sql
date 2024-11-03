-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createdomain.html
https://www.postgresql.org/docs/17/sql-alterdomain.html
https://www.postgresql.org/docs/17/sql-dropdomain.html
*/

CREATE DOMAIN us_postal_code AS TEXT
CHECK(
   VALUE ~ '^\d{5}$'
OR VALUE ~ '^\d{5}-\d{4}$'
);

CREATE TABLE us_snail_addy (
  address_id SERIAL PRIMARY KEY,
  street1 TEXT NOT NULL,
  street2 TEXT,
  street3 TEXT,
  city TEXT NOT NULL,
  postal us_postal_code NOT NULL
);


ALTER DOMAIN zipcode SET NOT NULL;

ALTER DOMAIN zipcode DROP NOT NULL;

ALTER DOMAIN zipcode ADD CONSTRAINT zipchk CHECK (char_length(VALUE) = 5);

ALTER DOMAIN zipcode DROP CONSTRAINT zipchk;

ALTER DOMAIN zipcode RENAME CONSTRAINT zipchk TO zip_check;

ALTER DOMAIN zipcode SET SCHEMA customers;

DROP DOMAIN box;

/*
CREATE DOMAIN name [ AS ] data_type
    [ COLLATE collation ]
    [ DEFAULT expression ]
    [ domain_constraint [ ... ] ]

where domain_constraint is:

[ CONSTRAINT constraint_name ]
{ NOT NULL | NULL | CHECK (expression) }


ALTER DOMAIN name
    { SET DEFAULT expression | DROP DEFAULT }
ALTER DOMAIN name
    { SET | DROP } NOT NULL
ALTER DOMAIN name
    ADD domain_constraint [ NOT VALID ]
ALTER DOMAIN name
    DROP CONSTRAINT [ IF EXISTS ] constraint_name [ RESTRICT | CASCADE ]
ALTER DOMAIN name
     RENAME CONSTRAINT constraint_name TO new_constraint_name
ALTER DOMAIN name
    VALIDATE CONSTRAINT constraint_name
ALTER DOMAIN name
    OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER DOMAIN name
    RENAME TO new_name
ALTER DOMAIN name
    SET SCHEMA new_schema

DROP DOMAIN [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]

*/
