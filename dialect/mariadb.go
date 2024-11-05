package dialect

import "strings"

type MariaDBDialect struct {
	dialect int
	name    string
}

func NewMariaDBDialect() *MariaDBDialect {
	var d MariaDBDialect

	d.dialect = MariaDB
	d.name = "MariaDB"

	return &d
}

func (d MariaDBDialect) Dialect() int {
	return d.dialect
}
func (d MariaDBDialect) DialectName() string {
	return d.name
}
func (d MariaDBDialect) CaseFolding() int {
	return NoFolding
}
func (d MariaDBDialect) IdentQuoteChar() string {
	return "\""
}
func (d MariaDBDialect) StringQuoteChar() string {
	return "'"
}

// MaxOperatorLength returns the length of the longest operator
// supported by the dialect
func (d MariaDBDialect) MaxOperatorLength() int {
	return 3
}

// IsDatatype returns a boolean indicating if the supplied string
// is considered to be a datatype in MariaDB
func (d MariaDBDialect) IsDatatype(s string) bool {

	var mariadbDatatypes = map[string]bool{
		"bigint":           true,
		"binary":           true,
		"bit":              true,
		"blob":             true,
		"boolean":          true,
		"bool":             true,
		"char":             true,
		"datetime":         true,
		"date":             true,
		"decimal":          true,
		"dec":              true,
		"double precision": true,
		"double":           true,
		"enum":             true,
		"float":            true,
		"integer":          true,
		"int":              true,
		"longblob":         true,
		"longtext":         true,
		"mediumblob":       true,
		"mediumint":        true,
		"mediumtext":       true,
		"set":              true,
		"smallint":         true,
		"text":             true,
		"timestamp":        true,
		"time":             true,
		"tinyblob":         true,
		"tinyint":          true,
		"tinytext":         true,
		"varbinary":        true,
		"varchar":          true,
		"year":             true,
	}

	if _, ok := mariadbDatatypes[strings.ToLower(s)]; ok {
		return true
	}

	return false
}

