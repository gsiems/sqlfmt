--d:postgres

SET search_path = tasker, pg_catalog ;

CREATE OR REPLACE FUNCTION activity_is_parent_of (
    a_activity_id integer,
    a_parent_id integer )
RETURNS boolean
STABLE
SECURITY DEFINER
-- Set a secure search_path
SET search_path = tasker, pg_catalog, pg_temp
LANGUAGE plpgsql
AS $$
DECLARE
    l_rec record ;

BEGIN

    IF a_activity_id IS NULL OR a_parent_id IS NULL THEN
        RETURN false ;
    END IF ;

    IF a_activity_id = a_parent_id THEN
        RETURN false ;
    END IF ;

    FOR l_rec IN
        SELECT 1
            FROM tasker.dv_activity_tree dat
            WHERE dat.activity_id = a_activity_id
                AND a_parent_id = ANY ( dat.parents )
            LIMIT 1
        LOOP

        RETURN true ;

    END LOOP ;

    RETURN false ;

END ;
$$ ;

ALTER FUNCTION activity_is_parent_of ( integer, integer ) OWNER TO tasker_owner ;

GRANT ALL ON FUNCTION activity_is_parent_of ( integer, integer ) TO tasker_user ;

REVOKE ALL ON FUNCTION activity_is_parent_of ( integer, integer ) FROM public ;
