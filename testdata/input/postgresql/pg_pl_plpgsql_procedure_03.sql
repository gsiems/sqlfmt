-- sqlfmt d:postgres

CREATE OR REPLACE PROCEDURE widget.upsert_rt_widget_type (
    a_id inout integer,
    a_parent_id in integer,
    a_name in text,
    a_is_active in text,
    a_user in text DEFAULT NULL,
    a_err inout text DEFAULT NULL )
SECURITY DEFINER
SET search_path = pg_catalog, public
LANGUAGE plpgsql
AS $$
/**
Procedure upsert_rt_widget_type

| Parameter                  | In/Out | Datatype | Remarks                                          |
| -------------------------- | ------ | -------- | ------------------------------------------------ |
| a_id                       | inout  | integer  | The ID of the widget type                        |
| a_parent_id                | in     | integer  | The ID of the parent type (if any)               |
| a_name                     | in     | text     | The name of the widget type                      |
| a_is_active                | in     | text     | Indicates if the record is active (available for use for new data) |
| a_user                     | in     | text     | The ID or username of the user doing the insert  |
| a_err                      | inout  | text     | The error that was generated, if any |

*/
DECLARE

    r record ;
    l_has_permission boolean ;
    l_is_active boolean ;
    l_user_id_updated integer ;
    l_desired_action text ;
    l_top_parent_id integer ;
    l_has_duplicate boolean ;

BEGIN

    call util_log.log_begin (
        util_log.dici ( a_id ),
        util_log.dici ( a_parent_id ),
        util_log.dici ( a_name ),
        util_log.dici ( a_is_active ),
        util_log.dici ( a_user ) ) ;

    IF a_id IS NULL THEN
        l_desired_action := 'insert' ;
    ELSE
        l_desired_action := 'update' ;
    END IF ;

    l_has_permission := app_api.can_do (
        a_user => a_user,
        a_action => l_desired_action,
        a_object_type => 'widget',
        a_id => NULL ) ;

    IF NOT l_has_permission THEN
        a_err := 'No can do' ;
        call util_log.log_exception ( a_err ) ;
        RETURN ;
    END IF ;

    l_user_id_updated := app_api.resolve_user_id ( a_user => a_user ) ;

    l_is_active := coalesce ( a_is_active, 'Y' ) = 'Y' ;

    IF a_parent_id IS NOT NULL THEN

        ----------------------------------------------------------------
        -- If the parent ID is not null then ensure that there are no cyclic
        -- references being created.
        -- That is, ensure that the new parent ID is not also a child of the ID
        IF a_id IS NOT NULL THEN

            IF a_id IS NOT DISTINCT FROM a_parent_id THEN
                a_err := 'The widget type cannot have itself for a parent' ;
                call util_log.log_exception ( a_err ) ;
                RETURN ;
            END IF ;

            FOR r IN (
                WITH RECURSIVE toc AS (
                    SELECT base.id,
                            '{}'::integer[] AS parents
                        FROM app_data.rt_widget_type base
                        WHERE base.parent_id IS NULL
                    UNION ALL
                    SELECT base.id,
                            q.parents || base.parent_id
                        FROM app_data.rt_widget_type base
                        JOIN toc q
                            ON ( base.parent_id = q.id
                                AND NOT base.id = ANY ( q.parents ) ) -- avoid any existing cyclic references
                )
                SELECT id
                    FROM toc
                    WHERE toc.id = a_parent_id -- the new parent ID
                        AND a_id = ANY ( toc.parents ) -- the child ID
                )
            LOOP

                a_err := 'No cyclic references allowed' ;
                call util_log.log_exception ( a_err ) ;
                RETURN ;

            END LOOP ;

        END IF ;

    END IF ;

    --------------------------------------------------------------------
    -- Ensure that the name is not a duplicate of any other record that has the same top-level parent ID
    -- IIF a_id is null and a_parent_id is null then we're looking for any record with a null parent ID that has a matching name
    -- IF a_parent_id is not null then determine the top_parent_id for the parent ID
    -- IF a_id is not null then determine the top_parent_id for a_id
    l_has_duplicate := false ;
    l_top_parent_id := NULL ;

    IF a_id IS NULL AND a_parent_id IS NULL THEN

        FOR r IN (
            SELECT id
                FROM widget.rv_widget_type
                WHERE name = a_name
                    AND parent_id IS NULL
                LIMIT 1 )
        LOOP

            l_has_duplicate := true ;

        END LOOP ;

    ELSIF a_parent_id IS NOT NULL THEN

        FOR r IN (
            SELECT top_parent_id
                FROM widget.rv_widget_type
                WHERE name = a_name
                    AND parent_id = a_parent_id
                LIMIT 1 )
        LOOP

            l_top_parent_id := r.top_parent_id ;

        END LOOP ;

    ELSE -- a_id IS NOT NULL

        FOR r IN (
            SELECT top_parent_id
                FROM widget.rv_widget_type
                WHERE name = a_name
                    AND id = a_id
                LIMIT 1 )
        LOOP

            l_top_parent_id := r.top_parent_id ;

        END LOOP ;

    END IF ;

    IF l_top_parent_id IS NOT NULL THEN

        FOR r IN (
            SELECT id
                FROM widget.rv_widget_type
                WHERE name = a_name
                    AND top_parent_id = l_top_parent_id
                    AND top_parent_id IS DISTINCT FROM id
                LIMIT 1 )
        LOOP

            l_has_duplicate := true ;

        END LOOP ;

    END IF ;

    IF l_has_duplicate THEN
        a_err := 'The new name cannot match any other name for the same top parent' ;
        call util_log.log_exception ( a_err ) ;
        RETURN ;
    END IF ;

    --------------------------------------------------------------------
    IF l_desired_action = 'insert' THEN

        INSERT INTO app_data.rt_widget_type (
                id,
                parent_id,
                name,
                is_active,
                created_tmsp,
                updated_tmsp,
                user_id_created,
                user_id_updated )
            SELECT nextval ( 'app_data.seq_rt_id' ) AS id,
                    a_parent_id,
                    a_name,
                    l_is_active,
                    now () AS created_tmsp,
                    now () AS updated_tmsp,
                    l_user_id_updated AS user_id_created,
                    l_user_id_updated AS user_id_updated
                RETURNING id INTO a_id ;

    ELSE

        UPDATE app_data.rt_widget_type
            SET parent_id = a_parent_id,
                name = a_name,
                is_active = l_is_active,
                updated_tmsp = now (),
                user_id_updated = l_user_id_updated
            WHERE id = a_id
                AND ( parent_id IS DISTINCT FROM a_parent_id
                    OR name IS DISTINCT FROM a_name
                    OR is_active IS DISTINCT FROM l_is_active ) ;

    END IF ;

EXCEPTION
    WHEN others THEN
        a_err := substr ( SQLSTATE::text || ' - ' || SQLERRM, 1, 200 ) ;
        call util_log.log_exception ( SQLSTATE::text || ' - ' || SQLERRM ) ;
END ;
$$ ;

ALTER PROCEDURE app_api.upsert_rt_widget_type ( integer, integer, text, text, text, text ) OWNER TO app_owner ;

GRANT EXECUTE ON PROCEDURE app_api.upsert_rt_widget_type ( integer, integer, text, text, text, text ) TO app_updt ;
