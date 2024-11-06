-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createcollation.html
https://www.postgresql.org/docs/17/sql-altercollation.html
https://www.postgresql.org/docs/17/sql-dropcollation.html
*/

CREATE COLLATION french (locale = 'fr_FR.utf8');

CREATE COLLATION german_phonebook (provider = icu, locale = 'de-u-co-phonebk');

CREATE COLLATION custom (provider = icu, locale = 'und', rules = '&V << w <<< W');

CREATE COLLATION german FROM "de_DE";

ALTER COLLATION "de_DE" RENAME TO german;
ALTER COLLATION "en_US" OWNER TO joe;

DROP COLLATION german;

/*
CREATE COLLATION [ IF NOT EXISTS ] name (
    [ LOCALE = locale, ]
    [ LC_COLLATE = lc_collate, ]
    [ LC_CTYPE = lc_ctype, ]
    [ PROVIDER = provider, ]
    [ DETERMINISTIC = boolean, ]
    [ RULES = rules, ]
    [ VERSION = version ]
)
CREATE COLLATION [ IF NOT EXISTS ] name FROM existing_collation

ALTER COLLATION name REFRESH VERSION

ALTER COLLATION name RENAME TO new_name
ALTER COLLATION name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER COLLATION name SET SCHEMA new_schema

DROP COLLATION [ IF EXISTS ] name [ CASCADE | RESTRICT ]

*/
