-- sqlfmt dialect: PostgreSQL

CREATE OR REPLACE FUNCTION util_meta.json_identifier (
    a_identifier text default null,
    a_json_casing text default null )
    RETURNS text
    LANGUAGE plpgsql
    STABLE
    SECURITY INVOKER
AS $func$
/**
Function json_identifier takes a database identifier (table name, column name, etc. ) and
    returns the json identifier for the identifier

    (lower) camelCase form of the identifier

| Parameter                      | In/Out | Datatype   | Description                                        |
| ------------------------------ | ------ | ---------- | -------------------------------------------------- |
| a_identifier                   | in     | text       | The identifier to transform                        |
| a_json_casing                  | in     | text       | The type of JSON casing to use {lowerCamel, upperCamel, snake} (defaults to lowerCamel) |

| Input             | JSON casing | Output            |
| ----------------- | ----------- | ----------------- |
| id                | lowerCamel  | id                |
| my_snazzy_id      | null        | mySnazzyId        |
| my_snazzy_id      | lowerCamel  | mySnazzyId        |
| my_snazzy_id      | upperCamel  | MySnazzyId        |
| my_snazzy_id      | snake       | my_snazzy_id      |

*/
declare

l_tokens text[] ;
l_token text ;
l_casing text ;
l_identifier text ;
l_separator text ;

begin

l_casing := coalesce ( util_meta.resolve_parameter ( a_name => 'json_casing', a_value =>  a_json_casing ), 'lowerCamel' ) as chars ;

if l_casing = 'snake' then
l_separator := '_' ;
else
l_separator := '' ;
end if ;

l_tokens := '{}'::text[] ;

l_identifier := regexp_replace ( a_identifier, '[^\w]', '_', 'g' ) ;

-- some form of camel case
foreach l_token in array string_to_array ( l_identifier, '_' ) loop

if l_token is null or l_token = '' then
null;

elsif l_casing = 'snake' then

l_tokens := array_append ( l_tokens, lower ( l_token ) ) ;

elsif l_casing = 'upperCamel' then

l_tokens := array_append ( l_tokens, initcap ( l_token ) ) ;

else -- lowerCamel

if cardinality ( l_tokens ) = 0 then
l_tokens := array_append ( l_tokens, lower ( l_token ) ) ;
else
l_tokens := array_append ( l_tokens, initcap ( l_token ) ) ;
end if ;

end if ;

end loop ;

return array_to_string ( l_tokens, l_separator ) ;

end ;
$func$ ;

alter function util_meta.json_identifier (text,text) owner to postgres ;

revoke execute on function util_meta.json_identifier (text,text) from public ;

grant execute on function util_meta.json_identifier (text,text) to postgres ;
