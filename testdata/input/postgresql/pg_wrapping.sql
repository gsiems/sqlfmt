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

SELECT func_01 (
            param_1 => 1,
            param_2 => func_02 ( param_10 => 1, param_11 => 2 ),
            param_3 => func_03 (
                    param_20 => 1,
                    param_21 => 2,
                    param_22 => 3,
                    param_23 => 4 ),
            param_4 => 42 ) ;

select foo
    where bar in ( 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345, 12345 ) ;
