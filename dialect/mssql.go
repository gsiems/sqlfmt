package dialect

import "strings"

type MSSQLDialect struct {
	dialect int
	name    string
}

func NewMSSQLDialect() *MSSQLDialect {
	var d MSSQLDialect

	d.dialect = MSSQL
	d.name = "MSSQL"

	return &d
}

func (d MSSQLDialect) Dialect() int {
	return d.dialect
}
func (d MSSQLDialect) DialectName() string {
	return d.name
}
func (d MSSQLDialect) CaseFolding() int {
	return NoFolding
}
func (d MSSQLDialect) IdentQuoteChar() string {
	return "\""
}
func (d MSSQLDialect) StringQuoteChar() string {
	return "'"
}

// MaxOperatorLength returns the length of the longest operator
// supported by the dialect
func (d MSSQLDialect) MaxOperatorLength() int {
	return 2
}

// IsDatatype returns a boolean indicating if the supplied string
// is considered to be a datatype in MSSQL
func (d MSSQLDialect) IsDatatype(s string) bool {

	var mssqlDatatypes = map[string]bool{
		"bigint":           true,
		"binary":           true,
		"bit":              true,
		"char":             true,
		"cursor":           true,
		"datetime2":        true,
		"datetimeoffset":   true,
		"datetime":         true,
		"date":             true,
		"decimal":          true,
		"float":            true,
		"image":            true,
		"int":              true,
		"money":            true,
		"nchar":            true,
		"ntext":            true,
		"numeric":          true,
		"nvarchar":         true,
		"real":             true,
		"smalldatetime":    true,
		"smallint":         true,
		"smallmoney":       true,
		"sql_variant":      true,
		"table":            true,
		"text":             true,
		"timestamp":        true,
		"time":             true,
		"tinyint":          true,
		"uniqueidentifier": true,
		"varbinary":        true,
		"varchar":          true,
		"xml":              true,
	}

	/*
	   MS Access Data Types
	   Data type 	Description 	Storage
	   Text 	Use for text or combinations of text and numbers. 255 characters maximum
	   Memo 	Memo is used for larger amounts of text. Stores up to 65,536 characters. Note: You cannot sort a memo field. However, they are searchable
	   Byte 	Allows whole numbers from 0 to 255 	1 byte
	   Integer 	Allows whole numbers between -32,768 and 32,767 	2 bytes
	   Long 	Allows whole numbers between -2,147,483,648 and 2,147,483,647 	4 bytes
	   Single 	Single precision floating-point. Will handle most decimals 	4 bytes
	   Double 	Double precision floating-point. Will handle most decimals 	8 bytes
	   Currency 	Use for currency. Holds up to 15 digits of whole dollars, plus 4 decimal places. Tip: You can choose which country's currency to use 	8 bytes
	   AutoNumber 	AutoNumber fields automatically give each record its own number, usually starting at 1 	4 bytes
	   Date/Time 	Use for dates and times 	8 bytes
	   Yes/No 	A logical field can be displayed as Yes/No, True/False, or On/Off. In code, use the constants True and False (equivalent to -1 and 0). Note: Null values are not allowed in Yes/No fields 	1 bit
	   Ole Object 	Can store pictures, audio, video, or other BLOBs (Binary Large Objects) 	up to 1GB
	   Hyperlink 	Contain links to other files, including web pages
	   Lookup Wizard 	Let you type a list of options, which can then be chosen from a drop-down list 	4 bytes

	*/

	k := strings.ToLower(s)
	if _, ok := mssqlDatatypes[k]; ok {
		return true
	}
	return false
}

