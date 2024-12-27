-- sqlfmt d:postgres

CREATE OR REPLACE FUNCTION app_json_api.get_stuff (
    a_id in integer DEFAULT NULL,
    a_name in text DEFAULT NULL,
    a_user in text DEFAULT NULL )
RETURNS text
LANGUAGE SQL
STABLE

AS $$

WITH t AS (
    SELECT t0.id,
            t0.name,
            t0.stuff_type,
            t0.full_name,
            t0.is_active
        FROM app_api.get_stuff (
                a_id => a_id,
                a_name => a_name,
                a_user => a_user ) t0
),
cte_01 AS (
    SELECT stuff_id,
            json_agg ( json_build_object (
                    'moreStuffId', more_stuff_id,
                    'moreStuff', more_stuff,
                    'isActive', is_active ) ) AS ja
        FROM (
            SELECT alia_01.stuff_id,
                    alia_01.more_stuff_id,
                    alia_01.more_stuff,
                    alia_01.is_active
                FROM app_api.dv_more_stuff alia_01
                JOIN t
                    ON ( t.id = alia_01.stuff_id )
                WHERE alia_01.is_active = 'Y'
                ORDER BY alia_01.more_stuff
            ) x
        GROUP BY stuff_id
),
cte_02 AS (
    SELECT t.id,
            t.name,
            t.stuff_type AS "stuffType",
            t.full_name AS "fullName",
            t.is_active AS "isActive",
            cte_01.ja AS "moreStuffs"
        FROM t
        LEFT JOIN cte_01
            ON ( cte_01.stuff_id = t.id )
)
SELECT row_to_json ( cte_02 ) AS json
    FROM cte_02 ;

$$ ;
