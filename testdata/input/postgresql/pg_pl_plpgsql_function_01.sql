-- sqlfmt d:postgres

CREATE FUNCTION decimal_to_dms (
    a_longitude IN numeric,
    a_latitude IN numeric )
RETURN t_coord

STABLE
LANGUAGE plpgsql

AS $$
DECLARE
    l_return t_coord ;

BEGIN

    l_return.longitude := ( to_char ( trunc ( a_longitude ) )
            || ( ( lpad ( trunc ( mod ( abs ( a_longitude ), 1 ) * 60 ), 2, '0' ) )::numeric, '09' )::text
            || ( ( mod ( ( mod ( abs ( a_longitude ), 1 ) * 60 ), 1 ) * 60, '09.9999' )::text ) )::numeric ;

    l_return.latitude := ( substr ( a_latitude, 1, 2 )
            || ( ( lpad ( trunc ( mod ( abs ( a_latitude ), 1 ) * 60 ), 2, '0' ) )::numeric, '09' )::text
            || ( mod ( ( mod ( abs ( a_latitude ), 1 ) * 60 ), 1 ) * 60, '09.9999' )::text )::numeric ;

    RETURN l_return ;
END ;
$$ ;
