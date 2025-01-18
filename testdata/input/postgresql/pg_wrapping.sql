-- sqlfmt d:postgres

SELECT col_00,
        CASE
            WHEN expression_1 THEN 1
            WHEN expression_2 THEN 2
            WHEN expression_3 THEN 3
            ELSE 4
            END AS col_01,
        concat_ws ( ':',
            CASE
                WHEN expression_4 THEN '5'
                WHEN expression_5 THEN '6'
                END,
            CASE
                WHEN expression_6 THEN '7'
                ELSE '8'
                END ) AS col_02,
        CASE expression_7
            WHEN a_7 THEN 9
            WHEN b_7 THEN 10
            WHEN c_7 THEN
                CASE
                    WHEN expression_8 THEN 11
                    ELSE 12
                    END
            ELSE 13
            END AS col_03,
        CASE expression_9
            WHEN a_9 THEN 14
            WHEN b_9 THEN 15
            WHEN c_9 THEN 16
            ELSE CASE WHEN expression_10 THEN 17 ELSE 18 END
            END AS col_04,
        CASE -- a comment
            WHEN expression_11 THEN 19
            WHEN expression_12 THEN 20 -- another comment
            WHEN expression_13 THEN 21
            ELSE 22
            END AS col_05;


SELECT col_00, CASE WHEN expression_1 THEN 1 WHEN expression_2 THEN 2 WHEN
expression_3 THEN 3 ELSE 4 END AS col_01, concat_ws ( ':', CASE WHEN
expression_4 THEN '5' WHEN expression_5 THEN '6' END, CASE WHEN expression_6
THEN '7' ELSE '8' END ) AS col_02, CASE expression_7 WHEN a_7 THEN 9 WHEN b_7
THEN 10 WHEN c_7 THEN CASE WHEN expression_8 THEN 11 ELSE 12 END ELSE 13 END AS
col_03, CASE expression_9 WHEN a_9 THEN 14 WHEN b_9 THEN 15 WHEN c_9 THEN 16
ELSE CASE WHEN expression_10 THEN 17 ELSE 18 END END AS col_04, CASE WHEN
expression_11 THEN 19 WHEN expression_12 THEN 20 WHEN expression_13 THEN 21
ELSE 22 END AS col_05;


SELECT n.id,
        n.name,
        CASE WHEN n.is_active IS NULL THEN 'N' ELSE coalesce ( n.is_active, o.is_active ) END AS is_active,
        CASE WHEN n.is_active IS NULL THEN 'N' ELSE coalesce ( n.is_active, o.is_active ) END AS is_also_active
    FROM n ;

with n as (
select concat_ws ( ',',
 CASE WHEN n.has_attrib_01       THEN 'attrib_01' END,
 CASE WHEN n.has_attrib_02       THEN 'attrib_02' END,
 CASE WHEN n.has_attrib_03       THEN 'attrib_03' END,
 CASE WHEN n.has_attrib_04       THEN 'attrib_04' END,
 CASE WHEN n.has_attrib_05       THEN 'attrib_05' END,
 CASE WHEN n.has_attrib_06       THEN 'attrib_06' END,
 CASE WHEN n.has_attrib_07       THEN 'attrib_07' END ) AS attributs,


)

