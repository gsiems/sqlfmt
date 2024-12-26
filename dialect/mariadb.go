package dialect

import (
	"regexp"
	"strings"
)

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
// (or string slice) is considered to be a datatype in MariaDB
func (d MariaDBDialect) IsDatatype(s ...string) bool {

	var mariadbDatatypes = map[string]bool{

		"bigint (n) signed":               true,
		"bigint (n)":                      true,
		"bigint (n) unsigned":             true,
		"bigint (n) zerofill":             true,
		"bigint signed":                   true,
		"bigint":                          true,
		"bigint unsigned":                 true,
		"bigint zerofill":                 true,
		"binary (n)":                      true,
		"binary":                          true,
		"bit (n)":                         true,
		"bit":                             true,
		"blob (n)":                        true,
		"blob":                            true,
		"boolean":                         true,
		"bool":                            true,
		"char byte (n)":                   true, // compatibility feature
		"char byte":                       true, // compatibility feature
		"char (n)":                        true,
		"char":                            true,
		"datetime":                        true,
		"date":                            true,
		"decimal (n,n) signed":            true,
		"decimal (n,n)":                   true,
		"decimal (n,n) unsigned":          true,
		"decimal (n,n) zerofill":          true,
		"decimal (n) signed":              true,
		"decimal (n)":                     true,
		"decimal (n) unsigned":            true,
		"decimal (n) zerofill":            true,
		"decimal signed":                  true,
		"decimal":                         true,
		"decimal unsigned":                true,
		"decimal zerofill":                true,
		"dec (n,n) signed":                true, // synonym for decimal
		"dec (n,n)":                       true, // synonym for decimal
		"dec (n,n) unsigned":              true, // synonym for decimal
		"dec (n,n) zerofill":              true, // synonym for decimal
		"dec (n) signed":                  true, // synonym for decimal
		"dec (n)":                         true, // synonym for decimal
		"dec (n) unsigned":                true, // synonym for decimal
		"dec (n) zerofill":                true, // synonym for decimal
		"dec signed":                      true, // synonym for decimal
		"dec":                             true, // synonym for decimal
		"dec unsigned":                    true, // synonym for decimal
		"dec zerofill":                    true, // synonym for decimal
		"double (n,n) signed":             true,
		"double (n,n)":                    true,
		"double (n,n) unsigned":           true,
		"double (n,n) zerofill":           true,
		"double precision (n,n) signed":   true,
		"double precision (n,n)":          true,
		"double precision (n,n) unsigned": true,
		"double precision (n,n) zerofill": true,
		"double precision signed":         true,
		"double precision":                true,
		"double precision unsigned":       true,
		"double precision zerofill":       true,
		"double signed":                   true,
		"double":                          true,
		"double unsigned":                 true,
		"double zerofill":                 true,
		"enum":                            true,
		"fixed (n,n) signed":              true, // other DB compatibility synonym for decimal
		"fixed (n,n)":                     true, // other DB compatibility synonym for decimal
		"fixed (n,n) unsigned":            true, // other DB compatibility synonym for decimal
		"fixed (n,n) zerofill":            true, // other DB compatibility synonym for decimal
		"fixed (n) signed":                true, // other DB compatibility synonym for decimal
		"fixed (n)":                       true, // other DB compatibility synonym for decimal
		"fixed (n) unsigned":              true, // other DB compatibility synonym for decimal
		"fixed (n) zerofill":              true, // other DB compatibility synonym for decimal
		"fixed signed":                    true, // other DB compatibility synonym for decimal
		"fixed":                           true, // other DB compatibility synonym for decimal
		"fixed unsigned":                  true, // other DB compatibility synonym for decimal
		"fixed zerofill":                  true, // other DB compatibility synonym for decimal
		"float (n,n) signed":              true,
		"float (n,n)":                     true,
		"float (n,n) unsigned":            true,
		"float (n,n) zerofill":            true,
		"float signed":                    true,
		"float":                           true,
		"float unsigned":                  true,
		"float zerofill":                  true,
		"integer (n) signed":              true,
		"integer (n)":                     true,
		"integer (n) unsigned":            true,
		"integer (n) zerofill":            true,
		"integer signed":                  true,
		"integer":                         true,
		"integer unsigned":                true,
		"integer zerofill":                true,
		"int (n) signed":                  true,
		"int (n)":                         true,
		"int (n) unsigned":                true,
		"int (n) zerofill":                true,
		"int signed":                      true,
		"int":                             true,
		"int unsigned":                    true,
		"int zerofill":                    true,
		"longblob":                        true,
		"longtext":                        true,
		"mediumblob":                      true,
		"mediumint (n) signed":            true,
		"mediumint (n)":                   true,
		"mediumint (n) unsigned":          true,
		"mediumint (n) zerofill":          true,
		"mediumint signed":                true,
		"mediumint":                       true,
		"mediumint unsigned":              true,
		"mediumint zerofill":              true,
		"mediumtext":                      true,
		"national char (n)":               true,
		"national char":                   true,
		"national varchar (n)":            true,
		"national varchar":                true,
		"number (n,n) signed":             true, // Oracle mode synonym for decimal
		"number (n,n)":                    true, // Oracle mode synonym for decimal
		"number (n,n) unsigned":           true, // Oracle mode synonym for decimal
		"number (n,n) zerofill":           true, // Oracle mode synonym for decimal
		"number (n) signed":               true, // Oracle mode synonym for decimal
		"number (n)":                      true, // Oracle mode synonym for decimal
		"number (n) unsigned":             true, // Oracle mode synonym for decimal
		"number (n) zerofill":             true, // Oracle mode synonym for decimal
		"number signed":                   true, // Oracle mode synonym for decimal
		"number":                          true, // Oracle mode synonym for decimal
		"number unsigned":                 true, // Oracle mode synonym for decimal
		"number zerofill":                 true, // Oracle mode synonym for decimal
		"numeric (n,n) signed":            true, // synonym for decimal
		"numeric (n,n)":                   true, // synonym for decimal
		"numeric (n,n) unsigned":          true, // synonym for decimal
		"numeric (n,n) zerofill":          true, // synonym for decimal
		"numeric (n) signed":              true, // synonym for decimal
		"numeric (n)":                     true, // synonym for decimal
		"numeric (n) unsigned":            true, // synonym for decimal
		"numeric (n) zerofill":            true, // synonym for decimal
		"numeric signed":                  true, // synonym for decimal
		"numeric":                         true, // synonym for decimal
		"numeric unsigned":                true, // synonym for decimal
		"numeric zerofill":                true, // synonym for decimal
		"real (n,n) signed":               true,
		"real (n,n)":                      true,
		"real (n,n) unsigned":             true,
		"real (n,n) zerofill":             true,
		"real signed":                     true,
		"real":                            true,
		"real unsigned":                   true,
		"real zerofill":                   true,
		"set":                             true,
		"smallint (n) signed":             true,
		"smallint (n)":                    true,
		"smallint (n) unsigned":           true,
		"smallint (n) zerofill":           true,
		"smallint signed":                 true,
		"smallint":                        true,
		"smallint unsigned":               true,
		"smallint zerofill":               true,
		"text (n)":                        true,
		"text":                            true,
		"timestamp":                       true,
		"time":                            true,
		"tinyblob":                        true,
		"tinyint (n) signed":              true,
		"tinyint (n)":                     true,
		"tinyint (n) unsigned":            true,
		"tinyint (n) zerofill":            true,
		"tinyint signed":                  true,
		"tinyint":                         true,
		"tinyint unsigned":                true,
		"tinyint zerofill":                true,
		"tinytext":                        true,
		"varbinary (n)":                   true,
		"varbinary":                       true,
		"varchar (n)":                     true,
		"varchar":                         true,
		"vector (n)":                      true,
		"vector":                          true,
		"year":                            true,
	}

	var z []string
	rn := regexp.MustCompile(`^[0-9]+$`)

	for i, v := range s {
		switch v {
		case "(":
			z = append(z, " "+v)
		case ")", ",":
			z = append(z, v)
		default:
			switch {
			case rn.MatchString(v):
				z = append(z, "n")
			case i == 0:
				z = append(z, v)
			default:
				z = append(z, " "+v)
			}
		}
	}

	k := strings.ToLower(strings.Join(z, ""))
	if _, ok := mariadbDatatypes[k]; ok {
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
