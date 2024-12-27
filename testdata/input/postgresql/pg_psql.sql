 -- sqlfmt d:postgres doNotFormat


SET statement_timeout = 0 ;
SET client_encoding = 'UTF8' ;
SET standard_conforming_strings = on ;
SET check_function_bodies = true ;
SET client_min_messages = warning ;

CREATE EXTENSION IF NOT EXISTS dblink ;

CREATE SCHEMA IF NOT EXISTS util_log ;

COMMENT ON SCHEMA util_log IS 'Schema and objects for logging database function and procedure calls' ;

ALTER SCHEMA util_log OWNER TO app_owner ;

\unset ON_ERROR_STOP

CREATE SERVER loopback_dblink FOREIGN DATA WRAPPER dblink_fdw
    OPTIONS ( hostaddr '127.0.0.1', dbname 'my_data' ) ;

ALTER SERVER loopback_dblink OWNER TO my_log_user ;

CREATE USER MAPPING FOR app_developer SERVER loopback_dblink
    OPTIONS ( user 'my_log_user', password '********' ) ;

\set ON_ERROR_STOP

-- Tables --------------------------------------------------------------
\i util_log/table/st_log_level.sql
\i util_log/table/dt_proc_log.sql
\i util_log/table/dt_last_logged.sql

-- Views ---------------------------------------------------------------
\i util_log/view/dv_proc_log.sql
\i util_log/view/dv_proc_log_today.sql
\i util_log/view/dv_proc_log_last_hour.sql
\i util_log/view/dv_proc_log_last_day.sql
\i util_log/view/dv_proc_log_last_week.sql

-- Functions -----------------------------------------------------------
\i util_log/function/dici.sql
\i util_log/function/manage_partitions.sql
\i util_log/function/update_last_logged.sql

-- Procedures ----------------------------------------------------------
\i util_log/procedure/log_to_dblink.sql
\i util_log/procedure/log_begin.sql
\i util_log/procedure/log_debug.sql
\i util_log/procedure/log_exception.sql
\i util_log/procedure/log_finish.sql
\i util_log/procedure/log_info.sql

-- Query bug -----------------------------------------------------------
\i util_log/function/query_bug.sql

GRANT EXECUTE ON FUNCTION util_log.manage_partitions TO app_developer ;
GRANT INSERT ON util_log.dt_proc_log TO app_developer ;
GRANT USAGE ON SCHEMA util_log TO app_developer ;

GRANT EXECUTE ON FUNCTION util_log.dici TO app_developer ;
GRANT EXECUTE ON PROCEDURE util_log.log_begin TO app_developer ;
GRANT EXECUTE ON PROCEDURE util_log.log_debug TO app_developer ;
GRANT EXECUTE ON PROCEDURE util_log.log_exception TO app_developer ;
GRANT EXECUTE ON PROCEDURE util_log.log_finish TO app_developer ;
GRANT EXECUTE ON PROCEDURE util_log.log_info TO app_developer ;