select
    n.precision + CASE
        WHEN coalesce ( n.min_value, 0 ) < 0 THEN 1
        ELSE 0
        END + CASE WHEN coalesce ( n.scale, 0 ) > 0 THEN 1 ELSE 0 END AS max_char_length,

    trim ( cast (
            CASE
        WHEN n.legacy_data_code
                    = '000' THEN '42'
        WHEN n.legacy_data_code
                    = '001' THEN '43'
        ELSE n.legacy_data_code
        END AS text ) ) AS column_042,
        CASE
            WHEN func_02 ( n.something_something_01 ) IS NOT NULL THEN func_02 ( n.something_something_01 )
            WHEN func_02 ( n.something_something_02 ) IS NOT NULL AND upper ( trim ( n.something_something_02 ) )
                <> upper ( trim ( n.something_something_03 ) )
                THEN trim ( n.something_something_04 )
            END AS column_043,

        coalesce ( n.parameter_01_cnt, 0 )
            + coalesce ( n.parameter_02_cnt, 0 )
            + coalesce ( n.parameter_03_cnt, 0 )
            + coalesce ( n.parameter_04_cnt, 0 ) AS param_cnts,
        CASE
            WHEN coalesce ( n.parameter_01_cnt, 0 ) > 0
                OR coalesce ( n.parameter_02_cnt, 0 ) > 0
                OR coalesce ( n.parameter_03_cnt, 0 ) > 0
                OR coalesce ( n.parameter_04_cnt, 0 ) > 0
                THEN true
            ELSE false
            END AS has_params,


        sum ( CASE WHEN n.name = '' AND parameter_03_cnt IS NOT NULL THEN parameter_03_cnt END ) AS cnt_03,
        sum ( CASE WHEN n.name = 'qwer' AND parameter_04_cnt IS NOT NULL THEN parameter_04_cnt END ) AS cnt_04,
        sum ( CASE WHEN n.name = 'qwertyui' AND parameter_05_cnt IS NOT NULL THEN parameter_05_cnt END ) AS cnt_05,
        sum ( CASE WHEN n.name = 'qwertyuiopas' AND parameter_06_cnt IS NOT NULL THEN parameter_06_cnt END ) AS cnt_06,
        sum ( CASE WHEN n.name = 'qwertyuiopasdfgh' AND parameter_07_cnt IS NOT NULL THEN parameter_07_cnt END ) AS cnt_07,
        sum ( CASE WHEN n.name = 'qwertyuiopasdfghjklz' AND parameter_08_cnt IS NOT NULL THEN parameter_08_cnt END ) AS cnt_08,
        CASE WHEN n.name = 'qwertyuiopasdfghjklzxcvb' AND n.some_parameter_09 THEN 'Yes' ELSE 'No' END  AS some_column_09,


        CASE
            WHEN n.order_date IS NOT NULL
                THEN ( concat_ws ( ' ', to_char ( n.order_date, 'yyyy-mm-dd' ), coalesce ( to_char ( n.order_time, 'hh24:mi:ss' ), '00:00:00' ) ) )::timestamp without time zone
            END AS column_09,
        CASE
            WHEN coalesce ( n.code, '' ) <> '' THEN n.resolve_code ( a_code => a_code )
            WHEN coalesce ( a_some_param_name, '' ) <> '' THEN n.resolve_param ( a_param => a_some_param_name )
            END AS column_10,
        concat_ws ( ',',
 CASE WHEN n.has_attrib_01       THEN 'attrib_01' END,
 CASE WHEN n.has_attrib_02       THEN 'attrib_02' END,
 CASE WHEN n.has_attrib_03       THEN 'attrib_03' END,
 CASE WHEN n.has_attrib_04       THEN 'attrib_04' END,
 CASE WHEN n.has_attrib_05       THEN 'attrib_05' END,
 CASE WHEN n.has_attrib_06       THEN 'attrib_06' END,
 CASE WHEN n.has_attrib_07       THEN 'attrib_07' END ) AS attributs,
        case when n.name = '' then o.name::text || ' ' || o.col01::text else  p.name::text || ' ' || p.col01::text end as some_thing1,
        case when n.name = 'qw' then o.name::text || ' ' || o.col02::text else  p.name::text || ' ' || p.col02::text end as some_thing2,
        case when n.name = 'qwe' then o.name::text || ' ' || o.col03::text else  p.name::text || ' ' || p.col03::text end as some_thing3,
        case when n.name = 'qwertyu' then o.name::text || ' ' || o.col04::text else  p.name::text || ' ' || p.col04::text end as some_thing4,
2
;

SELECT column_01 IS NOT NULL
            AND column_02 IS NOT NULL
            AND column_01 > 10.0
            AND column_01 < 20.0
            AND column_02 > 100.0
            AND column_02 < 200.0 ;

SELECT column_01 IS NOT NULL
            AND column_02 IS NOT NULL
            AND ( ( column_01 > 10.0
                    AND column_01 < 20.0 )
                OR ( column_02 > 100.0
                    AND column_02 < 200.0 ) ) ;

SELECT coalesce (
            func_01 ( 'foo', 'bar', 42 ),
            func_02 ( 'foo', 'bar', 42 ),
            func_03 ( 'foo', 'bar', 42 ),
            func_04 ( 'foo', 'bar', 42 ),
            func_05 ( 'foo', 'bar', 42 ),
            func_06 ( 'foo', 'bar', 42 ) ) ;

SELECT coalesce (func_01('foo','bar',42),func_02('foo','bar',42),func_03('foo','bar',42),func_04('foo','bar',42),func_05('foo','bar',42),func_06('foo','bar',42));

/*********************************/

SELECT func_01 (
            param_1 => 1,
            param_2 => func_02 ( param_021 => 1 ),
            param_3 => func_03 ( param_031 => 1, param_032 => 2 ),
            param_4 => func_04 ( 'param' ),
            param_5 => 5 ) ;

SELECT func_01(param_1=>1,param_2=>func_02(param_021=>1),param_3=>func_03(param_031=>1,param_032=>2),param_4=>func_04('param'),param_5=>5);

/*********************************/

SELECT func_01 (
            param_1 => 1,
            param_2 => func_02 ( param_10 => 1, param_11 => 2 ),
            param_3 => func_03 (
                    param_20 => 1,
                    param_21 => 2,
                    param_22 => 3,
                    param_23 => 4 ),
            param_4 => 42 ) ;