func (d MSSQLDialect) keyword(s string) (bool, bool) {

	/*
	   Microsoft SQL-Server keywords

	   https://docs.microsoft.com/en-us/sql/t-sql/language-elements/reserved-keywords-transact-sql?view=sql-server-ver15

	   The isReserved value is set to false as there is no indication (from
	   the above link) if the keywords are reserved or not.

	*/

	// map[keyword]isReserved
	var mssqlKeywords = map[string]bool{
		"ADD":                            false,
		"ALL":                            false,
		"ALTER":                          false,
		"AND":                            false,
		"ANY":                            false,
		"AS":                             false,
		"ASC":                            false,
		"AUTHORIZATION":                  false,
		"BACKUP":                         false,
		"BEGIN":                          false,
		"BETWEEN":                        false,
		"BREAK":                          false,
		"BROWSE":                         false,
		"BULK":                           false,
		"BY":                             false,
		"CASCADE":                        false,
		"CASE":                           false,
		"CHECK":                          false,
		"CHECKPOINT":                     false,
		"CLOSE":                          false,
		"CLUSTERED":                      false,
		"COALESCE":                       false,
		"COLLATE":                        false,
		"COLUMN":                         false,
		"COMMIT":                         false,
		"COMPUTE":                        false,
		"CONSTRAINT":                     false,
		"CONTAINS":                       false,
		"CONTAINSTABLE":                  false,
		"CONTINUE":                       false,
		"CONVERT":                        false,
		"CREATE":                         false,
		"CROSS":                          false,
		"CURRENT":                        false,
		"CURRENT_DATE":                   false,
		"CURRENT_TIME":                   false,
		"CURRENT_TIMESTAMP":              false,
		"CURRENT_USER":                   false,
		"CURSOR":                         false,
		"DATABASE":                       false,
		"DBCC":                           false,
		"DEALLOCATE":                     false,
		"DECLARE":                        false,
		"DEFAULT":                        false,
		"DELETE":                         false,
		"DENY":                           false,
		"DESC":                           false,
		"DISK":                           false,
		"DISTINCT":                       false,
		"DISTRIBUTED":                    false,
		"DOUBLE":                         false,
		"DROP":                           false,
		"DUMP":                           false,
		"ELSE":                           false,
		"END":                            false,
		"ERRLVL":                         false,
		"ESCAPE":                         false,
		"EXCEPT":                         false,
		"EXEC":                           false,
		"EXECUTE":                        false,
		"EXISTS":                         false,
		"EXIT":                           false,
		"EXTERNAL":                       false,
		"FETCH":                          false,
		"FILE":                           false,
		"FILLFACTOR":                     false,
		"FOR":                            false,
		"FOREIGN":                        false,
		"FREETEXT":                       false,
		"FREETEXTTABLE":                  false,
		"FROM":                           false,
		"FULL":                           false,
		"FUNCTION":                       false,
		"GOTO":                           false,
		"GRANT":                          false,
		"GROUP":                          false,
		"HAVING":                         false,
		"HOLDLOCK":                       false,
		"IDENTITY":                       false,
		"IDENTITYCOL":                    false,
		"IDENTITY_INSERT":                false,
		"IF":                             false,
		"IN":                             false,
		"INDEX":                          false,
		"INNER":                          false,
		"INSERT":                         false,
		"INTERSECT":                      false,
		"INTO":                           false,
		"IS":                             false,
		"JOIN":                           false,
		"KEY":                            false,
		"KILL":                           false,
		"LABEL":                          false,
		"LEFT":                           false,
		"LIKE":                           false,
		"LINENO":                         false,
		"LOAD":                           false,
		"MERGE":                          false,
		"NATIONAL":                       false,
		"NOCHECK":                        false,
		"NONCLUSTERED":                   false,
		"NOT":                            false,
		"NULL":                           false,
		"NULLIF":                         false,
		"OF":                             false,
		"OFF":                            false,
		"OFFSETS":                        false,
		"ON":                             false,
		"OPEN":                           false,
		"OPENDATASOURCE":                 false,
		"OPENQUERY":                      false,
		"OPENROWSET":                     false,
		"OPENXML":                        false,
		"OPTION":                         false,
		"OR":                             false,
		"ORDER":                          false,
		"OUTER":                          false,
		"OVER":                           false,
		"PERCENT":                        false,
		"PIVOT":                          false,
		"PLAN":                           false,
		"PRECISION":                      false,
		"PRIMARY":                        false,
		"PRINT":                          false,
		"PROC":                           false,
		"PROCEDURE":                      false,
		"PUBLIC":                         false,
		"RAISERROR":                      false,
		"READ":                           false,
		"READTEXT":                       false,
		"RECONFIGURE":                    false,
		"REFERENCES":                     false,
		"REPLICATION":                    false,
		"RESTORE":                        false,
		"RESTRICT":                       false,
		"RETURN":                         false,
		"REVERT":                         false,
		"REVOKE":                         false,
		"RIGHT":                          false,
		"ROLLBACK":                       false,
		"ROWCOUNT":                       false,
		"ROWGUIDCOL":                     false,
		"RULE":                           false,
		"SAVE":                           false,
		"SCHEMA":                         false,
		"SECURITYAUDIT":                  false,
		"SELECT":                         false,
		"SEMANTICKEYPHRASETABLE":         false,
		"SEMANTICSIMILARITYDETAILSTABLE": false,
		"SEMANTICSIMILARITYTABLE":        false,
		"SESSION_USER":                   false,
		"SET":                            false,
		"SETUSER":                        false,
		"SHUTDOWN":                       false,
		"SOME":                           false,
		"STATISTICS":                     false,
		"SYSTEM_USER":                    false,
		"TABLE":                          false,
		"TABLESAMPLE":                    false,
		"TEXTSIZE":                       false,
		"THEN":                           false,
		"TO":                             false,
		"TOP":                            false,
		"TRAN":                           false,
		"TRANSACTION":                    false,
		"TRIGGER":                        false,
		"TRUNCATE":                       false,
		"TRY_CONVERT":                    false,
		"TSEQUAL":                        false,
		"UNION":                          false,
		"UNIQUE":                         false,
		"UNPIVOT":                        false,
		"UPDATE":                         false,
		"UPDATETEXT":                     false,
		"USE":                            false,
		"USER":                           false,
		"VALUES":                         false,
		"VARYING":                        false,
		"VIEW":                           false,
		"WAITFOR":                        false,
		"WHEN":                           false,
		"WHERE":                          false,
		"WHILE":                          false,
		"WITH":                           false,
		"WITHIN GROUP":                   false,
		"WRITETEXT":                      false,
	}

	v, ok := mssqlKeywords[strings.ToUpper(s)]

	return ok, v
}

