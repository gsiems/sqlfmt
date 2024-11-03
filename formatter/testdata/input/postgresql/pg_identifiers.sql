-- sqlfmt dialect: PostgreSQL

-- Identifiers and single-quoted strings
SELECT 'some text' AS "COL1",
        '''some text''' AS "Col2",
        'it''s more text' AS Col3,
        'some ûñìçóde text' AS "col4",
        E'and\nfinally\nmore\ntext' as col5,
        'ñ' as "ñ" ;