SELECT func_01(param_1=>1,param_2=>func_02(param_10=>1,param_11=>2),param_3=>func_03(param_20=>1,param_21=>2,param_22=>3,param_23=>4),param_4=>42);

/*********************************/

select foo
    where bar in ( 12300, 12301, 12302, 12303, 12304, 12305, 12306, 12307,
        12308, 12309, 12310, 12311, 12312, 12313, 12314, 12315, 12316, 12317,
        12318, 12319, 12320, 12321, 12322, 12323, 12324, 12325, 12326, 12327,
        12328, 12329, 12330, 12331, 12332, 12333, 12334, 12335, 12336, 12337,
        12338, 12339, 12340, 12341, 12342, 12343, 12344, 12345, 12346, 12347,
        12348, 12349, 12350, 12351, 12352, 12353, 12354, 12355, 12356, 12357,
        12358, 12359, 12360, 12361, 12362, 12363, 12364, 12365, 12366, 12367,
        12368, 12369, 12370, 12371, 12372, 12373, 12374, 12375, 12376, 12377,
        12378, 12379, 12380, 12381, 12382, 12383 ) ;

select foo
    where bar in (12300,12301,12302,12303,12304,12305,12306,12307,12308,12309,12310,12311,12312,12313,12314,12315,12316,12317,12318,12319,12320,12321,12322,12323,12324,12325,12326,12327,12328,12329,12330,12331,12332,12333,12334,12335,12336,12337,12338,12339,12340,12341,12342,12343,12344,12345,12346,12347,12348,12349,12350,12351,12352,12353,12354,12355,12356,12357,12358,12359,12360,12361,12362,12363,12364,12365,12366,12367,12368,12369,12370,12371,12372,12373,12374,12375,12376,12377,12378,12379,12380,12381,12382,12383);

/*********************************/

SELECT 'Marley''s Ghost' AS heading,
        'Marley was dead: to begin with. There is no doubt whatever about that. The'
            || ' register of his burial was signed by the clergyman, the clerk, the undertaker,'
            || ' and the chief mourner. Scrooge signed it: and Scrooge’s name was good upon'
            || ' ’Change, for anything he chose to put his hand to. Old Marley was as dead as a'
            || ' door-nail.' AS para01 ;

select 'Marley''s Ghost' as heading,
        'Marley was dead: to begin with. There is no doubt whatever about that. The' || ' register of his burial was signed by the clergyman, the clerk, the undertaker,' || ' and the chief mourner. Scrooge signed it: and Scrooge’s name was good upon' || ' ’Change, for anything he chose to put his hand to. Old Marley was as dead as a' || ' door-nail.' as para01 ;

select
'UPDATE ' || l_table_name ||
                          ' SET id = id + ' || l_id || ','
                          ' other_id = ' || l_id_count ||
                          ' WHERE CURRENT OF ' || quote_ident(l_cursor::TEXT)
;


CREATE OR REPLACE PROCEDURE src_schema.foo (
    a_id in integer default null )
LANGUAGE plpgsql
AS $proc$

BEGIN


    call util_log.log_begin (
        util_log.dici ( a_param_01 ),
        util_log.dici ( a_param_02 ),
        --util_log.dici ( a_param_03 ),
        util_log.dici ( a_param_04 ),
        util_log.dici ( a_param_05 ),
        util_log.dici ( a_param_06 ),
        util_log.dici ( a_param_07 ) ) ;

    ------------------------------------------------------------------------
    -- just barely long enough to wrap on commas
    l_var_01 := concat_ws ( l_new_line, l_result, '', 'blah, blah, blah, blah, blah, b ' || a_parameter_name || ' ;' ) ;
    l_var_02 := concat_ws ( l_new_line, l_result, '', 'blah, blah, blah, blah, blah, bl ' || a_parameter_name || ' ;' ) ;
    l_var_03 := concat_ws ( l_new_line, l_result, '', 'blah, blah, blah, blah, blah, bla ' || a_parameter_name || ' ;' ) ;

    ------------------------------------------------------------------------
    -- more than long enough to wrap to wrap on both commas and concatenation operators
    l_var_04 := concat_ws ( l_new_line, l_result, '', 'blah, blah, blah, blah, blah ' || a_parameter_name || ' ; -- some final comment text chucked in a literal to make things too long' ) ;

    ------------------------------------------------------------------------
    -- more than long enough to wrap on commas but not long enough to wrap on concatenation operators
    l_var_05 := concat_ws (
        l_new_line,
        l_result,
        '',
        'blah, blah ' || a_parameter_name || ' blah, blah, blah ' || a_parameter_name || ' ;' ) ;

    ------------------------------------------------------------------------
    -- too short to wrap
    l_var_06 := concat_ws ( l_new_line, l_result, 'blah, blah ' || a_parameter_name ) ;

END ;
$proc$ ;