func (d MariaDBDialect) keyword(s string) (bool, bool) {

	/*
	   MariaDB keywords

	   https://mariadb.com/kb/en/library/reserved-words/

	    * Some keywords are exceptions for historical reasons, and are permitted as unquoted identifiers.
	    * In Oracle mode, from MariaDB 10.3, there are a number of extra reserved words.

	   The isReserved value is set to false as there is no indication (from
	   the above link) if the keywords are reserved or not.

	*/

	// map[keyword]isReserved
	var mariadbKeywords = map[string]bool{
		"ACCESSIBLE":                    false,
		"ACTION":                        false,
		"ADD":                           false,
		"ALL":                           false,
		"ALTER":                         false,
		"ANALYZE":                       false,
		"AND":                           false,
		"ASC":                           false,
		"ASENSITIVE":                    false,
		"AS":                            false,
		"BEFORE":                        false,
		"BETWEEN":                       false,
		"BIGINT":                        false,
		"BINARY":                        false,
		"BIT":                           false,
		"BLOB":                          false,
		"BODY":                          false,
		"BOTH":                          false,
		"BY":                            false,
		"CALL":                          false,
		"CASCADE":                       false,
		"CASE":                          false,
		"CHANGE":                        false,
		"CHARACTER":                     false,
		"CHAR":                          false,
		"CHECK":                         false,
		"COLLATE":                       false,
		"COLUMN":                        false,
		"CONDITION":                     false,
		"CONSTRAINT":                    false,
		"CONTINUE":                      false,
		"CONVERT":                       false,
		"CREATE":                        false,
		"CROSS":                         false,
		"CURRENT_DATE":                  false,
		"CURRENT_ROLE":                  false,
		"CURRENT_TIME":                  false,
		"CURRENT_TIMESTAMP":             false,
		"CURRENT_USER":                  false,
		"CURSOR":                        false,
		"DATABASE":                      false,
		"DATABASES":                     false,
		"DATE":                          false,
		"DAY_HOUR":                      false,
		"DAY_MICROSECOND":               false,
		"DAY_MINUTE":                    false,
		"DAY_SECOND":                    false,
		"DEC":                           false,
		"DECIMAL":                       false,
		"DECLARE":                       false,
		"DEFAULT":                       false,
		"DELAYED":                       false,
		"DELETE":                        false,
		"DESC":                          false,
		"DESCRIBE":                      false,
		"DETERMINISTIC":                 false,
		"DISTINCT":                      false,
		"DISTINCTROW":                   false,
		"DIV":                           false,
		"DO_DOMAIN_IDS":                 false,
		"DOUBLE":                        false,
		"DROP":                          false,
		"DUAL":                          false,
		"EACH":                          false,
		"ELSE":                          false,
		"ELSEIF":                        false,
		"ELSIF":                         false,
		"ENCLOSED":                      false,
		"ENUM":                          false,
		"ESCAPED":                       false,
		"EXCEPT":                        false,
		"EXISTS":                        false,
		"EXIT":                          false,
		"EXPLAIN":                       false,
		"FALSE":                         false,
		"FETCH":                         false,
		"FLOAT4":                        false,
		"FLOAT8":                        false,
		"FLOAT":                         false,
		"FORCE":                         false,
		"FOREIGN":                       false,
		"FOR":                           false,
		"FROM":                          false,
		"FULLTEXT":                      false,
		"GENERAL":                       false,
		"GOTO":                          false,
		"GRANT":                         false,
		"GROUP":                         false,
		"HAVING":                        false,
		"HIGH_PRIORITY":                 false,
		"HISTORY":                       false,
		"HOUR_MICROSECOND":              false,
		"HOUR_MINUTE":                   false,
		"HOUR_SECOND":                   false,
		"IF":                            false,
		"IGNORE_DOMAIN_IDS":             false,
		"IGNORE":                        false,
		"IGNORE_SERVER_IDS":             false,
		"INDEX":                         false,
		"IN":                            false,
		"INFILE":                        false,
		"INNER":                         false,
		"INOUT":                         false,
		"INSENSITIVE":                   false,
		"INSERT":                        false,
		"INT1":                          false,
		"INT2":                          false,
		"INT3":                          false,
		"INT4":                          false,
		"INT8":                          false,
		"INTEGER":                       false,
		"INTERSECT":                     false,
		"INTERVAL":                      false,
		"INT":                           false,
		"INTO":                          false,
		"IS":                            false,
		"ITERATE":                       false,
		"JOIN":                          false,
		"KEY":                           false,
		"KEYS":                          false,
		"KILL":                          false,
		"LEADING":                       false,
		"LEAVE":                         false,
		"LEFT":                          false,
		"LIKE":                          false,
		"LIMIT":                         false,
		"LINEAR":                        false,
		"LINES":                         false,
		"LOAD":                          false,
		"LOCALTIME":                     false,
		"LOCALTIMESTAMP":                false,
		"LOCK":                          false,
		"LONGBLOB":                      false,
		"LONG":                          false,
		"LONGTEXT":                      false,
		"LOOP":                          false,
		"LOW_PRIORITY":                  false,
		"MASTER_HEARTBEAT_PERIOD":       false,
		"MASTER_SSL_VERIFY_SERVER_CERT": false,
		"MATCH":                         false,
		"MAXVALUE":                      false,
		"MEDIUMBLOB":                    false,
		"MEDIUMINT":                     false,
		"MEDIUMTEXT":                    false,
		"MIDDLEINT":                     false,
		"MINUTE_MICROSECOND":            false,
		"MINUTE_SECOND":                 false,
		"MOD":                           false,
		"MODIFIES":                      false,
		"NATURAL":                       false,
		"NO":                            false,
		"NOT":                           false,
		"NO_WRITE_TO_BINLOG":            false,
		"NULL":                          false,
		"NUMERIC":                       false,
		"ON":                            false,
		"OPTIMIZE":                      false,
		"OPTIONALLY":                    false,
		"OPTION":                        false,
		"ORDER":                         false,
		"OR":                            false,
		"OTHERS":                        false,
		"OUTER":                         false,
		"OUT":                           false,
		"OUTFILE":                       false,
		"OVER":                          false,
		"PACKAGE":                       false,
		"PAGE_CHECKSUM":                 false,
		"PARSE_VCOL_EXPR":               false,
		"PARTITION":                     false,
		"PERIOD":                        false,
		"PRECISION":                     false,
		"PRIMARY":                       false,
		"PROCEDURE":                     false,
		"PURGE":                         false,
		"RAISE":                         false,
		"RANGE":                         false,
		"READ":                          false,
		"READS":                         false,
		"READ_WRITE":                    false,
		"REAL":                          false,
		"RECURSIVE":                     false,
		"REFERENCES":                    false,
		"REF_SYSTEM_ID":                 false,
		"REGEXP":                        false,
		"RELEASE":                       false,
		"RENAME":                        false,
		"REPEAT":                        false,
		"REPLACE":                       false,
		"REQUIRE":                       false,
		"RESIGNAL":                      false,
		"RESTRICT":                      false,
		"RETURN":                        false,
		"RETURNING":                     false,
		"REVOKE":                        false,
		"RIGHT":                         false,
		"RLIKE":                         false,
		"ROWS":                          false,
		"ROWTYPE":                       false,
		"SCHEMA":                        false,
		"SCHEMAS":                       false,
		"SECOND_MICROSECOND":            false,
		"SELECT":                        false,
		"SENSITIVE":                     false,
		"SEPARATOR":                     false,
		"SET":                           false,
		"SHOW":                          false,
		"SIGNAL":                        false,
		"SLOW":                          false,
		"SMALLINT":                      false,
		"SPATIAL":                       false,
		"SPECIFIC":                      false,
		"SQL_BIG_RESULT":                false,
		"SQL_CALC_FOUND_ROWS":           false,
		"SQLEXCEPTION":                  false,
		"SQL":                           false,
		"SQL_SMALL_RESULT":              false,
		"SQLSTATE":                      false,
		"SQLWARNING":                    false,
		"SSL":                           false,
		"STARTING":                      false,
		"STATS_AUTO_RECALC":             false,
		"STATS_PERSISTENT":              false,
		"STATS_SAMPLE_PAGES":            false,
		"STRAIGHT_JOIN":                 false,
		"SYSTEM":                        false,
		"SYSTEM_TIME":                   false,
		"TABLE":                         false,
		"TERMINATED":                    false,
		"TEXT":                          false,
		"THEN":                          false,
		"TIME":                          false,
		"TIMESTAMP":                     false,
		"TINYBLOB":                      false,
		"TINYINT":                       false,
		"TINYTEXT":                      false,
		"TO":                            false,
		"TRAILING":                      false,
		"TRIGGER":                       false,
		"TRUE":                          false,
		"UNDO":                          false,
		"UNION":                         false,
		"UNIQUE":                        false,
		"UNLOCK":                        false,
		"UNSIGNED":                      false,
		"UPDATE":                        false,
		"USAGE":                         false,
		"USE":                           false,
		"USING":                         false,
		"UTC_DATE":                      false,
		"UTC_TIME":                      false,
		"UTC_TIMESTAMP":                 false,
		"VALUES":                        false,
		"VARBINARY":                     false,
		"VARCHARACTER":                  false,
		"VARCHAR":                       false,
		"VARYING":                       false,
		"VERSIONING":                    false,
		"WHEN":                          false,
		"WHERE":                         false,
		"WHILE":                         false,
		"WINDOW":                        false,
		"WITH":                          false,
		"WITHOUT":                       false,
		"WRITE":                         false,
		"XOR":                           false,
		"YEAR_MONTH":                    false,
		"ZEROFILL":                      false,
	}

	v, ok := mariadbKeywords[strings.ToUpper(s)]

	return ok, v
}

