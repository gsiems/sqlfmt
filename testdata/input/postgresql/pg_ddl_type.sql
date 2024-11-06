-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtype.html
https://www.postgresql.org/docs/17/sql-altertype.html
https://www.postgresql.org/docs/17/sql-droptype.html
*/

CREATE TYPE compfoo AS (f1 int, f2 text);

CREATE TYPE bug_status AS ENUM ('new', 'open', 'closed');

CREATE TYPE float8_range AS RANGE (subtype = float8, subtype_diff = float8mi);

CREATE TYPE box;

CREATE TYPE box (
    INTERNALLENGTH = 16,
    INPUT = my_box_in_function,
    OUTPUT = my_box_out_function
);

CREATE TYPE bigobj (
    INPUT = lo_filein, OUTPUT = lo_fileout,
    INTERNALLENGTH = VARIABLE
);

CREATE TYPE foo.bar AS (
f1 int,
f2 text,
f3 numeric(5,2),
f4 varchar(10),
f5 date);

comment on compfoo is 'A spiffy new type';

ALTER TYPE electronic_mail RENAME TO email;

ALTER TYPE email OWNER TO joe;

ALTER TYPE email SET SCHEMA customers;

ALTER TYPE compfoo ADD ATTRIBUTE f3 int;

ALTER TYPE colors ADD VALUE 'orange' AFTER 'red';

ALTER TYPE colors RENAME VALUE 'purple' TO 'mauve';

ALTER TYPE mytype SET (
    SEND = mytypesend,
    RECEIVE = mytyperecv
);

DROP TYPE box;

CREATE TYPE "app_api"."ut_address" AS (
    "id" integer,
    "address1" text,
    "address2" text,
    "city" text,
    "stateId" integer,
    "state" text,
    "stateCode" text,
    "countryId" integer,
    "country" text,
    "countryCode" text,
    "postalCode" text,
    "createdTmsp" timestamp without time zone,
    "updatedTmsp" timestamp without time zone,
    "userIdCreated" integer,
    "createdBy" text,
    "userIdUpdated" integer,
    "updatedBy" text,
    "addressTypes" json ) ;

ALTER TYPE app_api.ut_address OWNER TO app_owner ;


/*
CREATE TYPE name AS
    ( [ attribute_name data_type [ COLLATE collation ] [, ... ] ] )

CREATE TYPE name AS ENUM
    ( [ 'label' [, ... ] ] )

CREATE TYPE name AS RANGE (
    SUBTYPE = subtype
    [ , SUBTYPE_OPCLASS = subtype_operator_class ]
    [ , COLLATION = collation ]
    [ , CANONICAL = canonical_function ]
    [ , SUBTYPE_DIFF = subtype_diff_function ]
    [ , MULTIRANGE_TYPE_NAME = multirange_type_name ]
)

CREATE TYPE name (
    INPUT = input_function,
    OUTPUT = output_function
    [ , RECEIVE = receive_function ]
    [ , SEND = send_function ]
    [ , TYPMOD_IN = type_modifier_input_function ]
    [ , TYPMOD_OUT = type_modifier_output_function ]
    [ , ANALYZE = analyze_function ]
    [ , SUBSCRIPT = subscript_function ]
    [ , INTERNALLENGTH = { internallength | VARIABLE } ]
    [ , PASSEDBYVALUE ]
    [ , ALIGNMENT = alignment ]
    [ , STORAGE = storage ]
    [ , LIKE = like_type ]
    [ , CATEGORY = category ]
    [ , PREFERRED = preferred ]
    [ , DEFAULT = default ]
    [ , ELEMENT = element ]
    [ , DELIMITER = delimiter ]
    [ , COLLATABLE = collatable ]
)

CREATE TYPE name

ALTER TYPE name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER TYPE name RENAME TO new_name
ALTER TYPE name SET SCHEMA new_schema
ALTER TYPE name RENAME ATTRIBUTE attribute_name TO new_attribute_name [ CASCADE | RESTRICT ]
ALTER TYPE name action [, ... ]
ALTER TYPE name ADD VALUE [ IF NOT EXISTS ] new_enum_value [ { BEFORE | AFTER } neighbor_enum_value ]
ALTER TYPE name RENAME VALUE existing_enum_value TO new_enum_value
ALTER TYPE name SET ( property = value [, ... ] )

where action is one of:

    ADD ATTRIBUTE attribute_name data_type [ COLLATE collation ] [ CASCADE | RESTRICT ]
    DROP ATTRIBUTE [ IF EXISTS ] attribute_name [ CASCADE | RESTRICT ]
    ALTER ATTRIBUTE attribute_name [ SET DATA ] TYPE data_type [ COLLATE collation ] [ CASCADE | RESTRICT ]

*/
