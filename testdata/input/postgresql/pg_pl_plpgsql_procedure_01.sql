-- sqlfmt d:postgres

CREATE OR REPLACE PROCEDURE widget.upsert_widget (
    a_id inout integer default null,
    a_owner_id in integer default null,
    a_color in text default null,
    a_user in text default null,
    a_err inout text default null )
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, widget_data
AS $$
/**
Procedure upsert_widget performs an update on dt_widget

| Parameter                      | In/Out | Datatype   | Description                                        |
| ------------------------------ | ------ | ---------- | -------------------------------------------------- |
| a_id                           | in     | integer    | The system generated ID (primary key).             |
| a_owner_id                     | in     | integer    | The ID of the widget                               |
| a_color                        | in     | text       | The current color of the widget                    |
| a_user                         | in     | text       | The ID or username of the user performing the upsert |
| a_err                          | inout  | text       | The (business or database) error that was generated, if any |

*/
DECLARE

    r record ;
    l_has_permission boolean ;
    l_owner_id integer ;
    l_action text ;

BEGIN

    call util_log.log_begin (
        util_log.dici ( a_id ),
        util_log.dici ( a_owner_id ),
        util_log.dici ( a_color ),
        util_log.dici ( a_user ) ) ;

    ----------------------------------------------------------------------------------------------------------
    -- If both a_id and a_owner_id are supplied then ensure that they match
    l_owner_id := a_owner_id ;

    IF a_id IS NOT NULL THEN
        FOR r IN (
            SELECT owner_id
                FROM widget_data.dt_widget
                WHERE id = a_id ) LOOP
            l_owner_id := r.owner_id ;
        END LOOP ;
    END IF ;

    IF a_id IS NULL THEN
        l_action := 'insert' ;
    ELSE
        l_action := 'update' ;
    END IF ;

    l_has_permission := widget.can_do (
        a_user => a_user,
        a_action => l_action,
        a_object_type => 'widget',
        a_id => a_id,
        a_parent_id => l_owner_id,
        a_parent_object_type => 'owner' ) ;

    IF NOT l_has_permission THEN
        a_err := 'Insufficient privileges or the owner does not exist' ;
        call util_log.log_exception ( a_err ) ;
        RETURN ;
    END IF ;

    call widget.priv_upsert_widget (
        a_id => a_id,
        a_owner_id => l_owner_id,
        a_color => a_color,
        a_comments => a_comments,
        a_user => a_user,
        a_err => a_err ) ;

EXCEPTION
    WHEN others THEN
        a_err := substr ( SQLSTATE::text || ' - ' || SQLERRM, 1, 200 ) ;
        call util_log.log_exception ( SQLSTATE::text || ' - ' || SQLERRM ) ;
END ;
$$ ;
