-- sqlfmt d:postgres

CREATE OR REPLACE FUNCTION widget.find_widget (
    a_user text,
    a_search_term text )
RETURNS SETOF widget.dv_widget
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, widget
LANGUAGE plpgsql
AS $$
/**
Function find_widget takes a user and search term and returns the list of
widgets that match the search term and that the user has privileges to view.

| Parameter                  | In/Out | Datatype | Remarks                                          |
| -------------------------- | ------ | -------- | ------------------------------------------------ |
| a_user                     | IN     | text     | The ID or username of the user doing the search  |
| a_search_term              | IN     | text     | The string to search for                         |

*/
DECLARE

    l_has_permission boolean ;

BEGIN

    l_has_permission := widget.can_do (
        a_user => a_user,
        a_action => 'select',
        a_object_type => 'widget',
        a_id => null ) ;

    RETURN QUERY
        WITH base AS (
            SELECT id,
                    model_number,
                    widget_color,
                    widget_status,
                    widget_location
                FROM widget.dv_widget
                WHERE l_has_permission
        ),
        mtch AS (
            SELECT id
                FROM base
                WHERE ( ( a_search_term IS NOT NULL
                            AND trim ( a_search_term ) <> ''
                            AND lower ( base::text ) ~ lower ( a_search_term ) )
                        OR ( trim ( coalesce ( a_search_term, '' ) ) = '' ) )
        )
        SELECT dw.*
            FROM widget.dv_widget dw
            JOIN mtch
                ON ( mtch.id = dw.id ) ;

END ;
$$ ;

ALTER FUNCTION widget.find_widget ( text, text ) OWNER TO app_owner ;
