package dialect

import "strings"

type MSAccessDialect struct {
	dialect int
	name    string
}

func NewMSAccessDialect() *MSAccessDialect {
	var d MSAccessDialect

	d.dialect = MSAccess
	d.name = "MSAccess"

	return &d
}

func (d MSAccessDialect) Dialect() int {
	return d.dialect
}
func (d MSAccessDialect) DialectName() string {
	return d.name
}
func (d MSAccessDialect) CaseFolding() int {
	return NoFolding
}
func (d MSAccessDialect) IdentQuoteChar() string {
	return "\""
}
func (d MSAccessDialect) StringQuoteChar() string {
	return "'"
}

// MaxOperatorLength returns the length of the longest operator
// supported by the dialect
func (d MSAccessDialect) MaxOperatorLength() int {
	return 2
}

// IsDatatype returns a boolean indicating if the supplied string
// (or string slice) is considered to be a datatype in MSAccess
func (d MSAccessDialect) IsDatatype(s ...string) bool {

	var msAccessDatatypes = map[string]bool{
		"attachment":         true,
		"autonumber":         true,
		"byte":               true,
		"calculated":         true,
		"calculated field":   true,
		"currency":           true,
		"date/time":          true,
		"date/time extended": true,
		"double":             true,
		"hyperlink":          true,
		"integer":            true,
		"large number":       true,
		"long":               true,
		"long text":          true,
		"lookup":             true,
		"lookup wizard":      true,
		"memo":               true,
		"number":             true,
		"ole object":         true,

		"rich text":          true,
		"short text":         true,
		"single":             true,
		"text":               true,
		"yes/no":             true,
	}

	var z []string

	for i, v := range s {
		switch {
		case i == 0:
			z = append(z, v)
		default:
			z = append(z, " "+v)
		}
	}

	k := strings.ToLower(strings.Join(z, ""))
	if _, ok := msAccessDatatypes[k]; ok {
		return true
	}

	return false
}

