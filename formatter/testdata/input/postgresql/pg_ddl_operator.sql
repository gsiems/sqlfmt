-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createoperator.html
https://www.postgresql.org/docs/17/sql-alteroperator.html
https://www.postgresql.org/docs/17/sql-dropoperator.html
*/

CREATE OPERATOR === (
    LEFTARG = box,
    RIGHTARG = box,
    FUNCTION = area_equal_function,
    COMMUTATOR = ===,
    NEGATOR = !==,
    RESTRICT = area_restriction_function,
    JOIN = area_join_function,
    HASHES, MERGES
);

ALTER OPERATOR @@ (text, text) OWNER TO joe;

ALTER OPERATOR && (int[], int[]) SET (RESTRICT = _int_contsel, JOIN = _int_contjoinsel);

ALTER OPERATOR && (int[], int[]) SET (COMMUTATOR = &&);


DROP OPERATOR ^ (integer, integer);

DROP OPERATOR ~ (none, bit);

DROP OPERATOR ~ (none, bit), ^ (integer, integer);


/*

CREATE OPERATOR name (
    {FUNCTION|PROCEDURE} = function_name
    [, LEFTARG = left_type ] [, RIGHTARG = right_type ]
    [, COMMUTATOR = com_op ] [, NEGATOR = neg_op ]
    [, RESTRICT = res_proc ] [, JOIN = join_proc ]
    [, HASHES ] [, MERGES ]
)

ALTER OPERATOR name ( { left_type | NONE } , right_type )
    OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }

ALTER OPERATOR name ( { left_type | NONE } , right_type )
    SET SCHEMA new_schema

ALTER OPERATOR name ( { left_type | NONE } , right_type )
    SET ( {  RESTRICT = { res_proc | NONE }
           | JOIN = { join_proc | NONE }
           | COMMUTATOR = com_op
           | NEGATOR = neg_op
           | HASHES
           | MERGES
          } [, ... ] )

*/
