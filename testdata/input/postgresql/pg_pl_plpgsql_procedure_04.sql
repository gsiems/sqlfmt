-- sqlfmt d:postgres

GRANT SELECT ON tgt_schema.tgt_table TO src_schema_owner ;
GRANT SELECT, INSERT, UPDATE, DELETE ON tgt_schema.tgt_table TO src_schema_owner ;

GRANT ALL ON SEQUENCE tgt_schema.archive_oid_seq TO src_schema_owner ;

CREATE OR REPLACE PROCEDURE src_schema.sync_tgt_data (
    a_id in integer default null,
    a_user in text default null,
    a_err inout text default null )
SECURITY DEFINER
SET search_path = pg_catalog, public
LANGUAGE plpgsql
AS $proc$
DECLARE

    src_r record ;
    tgt_r record ;

    l_src_action text := 'none' ;
    l_tgt_action text := 'none' ;

    l_timestamp timestamp(3) = clock_timestamp() AT TIME ZONE 'UTC' ;
    l_message_text text ;
    l_err text ;

    c_to_date constant timestamp without time zone := '9999-12-31 23:59:59'::timestamp without time zone ;

BEGIN

    call util_log.log_begin (
        util_log.dici ( a_id ),
        util_log.dici ( a_user ) ) ;

    -- If the upsert started on the tgt_schema side then we are done
    IF a_user ~ 'tgt_schema' THEN
        RETURN ;
    END IF ;


    FOR src_r IN (
        SELECT id,
                col_01,
                col_02,
                col_03,
                col_04,
                col_05,
                col_06,
                col_07,
                col_08,
                col_09
            FROM src_schema.get_src_data ( a_id ) x ) LOOP

        -- If the valid end is not infinity then we can assume that the record has been set as deleted;
        -- especially if the valid end - 1 day is less than the current
        IF src_r.sys_operation = 'D' OR now()::timestamp > ( upper ( src_r.valid_period ) - '1 day'::interval )::timestamp THEN
            l_src_action := 'delete' ;
        END IF ;

        -- Default to inserting if we are not deleting
        IF l_src_action = 'none' THEN
            l_src_action := 'insert' ;
        END IF ;

        -- set the default desired action
        IF l_src_action = 'delete' THEN
            l_tgt_action := l_src_action ;
        ELSE
            l_tgt_action := 'insert' ;
        END IF ;

        FOR tgt_r IN (
            SELECT id,
                    col_01,
                    col_02,
                    col_03,
                    col_04,
                    col_05,
                    col_06,
                    col_07,
                    col_08,
                    col_09
                FROM tgt_schema.dv_tgt_table
                WHERE id = src_r.id ) LOOP

            IF l_src_action IN ( 'insert', 'update' ) THEN

                l_tgt_action := 'none' ;

                IF tgt_r.col_01 IS DISTINCT FROM src_r.col_01
                        OR tgt_r.col_02 IS DISTINCT FROM src_r.col_02
                        OR tgt_r.col_03 IS DISTINCT FROM src_r.col_03
                        OR tgt_r.col_04 IS DISTINCT FROM src_r.col_04
                        OR tgt_r.col_05 IS DISTINCT FROM src_r.col_05
                        OR tgt_r.col_06 IS DISTINCT FROM src_r.col_06
                        OR tgt_r.col_07 IS DISTINCT FROM src_r.col_07
                        OR tgt_r.col_08 IS DISTINCT FROM src_r.col_08
                        OR tgt_r.col_09 IS DISTINCT FROM src_r.col_09
                    THEN

                    l_tgt_action := 'update' ;

                END IF ;

            END IF ;

        END LOOP ;

        IF l_tgt_action = 'none' THEN
            RETURN ;
        END IF ;

        IF l_tgt_action IN ( 'update', 'delete' ) THEN

            BEGIN

                UPDATE tgt_schema.tgt_table
                    SET valid_to_date = l_timestamp
                    WHERE id = src_r.id
                        AND valid_to_date = c_to_date ;

            EXCEPTION
                WHEN unique_violation THEN

                    l_err := SQLSTATE::text || ' - ' || SQLERRM ;
                    GET STACKED DIAGNOSTICS l_message_text = MESSAGE_TEXT ;

                    call util_log.log_exception ( l_err ) ;

                    IF ( l_message_text like '%valid_ct2%' ) THEN
                        DELETE FROM tgt_schema.tgt_table
                            WHERE id = src_r.id
                                AND valid_to_date = c_to_date ;
                    ELSE
                        a_err := substr ( l_err, 1, 200 ) ;
                        RETURN ;
                    END IF ;

                WHEN others THEN
                    a_err := substr ( SQLSTATE::text || ' - ' || SQLERRM, 1, 200 ) ;
                    call util_log.log_exception ( SQLSTATE::text || ' - ' || SQLERRM ) ;
                    RETURN ;
            END ;

        END IF ;

        IF l_tgt_action IN ( 'insert', 'update' ) THEN

            INSERT INTO tgt_schema.tgt_table (
                        id,
                        col_01,
                        col_02,
                        col_03,
                        col_04,
                        col_05,
                        col_06,
                        col_07,
                        col_08,
                        col_09,
                        valid_from_date,
                        valid_to_date )
                VALUES (
                        src_r.id,
                        src_r.col_01,
                        src_r.col_02,
                        src_r.col_03,
                        src_r.col_04,
                        src_r.col_05,
                        src_r.col_06,
                        src_r.col_07,
                        src_r.col_08,
                        src_r.col_09,
                        l_timestamp,
                        c_to_date ) ;

        END IF ;

    END LOOP ;

EXCEPTION
    WHEN others THEN
        a_err := SQLSTATE::text || ' - ' || SQLERRM ;
        call util_log.log_exception ( SQLSTATE::text || ' - ' || SQLERRM ) ;
END ;
$proc$ ;

ALTER PROCEDURE src_schema.sync_tgt_data ( integer, text, text ) OWNER TO src_schema_owner ;

GRANT EXECUTE ON PROCEDURE src_schema.sync_tgt_data ( integer, text, text ) TO app_owner ;
