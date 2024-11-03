-- sqlfmt d:sqlite

/*
References:
https://www.sqlite.org/lang_select.html
*/

SELECT args.table_catalog,
        args.table_schema,
        m.name AS table_name,
        NULL AS table_owner,
        upper ( m.type ) AS table_type,
        NULL AS row_count,
        NULL AS comment,
        CASE
            WHEN m.type = 'view' THEN m.sql
            END AS view_definition
    FROM sqlite_master m
    CROSS JOIN (
        SELECT file AS table_catalog,
                coalesce ( $1, '' ) AS table_schema
FROM pragma_database_list WHERE seq = 0

        ) AS args
    WHERE m.type IN ( 'table', 'view' )
        AND substr ( m.name, 1, 7 ) <>  'sqlite_' ;


SELECT m.name AS table_name,
        m.sql
    FROM sqlite_master m
    CROSS JOIN (
        SELECT coalesce ( $1, '' ) AS table_name
        ) AS args
    WHERE m.type = 'table'
        AND substr ( m.tbl_name, 1, 7 ) <>  'sqlite'
        AND ( args.table_name = '' OR args.table_name = m.name ) ;


SELECT args.table_catalog,
        args.table_schema,
        m.name AS table_name,
        cols.name AS column_name,
        cols.cid AS ordinal_position,
        cols."type" AS data_type,
        CASE
            WHEN cols."notnull" = 1 THEN 'NO'
            ELSE 'YES'
            END AS is_nullable,
        NULL AS column_default,
        NULL AS domain_catalog,
        NULL AS domain_schema,
        NULL AS domain_name,
        NULL AS comment
    FROM sqlite_master AS m
    JOIN pragma_table_info ( m.name ) AS cols
    CROSS JOIN (
        SELECT file AS table_catalog,
                coalesce ( $1, '' ) AS table_schema,
                coalesce ( $2, '' ) AS table_name
FROM pragma_database_list WHERE seq = 0
        ) AS args
    WHERE m.type IN ( 'table', 'view' )
        AND substr ( m.tbl_name, 1, 7 ) <>  'sqlite_'
        AND ( args.table_name = '' OR args.table_name = m.name ) ;

SELECT con.table_catalog AS index_catalog,
        con.table_schema AS index_schema,
        con.index_name,
        '' AS index_type,
        group_concat ( con.column_name, ', ' ) AS index_columns,
        con.table_catalog,
        con.table_schema,
        con.table_name,
        CASE
            WHEN max ( con.origin ) IN ( 'pk', 'u' ) THEN 'YES'
            ELSE 'NO'
            END AS is_unique,
        -- status
        '' AS comments
    FROM (
        SELECT args.table_catalog,
                tab.name AS table_name,
                args.table_schema,
                idx.name AS index_name,
                idx."unique",
                idx.origin,
                idx.partial,
                col.name AS column_name,
                col.seqno AS ordinal_position
            FROM sqlite_master AS tab
            CROSS JOIN (
                SELECT file AS table_catalog,
                        coalesce ( $1, '' ) AS table_schema,
                        coalesce ( $2, '' ) AS table_name
FROM pragma_database_list WHERE seq = 0
                  ) AS args
            JOIN pragma_index_list ( tab.name ) AS idx
            JOIN pragma_index_info ( idx.name ) AS col
            WHERE tab.type = 'table'
                AND substr ( tab.name, 1, 7 ) <>  'sqlite_'
                AND ( args.table_name = '' OR args.table_name = tab.name )
            ORDER BY tab.name,
                idx.name,
                col.seqno
        ) AS con
    GROUP BY con.table_schema,
        con.table_name,
        con.index_name ;

SELECT pk_col.table_catalog,
        pk_col.table_schema,
        pk_col.table_name,
        'pk_' || pk_col.table_name AS constraint_name,
        group_concat ( pk_col.column_name, ', ' ) AS constraint_columns,
        'Enabled' AS status,
        '' AS comments
    FROM (
        SELECT args.table_catalog,
                m.name as table_name,
                args.table_schema,
                col.name AS column_name,
                col.pk AS ordinal_position
            FROM sqlite_master AS m
            JOIN pragma_table_info ( m.name ) AS col
            CROSS JOIN (
                SELECT file AS table_catalog,
                        coalesce ( $1, '' ) AS table_schema,
                        coalesce ( $2, '' ) AS table_name
FROM pragma_database_list WHERE seq = 0
                  ) AS args
            WHERE m.type = 'table'
                AND substr ( m.tbl_name, 1, 7 ) <>  'sqlite_'
                AND ( args.table_name = '' OR args.table_name = m.name )
                AND col.pk > 0
            ORDER BY m.name,
                col.pk
        ) AS pk_col
    GROUP BY pk_col.table_schema,
        pk_col.table_name ;

