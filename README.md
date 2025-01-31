# sqlfmt

An attempt at a somewhat opinionated SQL formatting utility.

## Goals

* To not break anything (no adding, removing, or unintentionally re-arranging
code elements). The code, post formatting, should work the same as it did prior
to formatting.

* To format DML for various DBMS dialects with a primary focus on PostgreSQL,
SQLite, and Oracle.

* To format basic DDL.

* To format PostgreSQL functions and procedures where the language is either
plpgsql or sql.

* To format Oracle functions, procedures, and packages (PL/SQL). (TODO)

## Configuration

Formatting can be tuned using the parameters described below.

When fully implemented, values for configuration parameters, if different from
the default, can be specified using a configuration file, command line
arguments, or file directives.

Values are to be evaluated in the following order:

 1. default
 1. configuration file entries
 1. arguments to the sqlfmt command
 1. file directives

Note that:

* Parameter names are case-insensitive.
* The last matching entry is the one that will be used.

| Parameter         | default  | cfg file | command flag | file directive |
| ----------------- | -------- | -------- | ------------ | -------------- |
| configFile        | TODO     | n/a      | -c           | n/a            |
| dialect           | standard | [x]      | -d           | [x]            |
| indentSize        | 4        | [x]      | -indent      | [x]            |
| keywordCase       | upper    | [x]      | -k           | [x]            |
| maxLineLength     | 120      | [x]      | -l           | [x]            |
| preserveQuoting   | false    | [x]      | -q           | [x]            |
| wrapMultiTuples   | none     | [x]      | -t           | [x]            |
| inputFile         | stdin    | n/a      | -i           | n/a            |
| outputFile        | stdout   | n/a      | -o           | n/a            |
| noFormat          | false    | n/a      | n/a          | [x]            |

File directives are specified by placing a comment as the first line of the
file that contains the parameters to set as a semi-colon separated list. The
comment needs to start with "sqlfmt" followed by the parameters to set. The
primary intent behind file directives is to accommodate groups of files that
may target different database engines and also for indicating files that should
not have their formatting messed with.

 * **configFile** The configuration file to use for setting parameters.

 * **dialect** This is the database dialect to use for formatting. Dialect
 values are case-insensitive with valid values being:

| Value         | Alternate values | DBMS Dialect                      |
| ------------- | ---------------- | --------------------------------- |
| standard      |                  | The SQL standard                  |
| postgresql    | postgres, pg     | PostgreSQL                        |
| sqlite        |                  | SQLite                            |
| oracle        | ora              | Oracle                            |
| mariadb       |                  | MariaDB (best guess)              |
| msaccess      |                  | Microsoft Access (best guess)     |
| mssql         |                  | Microsoft SQL-Server (best guess) |
| mysql         |                  | MySQL (best guess)                |

Where "best guess" simply means that these are dialects that I have rarely
used/do not currently use but have been included because they might work well
enough and someone else may find them useful.

 * **indentSize** This is an integer value indicating the number of spaces to
 use when indenting. The default is to use 4 spaces per indent. Setting this
 value to 0 (zero) causes sqlfmt to use tabs for indentation instead of spaces.

 * **keywordCase** Indicates how specific keywords (such as SELECT, UPDATE,
 DELETE, GRANT, REVOKE, CREATE, etc.) are capitalized.

| Value | Description                                                          |
| ----- | -------------------------------------------------------------------- |
| upper | Set select keywords to upper case (other keywords will be set to lower case |
| lower | Set all keywords to lower case                                       |

 * **maxLineLength** This is an integer value indicating the number of
 characters in a line before sqlfmt attempts to wrap the line.

 * **preserveQuoting** This is a boolean used to tell sqlfmt to not attempt to
 unquote identifiers.

 * **wrapMultiTuples** This instructs sqlfmt how to treat VALUES statements
 that contain multiple tuples.

| Value | Description                                                          |
| ----- | -------------------------------------------------------------------- |
| all   | Each element is placed on a separate line                            |
| long  | Wrap elements when the length of the tuple exceeds the maxLineLength |
| none  | All elements are placed on the same line                             |

Note that this only applies to VALUES statements that contain multiple tuples.
For statements that only contain one tuple, the elements in the tuple will wrap
if there are more than 3 elements OR if the length of the elements exceeds the
maxLineLength.

 * **inputFile** The file to format.

 * **outputFile** The file to write the formatted results to.

 * **noFormat** This is a boolean used to indicate that the file should not be
 formatted. It should be noted that this option only really makes sense as a
 file directive.

## Compiling

    ```
    cd cmd
    go build sqlfmt.go
    ```

## Usage

 ```./sqlfmt -h```

 ```./sqlfmt -d postgresql -i /path/to/file/format.sql -o /path/to/write/file/to.sql```
