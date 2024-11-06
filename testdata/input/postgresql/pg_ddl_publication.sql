-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createpublication.html
https://www.postgresql.org/docs/17/sql-alterpublication.html
https://www.postgresql.org/docs/17/sql-droppublication.html
*/

CREATE PUBLICATION mypublication FOR TABLE users, departments;

CREATE PUBLICATION active_departments FOR TABLE departments WHERE (active IS TRUE);

CREATE PUBLICATION alltables FOR ALL TABLES;

CREATE PUBLICATION insert_only FOR TABLE mydata
    WITH (publish = 'insert');

CREATE PUBLICATION production_publication FOR TABLE users, departments, TABLES IN SCHEMA production;

CREATE PUBLICATION sales_publication FOR TABLES IN SCHEMA marketing, sales;

CREATE PUBLICATION users_filtered FOR TABLE users (user_id, firstname);


ALTER PUBLICATION noinsert SET (publish = 'update, delete');

ALTER PUBLICATION mypublication ADD TABLE users (user_id, firstname), departments;

ALTER PUBLICATION mypublication SET TABLE users (user_id, firstname, lastname), TABLE departments;

ALTER PUBLICATION sales_publication ADD TABLES IN SCHEMA marketing, sales;

ALTER PUBLICATION production_publication ADD TABLE users, departments, TABLES IN SCHEMA production;


DROP PUBLICATION mypublication;

/*

CREATE PUBLICATION name
    [ FOR ALL TABLES
      | FOR publication_object [, ... ] ]
    [ WITH ( publication_parameter [= value] [, ... ] ) ]

where publication_object is one of:

    TABLE [ ONLY ] table_name [ * ] [ ( column_name [, ... ] ) ] [ WHERE ( expression ) ] [, ... ]
    TABLES IN SCHEMA { schema_name | CURRENT_SCHEMA } [, ... ]

ALTER PUBLICATION name ADD publication_object [, ...]
ALTER PUBLICATION name SET publication_object [, ...]
ALTER PUBLICATION name DROP publication_object [, ...]
ALTER PUBLICATION name SET ( publication_parameter [= value] [, ... ] )
ALTER PUBLICATION name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER PUBLICATION name RENAME TO new_name

where publication_object is one of:

    TABLE [ ONLY ] table_name [ * ] [ ( column_name [, ... ] ) ] [ WHERE ( expression ) ] [, ... ]
    TABLES IN SCHEMA { schema_name | CURRENT_SCHEMA } [, ... ]

DROP PUBLICATION [ IF EXISTS ] name [, ...] [ CASCADE | RESTRICT ]


*/