func (d MSAccessDialect) keyword(s string) (bool, bool) {

	/*
	   Microsoft Access keywords

	   https://learn.microsoft.com/en-us/office/client-developer/access/reserved-words-access-custom-web-app#access-reserved-keywords

	   The isReserved value is set to false as there is no indication (from
	   the above link) if the keywords are reserved or not.

	*/

	// map[keyword]isReserved
	var msAccessKeywords = map[string]bool{
		"ADD":                            true,
		"ALL":                            true,
		"ALTER":                          true,
		"AND":                            true,
		"ANY":                            true,
		"ASC":                            true,
		"AS":                             true,
		"AUTHORIZATION":                  true,
		"BACKUP":                         true,
		"BEGIN":                          true,
		"BETWEEN":                        true,
		"BREAK":                          true,
		"BROWSE":                         true,
		"BULK":                           true,
		"BY":                             true,
		"CASCADE":                        true,
		"CASE":                           true,
		"CHECKPOINT":                     true,
		"CHECK":                          true,
		"CLOSE":                          true,
		"CLUSTERED":                      true,
		"COALESCE":                       true,
		"COLLATE":                        true,
		"COLUMN":                         true,
		"COMMIT":                         true,
		"COMPUTE":                        true,
		"CONSTRAINT":                     true,
		"CONTAINSTABLE":                  true,
		"CONTAINS":                       true,
		"CONTINUE":                       true,
		"CONVERT":                        true,
		"CREATE":                         true,
		"CROSS":                          true,
		"CURRENCY":                       true,
		"CURRENT_DATE":                   true,
		"CURRENT_TIMESTAMP":              true,
		"CURRENT_TIME":                   true,
		"CURRENT":                        true,
		"CURRENT_USER":                   true,
		"CURSOR":                         true,
		"DATABASE":                       true,
		"DATE":                           true,
		"DATEWITHTIME":                   true,
		"DAYOFYEAR":                      true,
		"DAY":                            true,
		"DBCC":                           true,
		"DEALLOCATE":                     true,
		"DECLARE":                        true,
		"DEFAULT":                        true,
		"DELETE":                         true,
		"DENY":                           true,
		"DESC":                           true,
		"DISK":                           true,
		"DISTINCT":                       true,
		"DISTRIBUTED":                    true,
		"DOUBLE":                         true,
		"DROP":                           true,
		"DUMP":                           true,
		"ELSE":                           true,
		"END":                            true,
		"ERRLVL":                         true,
		"ESCAPE":                         true,
		"EXCEPT":                         true,
		"EXEC":                           true,
		"EXECUTE":                        true,
		"EXISTS":                         true,
		"EXIT":                           true,
		"EXTERNAL":                       true,
		"FETCH":                          true,
		"FILE":                           true,
		"FILLFACTOR":                     true,
		"FLOAT":                          true,
		"FOREIGN":                        true,
		"FOR":                            true,
		"FREETEXTTABLE":                  true,
		"FREETEXT":                       true,
		"FROM":                           true,
		"FULL":                           true,
		"FUNCTION":                       true,
		"GOTO":                           true,
		"GRANT":                          true,
		"GROUP":                          true,
		"HAVING":                         true,
		"HOLDLOCK":                       true,
		"HOUR":                           true,
		"IDENTITYCOL":                    true,
		"IDENTITY_INSERT":                true,
		"IDENTITY":                       true,
		"IF":                             true,
		"INDEX":                          true,
		"INNER":                          true,
		"INSERT":                         true,
		"INTEGER":                        true,
		"INTERSECT":                      true,
		"INTO":                           true,
		"IN":                             true,
		"ISO_WEEK":                       true,
		"IS":                             true,
		"JOIN":                           true,
		"KEY":                            true,
		"KILL":                           true,
		"LEFT":                           true,
		"LIKE":                           true,
		"LINENO":                         true,
		"LOAD":                           true,
		"LONGTEXT":                       true,
		"MERGE":                          true,
		"MILLISECOND":                    true,
		"MINUTE":                         true,
		"MONTH":                          true,
		"NATIONAL":                       true,
		"NOCHECK":                        true,
		"NONCLUSTERED":                   true,
		"NO":                             true,
		"NOT":                            true,
		"NULLIF":                         true,
		"NULL":                           true,
		"OFFSETS":                        true,
		"OFF":                            true,
		"OF":                             true,
		"ON":                             true,
		"OPENDATASOURCE":                 true,
		"OPENQUERY":                      true,
		"OPENROWSET":                     true,
		"OPEN":                           true,
		"OPENXML":                        true,
		"OPTION":                         true,
		"ORDER":                          true,
		"OR":                             true,
		"OUTER":                          true,
		"OVER":                           true,
		"PERCENT":                        true,
		"PIVOT":                          true,
		"PLAN":                           true,
		"PRECISION":                      true,
		"PRIMARY":                        true,
		"PRINT":                          true,
		"PROCEDURE":                      true,
		"PROC":                           true,
		"PUBLIC":                         true,
		"QUARTER":                        true,
		"RAISERROR":                      true,
		"READTEXT":                       true,
		"READ":                           true,
		"RECONFIGURE":                    true,
		"REFERENCES":                     true,
		"REPLICATION":                    true,
		"RESTORE":                        true,
		"RESTRICT":                       true,
		"RETURN":                         true,
		"REVERT":                         true,
		"REVOKE":                         true,
		"RIGHT":                          true,
		"ROLLBACK":                       true,
		"ROWCOUNT":                       true,
		"ROWGUIDCOL":                     true,
		"RULE":                           true,
		"SAVE":                           true,
		"SCHEMA":                         true,
		"SECOND":                         true,
		"SECURITYAUDIT":                  true,
		"SELECT":                         true,
		"SEMANTICKEYPHRASETABLE":         true,
		"SEMANTICSIMILARITYDETAILSTABLE": true,
		"SEMANTICSIMILARITYTABLE":        true,
		"SESSION_USER":                   true,
		"SET":                            true,
		"SETUSER":                        true,
		"SHORTTEXT":                      true,
		"SHUTDOWN":                       true,
		"SOME":                           true,
		"STATISTICS":                     true,
		"SYSTEM_USER":                    true,
		"TABLESAMPLE":                    true,
		"TABLE":                          true,
		"TEXTSIZE":                       true,
		"TEXT":                           true,
		"THEN":                           true,
		"TIME":                           true,
		"TOP":                            true,
		"TO":                             true,
		"TRANSACTION":                    true,
		"TRAN":                           true,
		"TRIGGER":                        true,
		"TRUNCATE":                       true,
		"TRY_CONVERT":                    true,
		"TSEQUAL":                        true,
		"UNION":                          true,
		"UNIQUE":                         true,
		"UNPIVOT":                        true,
		"UPDATETEXT":                     true,
		"UPDATE":                         true,
		"USER":                           true,
		"USE":                            true,
		"VALUES":                         true,
		"VARYING":                        true,
		"VIEW":                           true,
		"WAITFOR":                        true,
		"WEEKDAY":                        true,
		"WEEK":                           true,
		"WHEN":                           true,
		"WHERE":                          true,
		"WHILE":                          true,
		"WITHIN GROUP":                   true,
		"WITHIN":                         true,
		"WITH":                           true,
		"WRITETEXT":                      true,
		"YEAR":                           true,
		"YESNO":                          true,
		"YES":                            true,
		"ABSOLUTE":                       true, // ODBC
		"ACTION":                         true, // ODBC
		"ADA":                            true, // ODBC
		"ALLOCATE":                       true, // ODBC
		"ARE":                            true, // ODBC
		"ASSERTION":                      true, // ODBC
		"AT":                             true, // ODBC
		//"AUTHORIZATION":                  true, // ODBC
		"AVG":                            true, // ODBC
		//"BEGIN":                          true, // ODBC
		//"BETWEEN":                        true, // ODBC
		"BIT_LENGTH":                     true, // ODBC
		"BIT":                            true, // ODBC
		"BOTH":                           true, // ODBC
		//"BY":                             true, // ODBC
		"CASCADED":                       true, // ODBC
		//"CASCADE":                        true, // ODBC
		//"CASE":                           true, // ODBC
		"CAST":                           true, // ODBC
		"CATALOG":                        true, // ODBC
		"CHARACTER_LENGTH":               true, // ODBC
		"CHARACTER":                      true, // ODBC
		"CHAR_LENGTH":                    true, // ODBC
		"CHAR":                           true, // ODBC
		//"CHECK":                          true, // ODBC
		//"CLOSE":                          true, // ODBC
		//"COALESCE":                       true, // ODBC
		//"COLLATE":                        true, // ODBC
		"COLLATION":                      true, // ODBC
		//"COLUMN":                         true, // ODBC
		//"COMMIT":                         true, // ODBC
		"CONNECTION":                     true, // ODBC
		"CONNECT":                        true, // ODBC
		"CONSTRAINTS":                    true, // ODBC
		//"CONSTRAINT":                     true, // ODBC
		//"CONTINUE":                       true, // ODBC
		//"CONVERT":                        true, // ODBC
		"CORRESPONDING":                  true, // ODBC
		"COUNT":                          true, // ODBC
		//"CREATE":                         true, // ODBC
		//"CROSS":                          true, // ODBC
		//"CURRENT_DATE":                   true, // ODBC
		//"CURRENT_TIMESTAMP":              true, // ODBC
		//"CURRENT_TIME":                   true, // ODBC
		//"CURRENT":                        true, // ODBC
		//"CURRENT_USER":                   true, // ODBC
		//"CURSOR":                         true, // ODBC
		//"DATE":                           true, // ODBC
		//"DAY":                            true, // ODBC
		//"DEALLOCATE":                     true, // ODBC
		"DECIMAL":                        true, // ODBC
		//"DECLARE":                        true, // ODBC
		"DEC":                            true, // ODBC
		//"DEFAULT":                        true, // ODBC
		"DEFERRABLE":                     true, // ODBC
		"DEFERRED":                       true, // ODBC
		//"DELETE":                         true, // ODBC
		"DESCRIBE":                       true, // ODBC
		"DESCRIPTOR":                     true, // ODBC
		//"DESC":                           true, // ODBC
		"DIAGNOSTICS":                    true, // ODBC
		"DISCONNECT":                     true, // ODBC
		//"DISTINCT":                       true, // ODBC
		"DOMAIN":                         true, // ODBC
		//"DOUBLE":                         true, // ODBC
		//"DROP":                           true, // ODBC
		//"ELSE":                           true, // ODBC
		"END-EXEC":                       true, // ODBC
		//"END":                            true, // ODBC
		//"ESCAPE":                         true, // ODBC
		"EXCEPTION":                      true, // ODBC
		//"EXCEPT":                         true, // ODBC
		"EXTRACT":                        true, // ODBC
		"FALSE":                          true, // ODBC
		"FIRST":                          true, // ODBC
		"FORTRAN":                        true, // ODBC
		"FOUND":                          true, // ODBC
		//"FROM":                           true, // ODBC
		//"FULL":                           true, // ODBC
		"GET":                            true, // ODBC
		"GLOBAL":                         true, // ODBC
		//"GOTO":                           true, // ODBC
		"GO":                             true, // ODBC
		//"GRANT":                          true, // ODBC
		//"GROUP":                          true, // ODBC
		//"HAVING":                         true, // ODBC
		//"HOUR":                           true, // ODBC
		//"IDENTITY":                       true, // ODBC
		"IMMEDIATE":                      true, // ODBC
		"INCLUDE":                        true, // ODBC
		//"INDEX":                          true, // ODBC
		"INDICATOR":                      true, // ODBC
		"INITIALLY":                      true, // ODBC
		//"INNER":                          true, // ODBC
		"INPUT":                          true, // ODBC
		"INSENSITIVE":                    true, // ODBC
		//"INSERT":                         true, // ODBC
		//"INTEGER":                        true, // ODBC
		//"INTERSECT":                      true, // ODBC
		"INTERVAL":                       true, // ODBC
		//"INTO":                           true, // ODBC
		//"IN":                             true, // ODBC
		"INT":                            true, // ODBC
		"ISOLATION":                      true, // ODBC
		//"IS":                             true, // ODBC
		//"JOIN":                           true, // ODBC
		//"KEY":                            true, // ODBC
		"LANGUAGE":                       true, // ODBC
		"LAST":                           true, // ODBC
		"LEADING":                        true, // ODBC
		//"LEFT":                           true, // ODBC
		"LEVEL":                          true, // ODBC
		//"LIKE":                           true, // ODBC
		"LOCAL":                          true, // ODBC
		"LOWER":                          true, // ODBC
		"MATCH":                          true, // ODBC
		"MAX":                            true, // ODBC
		"MIN":                            true, // ODBC
		//"MINUTE":                         true, // ODBC
		"MODULE":                         true, // ODBC
		//"MONTH":                          true, // ODBC
		"NAMES":                          true, // ODBC
		//"NATIONAL":                       true, // ODBC
		"NATURAL":                        true, // ODBC
		"NCHAR":                          true, // ODBC
		"NEXT":                           true, // ODBC
		"NONE":                           true, // ODBC
		//"NO":                             true, // ODBC
		//"NOT":                            true, // ODBC
		//"NULLIF":                         true, // ODBC
		//"NULL":                           true, // ODBC
		"NUMERIC":                        true, // ODBC
		"OCTET_LENGTH":                   true, // ODBC
		//"OF":                             true, // ODBC
		"ONLY":                           true, // ODBC
		//"ON":                             true, // ODBC
		//"OPEN":                           true, // ODBC
		//"OPTION":                         true, // ODBC
		//"ORDER":                          true, // ODBC
		//"OR":                             true, // ODBC
		//"OUTER":                          true, // ODBC
		"OUTPUT":                         true, // ODBC
		"OVERLAPS":                       true, // ODBC
		"PAD":                            true, // ODBC
		"PARTIAL":                        true, // ODBC
		"PASCAL":                         true, // ODBC
		"POSITION":                       true, // ODBC
		"PREPARE":                        true, // ODBC
		"PRESERVE":                       true, // ODBC
		"PRIOR":                          true, // ODBC
		"PRIVILEGES":                     true, // ODBC
		//"PUBLIC":                         true, // ODBC
		//"READ":                           true, // ODBC
		"REAL":                           true, // ODBC
		//"REFERENCES":                     true, // ODBC
		"RELATIVE":                       true, // ODBC
		//"RESTRICT":                       true, // ODBC
		//"REVOKE":                         true, // ODBC
		//"RIGHT":                          true, // ODBC
		//"ROLLBACK":                       true, // ODBC
		"ROWS":                           true, // ODBC
		//"SCHEMA":                         true, // ODBC
		"SCROLL":                         true, // ODBC
		//"SECOND":                         true, // ODBC
		"SECTION":                        true, // ODBC
		//"SELECT":                         true, // ODBC
		"SESSION":                        true, // ODBC
		//"SESSION_USER":                   true, // ODBC
		//"SET":                            true, // ODBC
		"SIZE":                           true, // ODBC
		"SMALLINT":                       true, // ODBC
		//"SOME":                           true, // ODBC
		"SPACE":                          true, // ODBC
		"SQLCA":                          true, // ODBC
		"SQLCODE":                        true, // ODBC
		"SQLERROR":                       true, // ODBC
		"SQLSTATE":                       true, // ODBC
		"SQL":                            true, // ODBC
		"SQLWARNING":                     true, // ODBC
		"SUBSTRING":                      true, // ODBC
		"SUM":                            true, // ODBC
		//"SYSTEM_USER":                    true, // ODBC
		//"TABLE":                          true, // ODBC
		"TEMPORARY":                      true, // ODBC
		//"THEN":                           true, // ODBC
		"TIMESTAMP":                      true, // ODBC
		//"TIME":                           true, // ODBC
		"TIMEZONE_HOUR":                  true, // ODBC
		"TIMEZONE_MINUTE":                true, // ODBC
		//"TO":                             true, // ODBC
		"TRAILING":                       true, // ODBC
		//"TRANSACTION":                    true, // ODBC
		"TRANSLATE":                      true, // ODBC
		"TRANSLATION":                    true, // ODBC
		"TRIM":                           true, // ODBC
		"TRUE":                           true, // ODBC
		//"UNION":                          true, // ODBC
		//"UNIQUE":                         true, // ODBC
		"UNKNOWN":                        true, // ODBC
		//"UPDATE":                         true, // ODBC
		"UPPER":                          true, // ODBC
		"USAGE":                          true, // ODBC
		//"USER":                           true, // ODBC
		"USING":                          true, // ODBC
		//"VALUES":                         true, // ODBC
		"VALUE":                          true, // ODBC
		"VARCHAR":                        true, // ODBC
		//"VARYING":                        true, // ODBC
		//"VIEW":                           true, // ODBC
		"WHENEVER":                       true, // ODBC
		//"WHEN":                           true, // ODBC
		//"WHERE":                          true, // ODBC
		//"WITH":                           true, // ODBC
		"WORK":                           true, // ODBC
		"WRITE":                          true, // ODBC
		//"YEAR":                           true, // ODBC
		"ZONE":                           true, // ODBC
	}

	v, ok := msAccessKeywords[strings.ToUpper(s)]

	return ok, v
}

