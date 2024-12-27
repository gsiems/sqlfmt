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
