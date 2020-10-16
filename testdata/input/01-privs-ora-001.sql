--d:oracle

/* Some block of comments */

GRANT EXECUTE ON some_schema_name.some_procedure TO user_name ;
GRANT EXECUTE ON some_schema_name.some_other_procedure TO user_name ;






-- The extra linefeeds above should get cleaned up

GRANT SELECT ON some_schema_name.table_one TO user_name WITH GRANT OPTION ;
GRANT SELECT ON some_schema_name.table_two TO user_name ;

GRANT SELECT ON some_schema_name.table_one -- needed for blah, blah, blah
    TO user_name WITH GRANT OPTION ;

-- Comment
GRANT INSERT, SELECT, UPDATE ON some_schema_name.table_three TO user_name ;
GRANT DELETE, INSERT, SELECT ON some_schema_name.table_four TO user_name ;
-- Finito