SELECT args.table_catalog,
        args.table_schema,
        con.table_name,
        con.column_names,
        idx_fk.index_name AS constraint_name,
        args.table_catalog AS ref_table_catalog,
        args.table_schema AS ref_table_schema,
        con.ref_table_name,
        con.ref_column_names,
        idx_uniq.index_name AS ref_constraint_name,
        con.match_option,
        con.update_rule,
        con.delete_rule,
        'YES' AS is_enforced,
        --is_deferrable,
        --initially_deferred,
        '' AS comments
    FROM (
        SELECT fk_col.table_name,
                group_concat ( fk_col.column_name, ', ' ) AS column_names,
                --fk_col.table_name || '_fk' || fk_col.id AS constraint_name,
                fk_col.ref_table_name,
                group_concat ( fk_col.ref_column_name, ', ' ) AS ref_column_names,
                fk_col.match_option,
                fk_col.update_rule,
                fk_col.delete_rule
            FROM (
                SELECT tab.name AS table_name,
                        con.id,
                        con."table" AS ref_table_name,
                        con."from" AS column_name,
                        con."to" AS ref_column_name,
                        con.seq AS ordinal_position,
                        con."match" AS match_option,
                        con.on_update AS update_rule,
                        con.on_delete AS delete_rule
                    FROM sqlite_master AS tab
                    JOIN pragma_foreign_key_list ( tab.name ) AS con
                    WHERE tab.type IN ( 'table' )
                        AND substr ( tab.name, 1, 7 ) <>  'sqlite_'
                    ORDER BY tab.name,
                        con.id,
                        con.seq
                ) AS fk_col
            GROUP BY fk_col.table_name,
                fk_col.ref_table_name,
                fk_col.id,
                fk_col.match_option,
                fk_col.update_rule,
                fk_col.delete_rule
        ) AS con
    CROSS JOIN (
        SELECT file AS table_catalog,
                coalesce ( $1, '' ) AS table_schema,
                coalesce ( $2, '' ) AS table_name
FROM pragma_database_list WHERE seq = 0
          ) AS args
    LEFT JOIN (
        SELECT idx_col.table_name,
                idx_col.index_name,
                group_concat ( idx_col.column_name, ', ' ) AS column_names
            FROM (
                SELECT tab.tbl_name AS table_name,
                        tab.name AS index_name,
                        col.name AS column_name,
                        col.seqno AS ordinal_position
                    FROM sqlite_master AS tab
                    JOIN pragma_index_info ( tab.name ) AS col
                    WHERE tab.type IN ( 'index' )
                        AND substr ( tab.name, 1, 7 ) <>  'sqlite_'
                    ORDER BY tab.tbl_name,
                        tab.name,
                        col.seqno
                ) AS idx_col
            GROUP BY idx_col.table_name,
                idx_col.index_name
        ) AS idx_fk
        ON ( con.table_name = idx_fk.table_name
            AND con.column_names = idx_fk.column_names )
    LEFT JOIN (
        SELECT table_name,
                index_name,
                group_concat ( column_name, ', ' ) AS column_names
            FROM (
                SELECT m.name AS table_name,
                        con.name AS index_name,
                        col.name AS column_name,
                        col.seqno AS ordinal_position
                    FROM sqlite_master AS m
                    JOIN pragma_index_list ( m.name ) AS con
                    JOIN pragma_index_info ( con.name ) AS col
                    WHERE con."unique" = 1
                ) AS idx_col
            GROUP BY idx_col.table_name,
                idx_col.index_name
        ) AS idx_uniq
        ON ( con.table_name = idx_uniq.table_name
            AND con.column_names = idx_uniq.column_names )
    WHERE ( con.table_name = args.table_name OR args.table_name = '' )
        OR ( con.ref_table_name = args.table_name OR args.table_name = '' ) ;