// IsKeyword returns a boolean indicating if the supplied string
// is considered to be a keyword in MariaDB
func (d MariaDBDialect) IsKeyword(s string) bool {
	isKey, _ := d.keyword(s)
	return isKey
}

// IsReservedKeyword returns a boolean indicating if the supplied
// string is considered to be a reserved keyword in MariaDB
func (d MariaDBDialect) IsReservedKeyword(s string) bool {
	isKey, isReserved := d.keyword(s)

	if isKey {
		return isReserved
	}
	return false
}

// IsOperator returns a boolean indicating if the supplied string
// is considered to be an operator in MariaDB
func (d MariaDBDialect) IsOperator(s string) bool {

	var mariadbOperators = map[string]bool{
		"<":   true,
		"<=":  true,
		"<=>": true,
		"=":   true,
		">":   true,
		">=":  true,
		"||":  true,
		"-":   true,
		":=":  true,
		"!":   true,
		"!=":  true,
		"/":   true,
		"*":   true,
		"&&":  true,
		"%":   true,
		"+":   true,
	}

	_, ok := mariadbOperators[strings.ToUpper(s)]
	return ok
}

// IsLabel returns a boolean indicating if the supplied string
// is considered to be a label in MariaDB
func (d MariaDBDialect) IsLabel(s string) bool {
	if len(s) < 2 {
		return false
	}
	if string(s[len(s)-1]) != ":" {
		return false
	}
	if !d.IsIdentifier(s[0 : len(s)-2]) {
		return false
	}
	return true
}

// IsIdentifier returns a boolean indicating if the supplied
// string is considered to be a non-quoted MariaDB identifier.
func (d MariaDBDialect) IsIdentifier(s string) bool {

	/*

	   From the documentation:

	   The following characters are valid, and allow identifiers to be unquoted:

	       ASCII: [0-9,a-z,A-Z$_] (numerals 0-9, basic Latin letters, both lowercase and uppercase, dollar sign, underscore)
	       Extended: U+0080 .. U+FFFF


	      * Identifiers are stored as Unicode (UTF-8)
	      * Identifiers may or may not be case-sensitive. See Indentifier Case-sensitivity.
	      * Database, table and column names can't end with space characters
	      * Identifier names may begin with a numeral, but can't only contain numerals unless quoted.
	      * An identifier starting with a numeral, followed by an 'e', may be parsed as a floating point number, and needs to be quoted.
	      * Identifiers are not permitted to contain the ASCII NUL character (U+0000) and supplementary characters (U+10000 and higher).
	      * Names such as 5e6, 9e are not prohibited, but it's strongly recommended not to use them, as they could lead to ambiguity in certain contexts, being treated as a number or expression.
	      * User variables cannot be used as part of an identifier, or as an identifier in an SQL statem
	*/

	const identChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$"
	const digitChars = "0123456789"

	allDigits := true

	chr := strings.Split(s, "")
	for i := 0; i < len(chr); i++ {

		matches := strings.Contains(identChars, chr[i])
		if !matches && chr[i] != "." {
			return false
		}

		if !strings.Contains(digitChars, chr[i]) {
			allDigits = false
		}
	}

	if allDigits {
		return false
	}

	return true
}
