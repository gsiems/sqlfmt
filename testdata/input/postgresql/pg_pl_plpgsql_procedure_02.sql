-- sqlfmt d:postgres

CREATE OR REPLACE FUNCTION util_meta.mk_object_migration (
    a_object_schema text DEFAULT NULL,
    a_object_name text DEFAULT NULL )
RETURNS text
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, util_meta
AS $$
/**
Function mk_object_migration generates a script for migrating the structure of a table, or any other database object

| Parameter                      | In/Out | Datatype   | Description                                        |
| ------------------------------ | ------ | ---------- | -------------------------------------------------- |
| a_object_schema                | in     | text       | The (name of the) schema that contains the object to migrate |
| a_object_name                  | in     | text       | The (name of the) object to create a migration script for |

*/
DECLARE

    r record ;
    r2 record ;

    l_columns text[] ;
    l_create_cmd text ;
    l_drop_cmd text ;
    l_found_dep boolean ;
    l_full_object_name text ;
    l_full_seq_name text ;
    l_grants text[] ;
    l_new_line text ;
    l_object_oid oid ;
    l_object_type text ;
    l_pk_cols text[] ;
    l_result text ;
    l_seq_name text ;
    l_seq_stmt text ;
    l_set_sequences text[] ;

BEGIN

    l_new_line := util_meta.new_line () ;

    ----------------------------------------------------------------------------
    -- Ensure that the specified object is valid
    FOR r IN (
        SELECT object_oid,
                object_type,
                full_object_name,
                '\i ' || concat_ws ( '/', directory_name, file_name ) AS create_cmd
            FROM util_meta.objects
            WHERE schema_name = a_object_schema
                AND object_name = a_object_name
            LIMIT 1 )
    LOOP

        l_object_oid := r.object_oid ;
        l_object_type := r.object_type ;
        l_full_object_name := r.full_object_name ;
        l_create_cmd := r.create_cmd ;

        IF l_object_type <> 'table' THEN

            FOR r2 IN (
                SELECT array_agg ( concat_ws ( ' ', 'DROP', object_type, full_object_name, '(' || calling_signature || ')', ';' ) ) AS drop_cmds
                    FROM util_meta.objects
                    WHERE schema_name = a_object_schema
                        AND object_name = a_object_name )
            LOOP

                l_drop_cmd := array_to_string ( r2.drop_cmds, l_new_line ) ;

            END LOOP ;

        END IF ;

    END LOOP ;

    IF l_object_type IS NULL THEN
        RETURN 'ERROR: invalid object' ;
    END IF ;

    -----------------------------------------------------------------------------
    -- For tables the intent is to move the existing table, create a new table and
    -- then use the old table to populate the new.
    -- For non-tables the intent is to simply drop the old object and re-create it
    IF l_object_type = 'table' THEN

        ------------------------------------------------------------------------
        -- Ensure that a backup table does not exist
        IF util_meta.is_valid_object ( 'bak_' || a_object_schema, a_object_name, 'table' ) THEN
            RETURN 'ERROR: a backup table already exists' ;
        END IF ;

        ------------------------------------------------------------------------
        -- Drop any foreign key relationships against the specified table
        FOR r IN (
            SELECT DISTINCT 'ALTER TABLE ' || full_table_name || ' DROP CONSTRAINT ' || constraint_name || ' ;' AS cmd
                FROM util_meta.foreign_keys
                WHERE ref_schema_name = a_object_schema
                    AND ref_table_name = a_object_name )
        LOOP

            l_result := concat_ws ( l_new_line, l_result, r.cmd ) ;

        END LOOP ;

        ------------------------------------------------------------------------
        -- Ensure that there is a backup schema to move the existing table to
        l_result := concat_ws ( l_new_line, l_result, '', 'CREATE SCHEMA IF NOT EXISTS bak_' || a_object_schema || ' ;' ) ;

        ------------------------------------------------------------------------
        -- Move the existing table
        l_result := concat_ws ( l_new_line, l_result, '', 'ALTER TABLE ' || l_full_object_name || ' SET SCHEMA bak_' || a_object_schema || ' ;' ) ;

        ------------------------------------------------------------------------
        -- Execute the table creation DDL file
        l_result := concat_ws ( l_new_line, l_result, '', l_create_cmd ) ;

        ------------------------------------------------------------------------
        -- Copy the data from the backup to the new table
        FOR r IN (
            SELECT column_name,
                    is_pk,
                    column_default
                FROM util_meta.columns
                WHERE schema_name = a_object_schema
                    AND object_name = a_object_name
                ORDER BY ordinal_position )
        LOOP

            l_columns := array_append ( l_columns, r.column_name ) ;

            IF r.column_default ~ '^nextval' THEN

                l_seq_name := split_part ( r.column_default, '''', 2 ) ;
                -- In some situations the sequence name is already fully qualified, in other situations it is not...
                -- Ensure that the sequence name is fully qualified.
                IF l_seq_name like a_object_schema || '.%' THEN
                    l_full_seq_name := l_seq_name ;
                ELSE
                    l_full_seq_name := a_object_schema || '.' || l_seq_name ;
                END IF ;

                l_seq_stmt := concat_ws ( l_new_line, 'WITH cv AS (', util_meta.indent ( 1 ) || 'SELECT 1 AS rn,', util_meta.indent ( 3 ) || 'last_value', util_meta.indent ( 2 ) || 'FROM ' || l_full_seq_name, '),', 'mv AS (', util_meta.indent ( 1 ) || 'SELECT 1 AS rn,', util_meta.indent ( 3 ) || 'max ( ' || r.column_name || ' ) AS max_value', util_meta.indent ( 2 ) || 'FROM ' || l_full_object_name, ')', 'SELECT pg_catalog.setval ( ' || quote_literal ( l_full_seq_name ) || ', mv.max_value, true )', util_meta.indent ( 1 ) || 'FROM mv', util_meta.indent ( 1 ) || 'JOIN cv', util_meta.indent ( 2 ) || 'ON ( cv.rn = mv.rn )', util_meta.indent ( 1 ) || 'WHERE mv.max_value > cv.last_value ;' ) ;

                l_set_sequences := array_append ( l_set_sequences, l_seq_stmt ) ;

            END IF ;

            IF r.is_pk THEN
                l_pk_cols := array_append ( l_pk_cols, r.column_name ) ;
            END IF ;

        END LOOP ;

        l_result := concat_ws ( l_new_line, l_result, '', 'INSERT INTO ' || l_full_object_name || ' (', util_meta.indent ( 3 ) || array_to_string ( l_columns, ',' || l_new_line || util_meta.indent ( 3 ) ) || ' )', util_meta.indent ( 1 ) || 'SELECT ' || array_to_string ( l_columns, ',' || l_new_line || util_meta.indent ( 3 ) ), util_meta.indent ( 2 ) || 'FROM bak_' || a_object_schema || '.' || a_object_name, util_meta.indent ( 2 ) || concat_ws ( ' ', 'ORDER BY ' || array_to_string ( l_pk_cols, ',' || l_new_line || util_meta.indent ( 3 ) ), ';' ) ) ;

        ------------------------------------------------------------------------
        -- VACUUM ANALYZE
        l_result := concat_ws ( l_new_line, l_result, '', 'VACUUM ANALYZE ' || l_full_object_name || ' ;' ) ;

        ------------------------------------------------------------------------
        -- Reset sequences
        IF array_length ( l_set_sequences, 1 ) > 0 THEN
            l_result := concat_ws ( l_new_line, l_result, '', array_to_string ( l_set_sequences, l_new_line ) ) ;
        END IF ;

        ------------------------------------------------------------------------
        -- Re-create the foreign keys against the re-built table
        FOR r IN (
            WITH fk AS (
                SELECT DISTINCT full_table_name,
                        constraint_name,
                        column_names,
                        ref_full_table_name,
                        ref_column_names,
                        CASE
                            WHEN update_rule <> 'NO ACTION' THEN ' ON UPDATE ' || update_rule
                            END AS on_update,
                        CASE
                            WHEN delete_rule <> 'NO ACTION' THEN ' ON DELETE ' || delete_rule
                            END AS on_delete
                    FROM util_meta.foreign_keys
                    WHERE ref_schema_name = a_object_schema
                        AND ref_table_name = a_object_name
            )
            SELECT concat_ws ( ' ', 'ALTER TABLE', full_table_name, 'ADD CONSTRAINT', constraint_name, 'FOREIGN KEY (', column_names, ') REFERENCES (', ref_column_names, ')', on_update, on_delete, ' ;' ) AS cmd
                FROM fk
                ORDER BY constraint_name )
        LOOP

            l_result := concat_ws ( l_new_line, l_result, '', r.cmd ) ;

        END LOOP ;

    END IF ;

    ----------------------------------------------------------------------------
    ----------------------------------------------------------------------------
    -- Drop and re-create any objects that have a known (to the database) dependency
    -- on the object being migrated.
    -- If done properly then it should be possible to drop the backed up table when
    -- finished without having to deal with errors due to dependent objects.
    l_result := concat_ws ( l_new_line, l_result, '', concat_ws ( ' ', '-- Re-create objects that depend on', l_full_object_name, 'oid:', l_object_oid::text ) ) ;

EXECUTE    'create temporary table temp_obj_deps AS
    SELECT 0::int AS tree_depth,
            object_oid,
            schema_name,
            object_name,
            full_object_name,
            object_type,
            directory_name,
            file_name,
            calling_signature
        FROM util_meta.objects
        WHERE object_oid = ' || l_object_oid::text ;

create temporary table temp_all_deps AS
    SELECT object_oid,
            dep_object_oid
        FROM util_meta.dependencies
        WHERE object_oid <> dep_object_oid ;

    l_found_dep := true ;
    while l_found_dep
    LOOP
        l_found_dep := false ;

        FOR r IN (
            WITH mx AS (
                SELECT max ( tree_depth ) + 1 AS tree_depth
                    FROM temp_obj_deps
            )
            SELECT DISTINCT mx.tree_depth,
                    obj.object_oid,
                    obj.schema_name,
                    obj.object_name,
                    obj.full_object_name,
                    obj.object_type,
                    obj.directory_name,
                    obj.file_name,
                    obj.calling_signature
                FROM temp_all_deps dep
                CROSS JOIN mx
                JOIN util_meta.objects obj
                    ON ( obj.object_oid = dep.dep_object_oid )
                JOIN temp_obj_deps ttd
                    ON ( ttd.object_oid = dep.object_oid )
                LEFT JOIN temp_obj_deps ttd2
                    ON ( ttd2.object_oid = obj.object_oid )
                WHERE ttd2.object_oid IS NULL )
        LOOP

            -- Guard against a runaway loop. Shouldn't happen (unless the query is changed)
            -- If the depth is anywhere near 20 then there are bigger issues at play
            IF r.tree_depth <= 20 THEN
                l_found_dep := true ;
            END IF ;

            INSERT INTO temp_obj_deps (
                    tree_depth,
                    object_oid,
                    schema_name,
                    object_name,
                    full_object_name,
                    object_type,
                    directory_name,
                    file_name,
                    calling_signature )
                        VALUES
                    (
                    r.tree_depth,
                    r.object_oid::oid, -- ::oid as plpgsql_check was being picky
                    r.schema_name,
                    r.object_name,
                    r.full_object_name,
                    r.object_type,
                    r.directory_name,
                    r.file_name,
                    r.calling_signature ) ;

        END LOOP ;

    END LOOP ;

    -- The next problem that needs solving is that the dependency order may not be correct
    -- due to some objects showing up more than once at different levels of the dependency
    -- tree for the table.
    l_found_dep := true ;
    while l_found_dep
    LOOP
        l_found_dep := false ;

        FOR r IN (
            WITH n AS (
                SELECT DISTINCT tad.object_oid,
                        tad.dep_object_oid,
                        dep.tree_depth,
                        CASE
                            WHEN dep.tree_depth <= obj.tree_depth THEN obj.tree_depth + 1
                            ELSE dep.tree_depth
                            END AS new_depth
                    FROM temp_all_deps tad
                    JOIN temp_obj_deps obj
                        ON ( obj.object_oid = tad.object_oid )
                    JOIN temp_obj_deps dep
                        ON ( dep.object_oid = tad.dep_object_oid )
                    WHERE tad.object_oid <> tad.dep_object_oid
            )
            SELECT n.dep_object_oid AS object_oid,
                    n.new_depth
                FROM n
                WHERE n.tree_depth <> n.new_depth )
        LOOP

            -- Guard against a runaway loop here also. Shouldn't happen (unless the query is changed)
            IF r.new_depth <= 20 THEN
                l_found_dep := true ;
            END IF ;

            UPDATE temp_obj_deps SET tree_depth = r.new_depth
                WHERE object_oid = r.object_oid ;

        END LOOP ;

    END LOOP ;

    -- Generate the commands for dropping the dependent objects
    FOR r IN (
        WITH x AS (
            SELECT object_type,
                    full_object_name,
                    calling_signature,
                    min ( tree_depth ) AS tree_depth
                FROM temp_obj_deps
                WHERE tree_depth > 0
                GROUP BY object_type,
                    full_object_name,
                    calling_signature
        )
        SELECT concat_ws ( ' ', 'DROP', object_type, full_object_name, '(' || calling_signature || ')', ';' ) AS cmd
            FROM x
            ORDER BY tree_depth DESC,
                full_object_name,
                calling_signature )
    LOOP

        l_result := concat_ws ( l_new_line, l_result, '', r.cmd ) ;

    END LOOP ;

    ----------------------------------------------------------------------------
    -- If the object is not a table then drop and re-create it
    IF l_object_type <> 'table' THEN

        l_result := concat_ws ( l_new_line, l_result, '', l_drop_cmd, '', l_create_cmd ) ;

    END IF ;

    ----------------------------------------------------------------------------
    -- Generate the psql commands for re-creating the dependent objects to ensure that they point to the new vs the old
    -- TODO: restore the grants on dependent objects also
    FOR r IN (
        WITH x AS (
            SELECT directory_name,
                    file_name,
                    schema_name,
                    object_name,
                    full_object_name,
                    max ( tree_depth ) AS tree_depth
                FROM temp_obj_deps
                WHERE tree_depth > 0
                GROUP BY directory_name,
                    file_name,
                    schema_name,
                    object_name,
                    full_object_name
        )
        SELECT schema_name,
                object_name,
                '\i ' || concat_ws ( '/', directory_name, file_name ) AS cmd
            FROM x
            ORDER BY tree_depth ASC,
                full_object_name )
    LOOP

        l_result := concat_ws ( l_new_line, l_result, '', r.cmd ) ;

        FOR r2 IN (
            WITH base AS (
                SELECT privilege_type,
                        CASE
                            WHEN object_type NOT IN ( 'table', 'view', 'materialized view', 'foreign table' ) THEN upper ( object_type )
                            END AS obj_type,
                        CASE
                            WHEN object_type IN ( 'schema', 'database' ) THEN object_name
                            ELSE object_schema || '.' || object_name
                            END AS obj_name,
                        grantee,
                        CASE
                            WHEN is_grantable THEN 'WITH GRANT OPTION'
                            END AS with_grant
                    FROM util_meta.object_grants
                    WHERE object_schema = r.schema_name
                        AND object_name = r.object_name
                    ORDER BY privilege_type,
                        grantee
            )
            SELECT concat_ws ( ' ', 'GRANT', privilege_type, 'ON', obj_type, obj_name, 'TO', grantee, with_grant, ';' ) AS cmd
                FROM base )
        LOOP

            l_grants := array_append ( l_grants, r2.cmd ) ;

        END LOOP ;

    END LOOP ;

execute 'DROP TABLE temp_all_deps' ;
DROP TABLE temp_obj_deps ;

    ----------------------------------------------------------------------------
    ----------------------------------------------------------------------------
    -- Restore any grants on the migrated object
    FOR r IN (
        WITH base AS (
            SELECT privilege_type,
                    CASE
                        WHEN object_type NOT IN ( 'table', 'view', 'materialized view', 'foreign table' ) THEN upper ( object_type )
                        END AS obj_type,
                    CASE
                        WHEN object_type IN ( 'schema', 'database' ) THEN object_name
                        ELSE object_schema || '.' || object_name
                        END AS obj_name,
                    grantee,
                    CASE
                        WHEN is_grantable THEN 'WITH GRANT OPTION'
                        END AS with_grant
                FROM util_meta.object_grants
                WHERE object_schema = a_object_schema
                    AND object_name = a_object_name
                ORDER BY privilege_type,
                    grantee
        )
        SELECT concat_ws ( ' ', 'GRANT', privilege_type, 'ON', obj_type, obj_name, 'TO', grantee, with_grant, ';' ) AS cmd
            FROM base )
    LOOP

        l_grants := array_append ( l_grants, r.cmd ) ;

    END LOOP ;

    IF array_length ( l_grants, 1 ) > 0 THEN
        l_result := concat_ws ( l_new_line, l_result, '', array_to_string ( l_grants, l_new_line ) ) ;

    END IF ;

    ----------------------------------------------------------------------------
    ----------------------------------------------------------------------------
    IF l_object_type = 'table' THEN

        ------------------------------------------------------------------------
        -- Set a reminder
        l_result := concat_ws ( l_new_line, l_result, '', '-- Remember to "DROP TABLE bak_' || l_full_object_name || ' ;" once the migration is verified', 'RAISE NOTICE ''Remember to "DROP TABLE bak_' || l_full_object_name || ' ;" once the migration is verified''' ) ;

    END IF ;

    RETURN concat_ws ( l_new_line, l_result, '' ) ;

END ;
$$ ;
