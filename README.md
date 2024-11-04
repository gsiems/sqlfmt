# sqlfmt

A somewhat opinionated SQL formatter.

## Goals

* To not break anything (no adding, removing, or unintentionally re-arranging
code elements). The code, post formatting, should work the same as it did prior
to formatting.

* To format DML for various DBMS dialects with a primary focus on PostgreSQL,
SQLite, and Oracle.

* To format basic DDL.

* To format PostgreSQL functions and procedures where the language is either
plpgsql or sql.

* To format Oracle functions, procedures, and packages (PL/SQL).
