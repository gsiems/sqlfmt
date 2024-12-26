package dialect

import (
	"strings"
)

type DbDialect interface {
	Dialect() int
	DialectName() string
	CaseFolding() int
	IdentQuoteChar() string
	StringQuoteChar() string
	MaxOperatorLength() int
	IsDatatype(s ...string) bool
	keyword(s string) (bool, bool)
	IsKeyword(s string) bool
	IsReservedKeyword(s string) bool
	IsOperator(s string) bool
	IsLabel(s string) bool
	IsIdentifier(s string) bool
}

func StrToDialect(v string) int {

	switch strings.ToLower(v) {
	case "mariadb":
		return MariaDB
	case "mssql":
		return MSSQL
	case "msaccess", "access":
		return MSAccess
	case "mysql":
		return MySQL
	case "oracle", "ora":
		return Oracle
	case "postgresql", "postgres", "pg":
		return PostgreSQL
	case "sqlite":
		return SQLite
	}
	// default to the standard
	return StandardSQL
}

func NewDialect(v string) DbDialect {

	switch StrToDialect(v) {
	case MariaDB:
		return NewMariaDBDialect()
	case MSAccess:
		return NewMSAccessDialect()
	case MSSQL:
		return NewMSSQLDialect()
	case MySQL:
		return NewMySQLDialect()
	case Oracle:
		return NewOracleDialect()
	case PostgreSQL:
		return NewPostgreSQLDialect()
	case SQLite:
		return NewSQLiteDialect()
	}
	// default to the standard
	return NewStandardSQLDialect()
}
