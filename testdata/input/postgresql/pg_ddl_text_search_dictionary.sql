-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-createtsdictionary.html
https://www.postgresql.org/docs/17/sql-altertsdictionary.html
https://www.postgresql.org/docs/17/sql-droptsdictionary.html
*/

CREATE TEXT SEARCH DICTIONARY my_russian (
    template = snowball,
    language = russian,
    stopwords = myrussian
);

ALTER TEXT SEARCH DICTIONARY my_dict ( StopWords = newrussian );

ALTER TEXT SEARCH DICTIONARY my_dict ( language = dutch, StopWords );

ALTER TEXT SEARCH DICTIONARY my_dict ( dummy );

DROP TEXT SEARCH DICTIONARY english;

/*
CREATE TEXT SEARCH DICTIONARY name (
    TEMPLATE = template
    [, option = value [, ... ]]
)

ALTER TEXT SEARCH DICTIONARY name (
    option [ = value ] [, ... ]
)
ALTER TEXT SEARCH DICTIONARY name RENAME TO new_name
ALTER TEXT SEARCH DICTIONARY name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
ALTER TEXT SEARCH DICTIONARY name SET SCHEMA new_schema

DROP TEXT SEARCH DICTIONARY [ IF EXISTS ] name [ CASCADE | RESTRICT ]
*/
