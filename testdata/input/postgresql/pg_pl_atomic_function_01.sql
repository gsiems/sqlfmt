-- sqlfmt d:postgres

CREATE OR REPLACE FUNCTION util_coord.is_mn_lat_long (
    a_latitude numeric DEFAULT NULL,
    a_longitude numeric DEFAULT NULL )
RETURNS boolean
LANGUAGE SQL
STABLE
SECURITY INVOKER
BEGIN ATOMIC
/**
Function is_mn_lat_long sanity check a lat/long to ensure that they are
roughly in the state of Minnesota

| Parameter                      | In/Out | Datatype   | Remarks                                            |
| ------------------------------ | ------ | ---------- | -------------------------------------------------- |
| a_latitude                     | in     | numeric    | The latitude of the point                          |
| a_longitude                    | in     | numeric    | The longitude of the point                         |

*/

    SELECT a_latitude IS NOT NULL
            AND a_longitude IS NOT NULL
            AND a_latitude > 43.0
            AND a_latitude < 49.6
            AND a_longitude > -98.0
            AND a_longitude < -89.0 ;

END ;
