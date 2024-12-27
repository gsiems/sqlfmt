-- sqlfmt dialect: PostgreSQL

CREATE OR REPLACE FUNCTION util_meta.get_schema (
    a_schema_name text )
RETURNS TABLE (
    schema_oid oid,
    schema_name text,
    owner_oid oid,
    directory_name text
)
LANGUAGE sql
STABLE
SECURITY DEFINER
AS $$
WITH base AS (
    SELECT n.oid AS schema_oid,
            n.nspname::text AS schema_name,
            n.nspowner AS owner_oid,
            pg_catalog.pg_get_userbyid ( n.nspowner )::text AS owner_name,
            concat_ws ( '/', 'schema', n.nspname::text ) AS directory_name
        FROM pg_catalog.pg_namespace n
        LEFT JOIN pg_catalog.pg_extension px
            ON ( px.extnamespace = n.oid )
        WHERE n.nspname !~ '^pg_'
            AND n.nspname <> 'information_schema'
            AND px.oid IS NULL
            AND n.nspname::text = a_schema_name
)
SELECT *
    FROM base ;

$$ ;

ALTER FUNCTION util_meta.get_schema ( text ) OWNER TO postgres ;

REVOKE EXECUTE ON FUNCTION util_meta.get_schema ( text ) FROM public ;

GRANT EXECUTE ON FUNCTION util_meta.get_schema ( text ) TO postgres ;
