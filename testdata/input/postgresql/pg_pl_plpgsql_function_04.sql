-- sqlfmt d:postgres

SET search_path = tasker, pg_catalog ;

CREATE OR REPLACE FUNCTION activity_is_parent_of (
    a_activity_id integer,
    a_parent_id integer )
RETURNS boolean
-- Set a secure search_path
SET search_path = tasker, pg_catalog, pg_temp
AS $$
DECLARE
    l_rec record ;

BEGIN
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

    EXCEPTION
        WHEN others THEN
            GET STACKED DIAGNOSTICS l_pg_cx = PG_CONTEXT,
                                    l_pg_ed = PG_EXCEPTION_DETAIL,
                                    l_pg_ec = PG_EXCEPTION_CONTEXT ;
            l_err := format ( '%s - %s:\n    %s\n     %s\n   %s', SQLSTATE, SQLERRM, l_pg_cx, l_pg_ed, l_pg_ec ) ;
            call util_log.log_exception ( l_err ) ;
            RAISE NOTICE E'EXCEPTION: %', l_err ;

    END ;

    RETURN false ;

END ;
$$
STABLE
SECURITY DEFINER
LANGUAGE plpgsql

 ;

ALTER FUNCTION activity_is_parent_of ( integer, integer ) OWNER TO tasker_owner ;

GRANT ALL ON FUNCTION activity_is_parent_of ( integer, integer ) TO tasker_user ;

REVOKE ALL ON FUNCTION activity_is_parent_of ( integer, integer ) FROM public ;
