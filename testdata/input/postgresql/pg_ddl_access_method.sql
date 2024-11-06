-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/17/sql-create-access-method.html
https://www.postgresql.org/docs/17/sql-drop-access-method.html
*/

CREATE ACCESS METHOD heptree TYPE INDEX HANDLER heptree_handler;

DROP ACCESS METHOD heptree;

/*
CREATE ACCESS METHOD name
    TYPE access_method_type
    HANDLER handler_function

DROP ACCESS METHOD [ IF EXISTS ] name [ CASCADE | RESTRICT ]

*/
