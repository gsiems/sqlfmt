-- sqlfmt d:postgres

-- sqlfmt d:postgres

GRANT TEMPORARY ON DATABASE app_db TO app_owner ;

CREATE OR REPLACE FUNCTION app_data.refresh_mv()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $trig$
/**
Function refresh_mv refreshes the materialized views after the update to any of their underlying tables

*/
BEGIN

    REFRESH MATERIALIZED VIEW CONCURRENTLY app_data.mv_01 ;
    REFRESH MATERIALIZED VIEW CONCURRENTLY app_data.mv_02 ;

    RETURN NEW ;

END ;
$trig$ ;

ALTER FUNCTION app_data.refresh_mv OWNER TO app_owner ;

--------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS zz_refresh_mv ON app_data.st_table_01 ;

CREATE TRIGGER zz_refresh_mv
    AFTER INSERT OR UPDATE OR DELETE ON app_data.st_table_01
    FOR EACH ROW
    EXECUTE FUNCTION app_data.refresh_mv () ;

--------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS zz_refresh_mv ON app_data.st_table_02 ;

CREATE TRIGGER zz_refresh_mv
    AFTER INSERT OR UPDATE OR DELETE ON app_data.st_table_02
    FOR EACH ROW
    EXECUTE FUNCTION app_data.refresh_mv () ;

--------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS zz_refresh_mv ON app_data.st_table_03 ;

CREATE TRIGGER zz_refresh_mv
    AFTER INSERT OR UPDATE OR DELETE ON app_data.st_table_03
    FOR EACH ROW
    EXECUTE FUNCTION app_data.refresh_mv () ;

--------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS zz_refresh_mv ON app_data.dt_table_01 ;

CREATE TRIGGER zz_refresh_mv
    AFTER INSERT OR UPDATE OR DELETE ON app_data.dt_table_01
    FOR EACH ROW
    EXECUTE FUNCTION app_data.refresh_mv () ;

--------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS zz_refresh_mv ON app_data.dt_table_02 ;

CREATE TRIGGER zz_refresh_mv
    AFTER INSERT OR UPDATE OR DELETE ON app_data.dt_table_02
    FOR EACH ROW
    EXECUTE FUNCTION app_data.refresh_mv () ;

CREATE MATERIALIZED VIEW app_rpt.rpt_mv
AS
SELECT *
    FROM fdw_schema.fdw_table
    WITH NO DATA ;

CREATE UNIQUE INDEX rpt_mv_idx ON app_rpt.rpt_mv (
    id,
    col_02,
    col_03 ) ;

ALTER MATERIALIZED VIEW app_rpt.rpt_mv OWNER TO app_owner ;

GRANT SELECT ON app_rpt.rpt_mv TO current_user ;

DO
$$
BEGIN
    REFRESH MATERIALIZED VIEW app_rpt.rpt_mv ;
EXCEPTION
    WHEN PROHIBITED_SQL_STATEMENT_ATTEMPTED THEN
        RAISE WARNING 'No user mapping. This happens at build time, and can be safely ignored' ;
END ;
$$ ;