// IsKeyword returns a boolean indicating if the supplied string
// is considered to be a keyword in MSSQL
func (d MSSQLDialect) IsKeyword(s string) bool {
	isKey, _ := d.keyword(s)
	return isKey
}

// IsReservedKeyword returns a boolean indicating if the supplied
// string is considered to be a reserved keyword in MSSQL
func (d MSSQLDialect) IsReservedKeyword(s string) bool {
	isKey, isReserved := d.keyword(s)

	if isKey {
		return isReserved
	}
	return false
}

// IsOperator returns a boolean indicating if the supplied string
// is considered to be an operator in MSSQL
func (d MSSQLDialect) IsOperator(s string) bool {

	var mssqlOperators = map[string]bool{
		"^":  true,
		"^=": true,
		"~":  true,
		"<":  true,
		"<=": true,
		"<>": true,
		"=":  true,
		">":  true,
		">=": true,
		"|":  true,
		"|=": true,
		"-":  true,
		"-=": true,
		"::": true,
		"!<": true,
		"!=": true,
		"!>": true,
		"/":  true,
		"/=": true,
		"*":  true,
		"*=": true,
		"&":  true,
		"&=": true,
		"%":  true,
		"%=": true,
		"+":  true,
		"+=": true,
	}

	_, ok := mssqlOperators[s]
	return ok
}

// IsLabel returns a boolean indicating if the supplied string
// is considered to be a label in MSSQL
func (d MSSQLDialect) IsLabel(s string) bool {
	if len(s) < 2 {
		return false
	}
	if string(s[len(s)-1]) != ":" {
		return false
	}
	if !d.IsIdentifier(s[0:len(s)-2]) {
		return false
	}
	return true
}

// IsIdentifier returns a boolean indicating if the supplied
// string is considered to be a non-quoted MSSQL identifier.
func (d MSSQLDialect) IsIdentifier(s string) bool {

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