// IsKeyword returns a boolean indicating if the supplied string
// is considered to be a keyword in MSAccess
func (d MSAccessDialect) IsKeyword(s string) bool {
	isKey, _ := d.keyword(s)
	return isKey
}

// IsReservedKeyword returns a boolean indicating if the supplied
// string is considered to be a reserved keyword in MSAccess
func (d MSAccessDialect) IsReservedKeyword(s string) bool {
	isKey, isReserved := d.keyword(s)

	if isKey {
		return isReserved
	}
	return false
}

// IsOperator returns a boolean indicating if the supplied string
// is considered to be an operator in MSAccess
func (d MSAccessDialect) IsOperator(s string) bool {

	var msAccessOperators = map[string]bool{
		"<":   true,
		"&":   true,
		"*":   true,
		"+":   true,
		"-":   true,
		"/":   true,
		"<=":  true,
		"<> ": true,
		"=":   true,
		">":   true,
		">=":  true,
		"\\":  true,
		"^":   true,
		"mod": true,
	}

	_, ok := msAccessOperators[s]
	return ok
}

// IsLabel returns a boolean indicating if the supplied string
// is considered to be a label in MSAccess
func (d MSAccessDialect) IsLabel(s string) bool {
	return false
}

// IsIdentifier returns a boolean indicating if the supplied
// string is considered to be a non-quoted MSAccess identifier.
func (d MSAccessDialect) IsIdentifier(s string) bool {

	/*

		From the documentation found:

		   The first character must be one of the following:

		       A letter as defined by the Unicode Standard 3.2. The Unicode
		       definition of letters includes Latin characters from a through
		       z, from A through Z, and also letter characters from other
		       languages.

		       The underscore (_), at sign (@), or number sign (#).

		   ...

		   Subsequent characters can include the following:
		       Letters as defined in the Unicode Standard 3.2.
		       Decimal numbers from either Basic Latin or other national scripts.
		       The at sign, dollar sign ($), number sign, or underscore.

		       Embedded spaces or special characters are not allowed.

		       Supplementary characters are not allowed.

	*/

	const firstIdentChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_#@"
	const identChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_#@$"

	chr := strings.Split(s, "")
	for i := 0; i < len(chr); i++ {

		if i == 0 {
			matches := strings.Contains(firstIdentChars, chr[i])
			if !matches {
				return false
			}

		} else {
			matches := strings.Contains(identChars, chr[i])
			if !matches && chr[i] != "." {
				return false
			}

		}
	}

	return true
}
