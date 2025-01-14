package dialect

import (
	"regexp"
	"strings"
)

type StandardSQLDialect struct {
	dialect int
	name    string
}

func NewStandardSQLDialect() *StandardSQLDialect {
	var d StandardSQLDialect

	d.dialect = StandardSQL
	d.name = "StandardSQL"

	return &d
}

func (d StandardSQLDialect) Dialect() int {
	return d.dialect
}
func (d StandardSQLDialect) DialectName() string {
	return d.name
}
func (d StandardSQLDialect) CaseFolding() int {
	return FoldUpper
}
func (d StandardSQLDialect) IdentQuoteChar() string {
	return "\""
}
func (d StandardSQLDialect) StringQuoteChar() string {
	return "'"
}

// MaxOperatorLength returns the length of the longest operator
// supported by the dialect
func (d StandardSQLDialect) MaxOperatorLength() int {
	return 2
}

// IsDatatype returns a boolean indicating if the supplied string
// (or string slice) is considered to be a datatype in ISO Standared SQL
func (d StandardSQLDialect) IsDatatype(s ...string) bool {

	sqlStandardDatatypes := map[string]bool{
		"bigint":                       true,
		"binary large object":          true,
		"binary":                       true,
		"binary varying":               true,
		"bit":                          true,
		"bit varying":                  true,
		"bit varying (n)":              true,
		"boolean":                      true,
		"character large object":       true,
		"clob":                         true,
		"character":                    true,
		"character (n)":                true,
		"character varying":            true,
		"character varying (n)":        true,
		"char":                         true,
		"char (n)":                     true,
		"date":                         true,
		"decimal":                      true,
		"decimal (n)":                  true,
		"decimal (n,n)":                true,
		"double precision":             true,
		"float":                        true,
		"float (n)":                    true,
		"float (n,n)":                  true,
		"integer":                      true,
		"int":                          true,
		"interval":                     true,
		"interval day to second":       true, // expanded fields
		"interval year to month":       true, // expanded fields
		"national character":           true,
		"national character varying":   true,
		"nclob":                        true,
		"nchar":                        true,
		"nchar varying":                true,
		"numeric":                      true,
		"numeric (n)":                  true,
		"numeric (n,n)":                true,
		"real":                         true,
		"smallint":                     true,
		"timestamp":                    true,
		"timestamp (n)":                true,
		"timestamp with time zone":     true,
		"timestamp (n) with time zone": true,
		"time":                         true,
		"time (n)":                     true,
		"time with time zone":          true,
		"time (n) with time zone":      true,
		"tinyint":                      true,
		"varchar":                      true,
		"varchar (n)":                  true,
		"xml":                          true,
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
	if _, ok := sqlStandardDatatypes[k]; ok {
		return true
	}

	return false
}

// IsDatatypePart returns a boolean indicating if the supplied string
// is considered to be part of a datatype definition
func (d StandardSQLDialect) IsDatatypePart(s string) bool {

	switch strings.ToLower(s) {
	case "bigint", "binary", "bit", "boolean", "char", "character", "clob",
		"date", "day", "decimal", "double", "float", "int", "integer",
		"interval", "large", "month", "national", "nchar", "nclob", "numeric",
		"object", "precision", "real", "second", "smallint", "time",
		"timestamp", "tinyint", "to", "varchar", "varying", "with", "xml",
		"year", "zone":

		return true
	}

	return false
}

func (d StandardSQLDialect) keyword(s string) (bool, bool) {

	/*
	   Keywords in the SQL Standard

	       https://www.postgresql.org/docs/current/sql-keywords-appendix.html

	       Haven't found any other better list

	*/

	sqlStandardKeywords := map[string]bool{
		//"A": false,
		//"C": false,
		//"G": false,
		//"K": false,
		//"M": false,
		//"P": false,
		//"T": false,
		"ABS":                              true,
		"ABSENT":                           false,
		"ABSOLUTE":                         true,
		"ACCORDING":                        false,
		"ACOS":                             true,
		"ACTION":                           true,
		"ADA":                              false,
		"ADD":                              true,
		"ADMIN":                            false,
		"AFTER":                            false,
		"ALL":                              true,
		"ALLOCATE":                         true,
		"ALTER":                            true,
		"ALWAYS":                           false,
		"AND":                              true,
		"ANY":                              true,
		"ARE":                              true,
		"ARRAY":                            true,
		"ARRAY_AGG":                        true,
		"ARRAY_MAX_CARDINALITY":            true,
		"AS":                               true,
		"ASC":                              true,
		"ASENSITIVE":                       true,
		"ASIN":                             true,
		"ASSERTION":                        true,
		"ASSIGNMENT":                       false,
		"ASYMMETRIC":                       true,
		"AT":                               true,
		"ATAN":                             true,
		"ATOMIC":                           true,
		"ATTRIBUTE":                        false,
		"ATTRIBUTES":                       false,
		"AUTHORIZATION":                    true,
		"AVG":                              true,
		"BASE64":                           false,
		"BEFORE":                           false,
		"BEGIN":                            true,
		"BEGIN_FRAME":                      true,
		"BEGIN_PARTITION":                  true,
		"BERNOULLI":                        false,
		"BETWEEN":                          true,
		"BIGINT":                           true,
		"BINARY":                           true,
		"BIT":                              true,
		"BIT_LENGTH":                       true,
		"BLOB":                             true,
		"BLOCKED":                          false,
		"BOM":                              false,
		"BOOLEAN":                          true,
		"BOTH":                             true,
		"BREADTH":                          false,
		"BY":                               true,
		"CALL":                             true,
		"CALLED":                           true,
		"CARDINALITY":                      true,
		"CASCADE":                          true,
		"CASCADED":                         true,
		"CASE":                             true,
		"CAST":                             true,
		"CATALOG":                          true,
		"CATALOG_NAME":                     false,
		"CEIL":                             true,
		"CEILING":                          true,
		"CHAIN":                            false,
		"CHAINING":                         false,
		"CHAR":                             true,
		"CHAR_LENGTH":                      true,
		"CHARACTER":                        true,
		"CHARACTER_LENGTH":                 true,
		"CHARACTER_SET_CATALOG":            false,
		"CHARACTER_SET_NAME":               false,
		"CHARACTER_SET_SCHEMA":             false,
		"CHARACTERISTICS":                  false,
		"CHARACTERS":                       false,
		"CHECK":                            true,
		"CLASS_ORIGIN":                     false,
		"CLASSIFIER":                       true,
		"CLOB":                             true,
		"CLOSE":                            true,
		"COALESCE":                         true,
		"COBOL":                            false,
		"COLLATE":                          true,
		"COLLATION":                        true,
		"COLLATION_CATALOG":                false,
		"COLLATION_NAME":                   false,
		"COLLATION_SCHEMA":                 false,
		"COLLECT":                          true,
		"COLUMN":                           true,
		"COLUMN_NAME":                      false,
		"COLUMNS":                          false,
		"COMMAND_FUNCTION":                 false,
		"COMMAND_FUNCTION_CODE":            false,
		"COMMIT":                           true,
		"COMMITTED":                        false,
		"CONDITION":                        true,
		"CONDITION_NUMBER":                 false,
		"CONDITIONAL":                      false,
		"CONNECT":                          true,
		"CONNECTION":                       true,
		"CONNECTION_NAME":                  false,
		"CONSTRAINT":                       true,
		"CONSTRAINT_CATALOG":               false,
		"CONSTRAINT_NAME":                  false,
		"CONSTRAINT_SCHEMA":                false,
		"CONSTRAINTS":                      true,
		"CONSTRUCTOR":                      false,
		"CONTAINS":                         true,
		"CONTENT":                          false,
		"CONTINUE":                         true,
		"CONTROL":                          false,
		"CONVERT":                          true,
		"COPY":                             true,
		"CORR":                             true,
		"CORRESPONDING":                    true,
		"COS":                              true,
		"COSH":                             true,
		"COUNT":                            true,
		"COVAR_POP":                        true,
		"COVAR_SAMP":                       true,
		"CREATE":                           true,
		"CROSS":                            true,
		"CUBE":                             true,
		"CUME_DIST":                        true,
		"CURRENT":                          true,
		"CURRENT_CATALOG":                  true,
		"CURRENT_DATE":                     true,
		"CURRENT_DEFAULT_TRANSFORM_GROUP":  true,
		"CURRENT_PATH":                     true,
		"CURRENT_ROLE":                     true,
		"CURRENT_ROW":                      true,
		"CURRENT_SCHEMA":                   true,
		"CURRENT_TIME":                     true,
		"CURRENT_TIMESTAMP":                true,
		"CURRENT_TRANSFORM_GROUP_FOR_TYPE": true,
		"CURRENT_USER":                     true,
		"CURSOR":                           true,
		"CURSOR_NAME":                      false,
		"CYCLE":                            true,
		"DATA":                             false,
		"DATALINK":                         true,
		"DATE":                             true,
		"DATETIME_INTERVAL_CODE":           false,
		"DATETIME_INTERVAL_PRECISION":      false,
		"DAY":                              true,
		"DB":                               false,
		"DEALLOCATE":                       true,
		"DEC":                              true,
		"DECFLOAT":                         true,
		"DECIMAL":                          true,
		"DECLARE":                          true,
		"DEFAULT":                          true,
		"DEFAULTS":                         false,
		"DEFERRABLE":                       true,
		"DEFERRED":                         true,
		"DEFINE":                           true,
		"DEFINED":                          false,
		"DEFINER":                          false,
		"DEGREE":                           false,
		"DELETE":                           true,
		"DENSE_RANK":                       true,
		"DEPTH":                            false,
		"DEREF":                            true,
		"DERIVED":                          false,
		"DESC":                             true,
		"DESCRIBE":                         true,
		"DESCRIPTOR":                       true,
		"DETERMINISTIC":                    true,
		"DIAGNOSTICS":                      true,
		"DISCONNECT":                       true,
		"DISPATCH":                         false,
		"DISTINCT":                         true,
		"DLNEWCOPY":                        true,
		"DLPREVIOUSCOPY":                   true,
		"DLURLCOMPLETE":                    true,
		"DLURLCOMPLETEONLY":                true,
		"DLURLCOMPLETEWRITE":               true,
		"DLURLPATH":                        true,
		"DLURLPATHONLY":                    true,
		"DLURLPATHWRITE":                   true,
		"DLURLSCHEME":                      true,
		"DLURLSERVER":                      true,
		"DLVALUE":                          true,
		"DOCUMENT":                         false,
		"DOMAIN":                           true,
		"DOUBLE":                           true,
		"DROP":                             true,
		"DYNAMIC":                          true,
		"DYNAMIC_FUNCTION":                 false,
		"DYNAMIC_FUNCTION_CODE":            false,
		"EACH":                             true,
		"ELEMENT":                          true,
		"ELSE":                             true,
		"EMPTY":                            true,
		"ENCODING":                         false,
		"END":                              true,
		"END_FRAME":                        true,
		"END_PARTITION":                    true,
		"END-EXEC":                         true,
		"ENFORCED":                         false,
		"EQUALS":                           true,
		"ERROR":                            false,
		"ESCAPE":                           true,
		"EVERY":                            true,
		"EXCEPT":                           true,
		"EXCEPTION":                        true,
		"EXCLUDE":                          false,
		"EXCLUDING":                        false,
		"EXEC":                             true,
		"EXECUTE":                          true,
		"EXISTS":                           true,
		"EXP":                              true,
		"EXPRESSION":                       false,
		"EXTERNAL":                         true,
		"EXTRACT":                          true,
		"FALSE":                            true,
		"FETCH":                            true,
		"FILE":                             false,
		"FILTER":                           true,
		"FINAL":                            false,
		"FINISH":                           false,
		"FIRST":                            true,
		"FIRST_VALUE":                      true,
		"FLAG":                             false,
		"FLOAT":                            true,
		"FLOOR":                            true,
		"FOLLOWING":                        false,
		"FOR":                              true,
		"FOREIGN":                          true,
		"FORMAT":                           false,
		"FORTRAN":                          false,
		"FOUND":                            true,
		"FRAME_ROW":                        true,
		"FREE":                             true,
		"FROM":                             true,
		"FS":                               false,
		"FULFILL":                          false,
		"FULL":                             true,
		"FUNCTION":                         true,
		"FUSION":                           true,
		"GENERAL":                          false,
		"GENERATED":                        false,
		"GET":                              true,
		"GLOBAL":                           true,
		"GO":                               true,
		"GOTO":                             true,
		"GRANT":                            true,
		"GRANTED":                          false,
		"GROUP":                            true,
		"GROUPING":                         true,
		"GROUPS":                           true,
		"HAVING":                           true,
		"HEX":                              false,
		"HIERARCHY":                        false,
		"HOLD":                             true,
		"HOUR":                             true,
		"ID":                               false,
		"IDENTITY":                         true,
		"IGNORE":                           false,
		"IMMEDIATE":                        true,
		"IMMEDIATELY":                      false,
		"IMPLEMENTATION":                   false,
		"IMPORT":                           true,
		"IN":                               true,
		"INCLUDING":                        false,
		"INCREMENT":                        false,
		"INDENT":                           false,
		"INDICATOR":                        true,
		"INITIAL":                          true,
		"INITIALLY":                        true,
		"INNER":                            true,
		"INOUT":                            true,
		"INPUT":                            true,
		"INSENSITIVE":                      true,
		"INSERT":                           true,
		"INSTANCE":                         false,
		"INSTANTIABLE":                     false,
		"INSTEAD":                          false,
		"INT":                              true,
		"INTEGER":                          true,
		"INTEGRITY":                        false,
		"INTERSECT":                        true,
		"INTERSECTION":                     true,
		"INTERVAL":                         true,
		"INTO":                             true,
		"INVOKER":                          false,
		"IS":                               true,
		"ISOLATION":                        true,
		"JOIN":                             true,
		"JSON":                             false,
		"JSON_ARRAY":                       true,
		"JSON_ARRAYAGG":                    true,
		"JSON_EXISTS":                      true,
		"JSON_OBJECT":                      true,
		"JSON_OBJECTAGG":                   true,
		"JSON_QUERY":                       true,
		"JSON_TABLE":                       true,
		"JSON_TABLE_PRIMITIVE":             true,
		"JSON_VALUE":                       true,
		"KEEP":                             false,
		"KEY":                              true,
		"KEY_MEMBER":                       false,
		"KEY_TYPE":                         false,
		"KEYS":                             false,
		"LAG":                              true,
		"LANGUAGE":                         true,
		"LARGE":                            true,
		"LAST":                             true,
		"LAST_VALUE":                       true,
		"LATERAL":                          true,
		"LEAD":                             true,
		"LEADING":                          true,
		"LEFT":                             true,
		"LENGTH":                           false,
		"LEVEL":                            true,
		"LIBRARY":                          false,
		"LIKE":                             true,
		"LIKE_REGEX":                       true,
		"LIMIT":                            false,
		"LINK":                             false,
		"LISTAGG":                          true,
		"LN":                               true,
		"LOCAL":                            true,
		"LOCALTIME":                        true,
		"LOCALTIMESTAMP":                   true,
		"LOCATION":                         false,
		"LOCATOR":                          false,
		"LOG":                              true,
		"LOG10":                            true,
		"LOWER":                            true,
		"MAP":                              false,
		"MAPPING":                          false,
		"MATCH":                            true,
		"MATCH_NUMBER":                     true,
		"MATCH_RECOGNIZE":                  true,
		"MATCHED":                          false,
		"MATCHES":                          true,
		"MAX":                              true,
		"MAXVALUE":                         false,
		"MEASURES":                         true,
		"MEMBER":                           true,
		"MERGE":                            true,
		"MESSAGE_LENGTH":                   false,
		"MESSAGE_OCTET_LENGTH":             false,
		"MESSAGE_TEXT":                     false,
		"METHOD":                           true,
		"MIN":                              true,
		"MINUTE":                           true,
		"MINVALUE":                         false,
		"MOD":                              true,
		"MODIFIES":                         true,
		"MODULE":                           true,
		"MONTH":                            true,
		"MORE":                             false,
		"MULTISET":                         true,
		"MUMPS":                            false,
		"NAME":                             false,
		"NAMES":                            true,
		"NAMESPACE":                        false,
		"NATIONAL":                         true,
		"NATURAL":                          true,
		"NCHAR":                            true,
		"NCLOB":                            true,
		"NESTED":                           false,
		"NESTING":                          false,
		"NEW":                              true,
		"NEXT":                             true,
		"NFC":                              false,
		"NFD":                              false,
		"NFKC":                             false,
		"NFKD":                             false,
		"NIL":                              false,
		"NO":                               true,
		"NONE":                             true,
		"NORMALIZE":                        true,
		"NORMALIZED":                       false,
		"NOT":                              true,
		"NTH_VALUE":                        true,
		"NTILE":                            true,
		"NULL":                             true,
		"NULLABLE":                         false,
		"NULLIF":                           true,
		"NULLS":                            false,
		"NUMBER":                           false,
		"NUMERIC":                          true,
		"OBJECT":                           false,
		"OCCURRENCES_REGEX":                true,
		"OCTET_LENGTH":                     true,
		"OCTETS":                           false,
		"OF":                               true,
		"OFF":                              false,
		"OFFSET":                           true,
		"OLD":                              true,
		"OMIT":                             true,
		"ON":                               true,
		"ONE":                              true,
		"ONLY":                             true,
		"OPEN":                             true,
		"OPTION":                           true,
		"OPTIONS":                          false,
		"OR":                               true,
		"ORDER":                            true,
		"ORDERING":                         false,
		"ORDINALITY":                       false,
		"OTHERS":                           false,
		"OUT":                              true,
		"OUTER":                            true,
		"OUTPUT":                           true,
		"OVER":                             true,
		"OVERFLOW":                         false,
		"OVERLAPS":                         true,
		"OVERLAY":                          true,
		"OVERRIDING":                       false,
		"PAD":                              true,
		"PARAMETER":                        true,
		"PARAMETER_MODE":                   false,
		"PARAMETER_NAME":                   false,
		"PARAMETER_ORDINAL_POSITION":       false,
		"PARAMETER_SPECIFIC_CATALOG":       false,
		"PARAMETER_SPECIFIC_NAME":          false,
		"PARAMETER_SPECIFIC_SCHEMA":        false,
		"PARTIAL":                          true,
		"PARTITION":                        true,
		"PASCAL":                           false,
		"PASS":                             false,
		"PASSING":                          false,
		"PASSTHROUGH":                      false,
		"PAST":                             false,
		"PATH":                             false,
		"PATTERN":                          true,
		"PER":                              true,
		"PERCENT":                          true,
		"PERCENT_RANK":                     true,
		"PERCENTILE_CONT":                  true,
		"PERCENTILE_DISC":                  true,
		"PERIOD":                           true,
		"PERMISSION":                       false,
		"PERMUTE":                          true,
		"PLACING":                          false,
		"PLAN":                             false,
		"PLI":                              false,
		"PORTION":                          true,
		"POSITION":                         true,
		"POSITION_REGEX":                   true,
		"POWER":                            true,
		"PRECEDES":                         true,
		"PRECEDING":                        false,
		"PRECISION":                        true,
		"PREPARE":                          true,
		"PRESERVE":                         true,
		"PRIMARY":                          true,
		"PRIOR":                            true,
		"PRIVATE":                          false,
		"PRIVILEGES":                       true,
		"PROCEDURE":                        true,
		"PRUNE":                            false,
		"PTF":                              true,
		"PUBLIC":                           true,
		"QUOTES":                           false,
		"RANGE":                            true,
		"RANK":                             true,
		"READ":                             true,
		"READS":                            true,
		"REAL":                             true,
		"RECOVERY":                         false,
		"RECURSIVE":                        true,
		"REF":                              true,
		"REFERENCES":                       true,
		"REFERENCING":                      true,
		"REGR_AVGX":                        true,
		"REGR_AVGY":                        true,
		"REGR_COUNT":                       true,
		"REGR_INTERCEPT":                   true,
		"REGR_R2":                          true,
		"REGR_SLOPE":                       true,
		"REGR_SXX":                         true,
		"REGR_SXY":                         true,
		"REGR_SYY":                         true,
		"RELATIVE":                         true,
		"RELEASE":                          true,
		"REPEATABLE":                       false,
		"REQUIRING":                        false,
		"RESPECT":                          false,
		"RESTART":                          false,
		"RESTORE":                          false,
		"RESTRICT":                         true,
		"RESULT":                           true,
		"RETURN":                           true,
		"RETURNED_CARDINALITY":             false,
		"RETURNED_LENGTH":                  false,
		"RETURNED_OCTET_LENGTH":            false,
		"RETURNED_SQLSTATE":                false,
		"RETURNING":                        false,
		"RETURNS":                          true,
		"REVOKE":                           true,
		"RIGHT":                            true,
		"ROLE":                             false,
		"ROLLBACK":                         true,
		"ROLLUP":                           true,
		"ROUTINE":                          false,
		"ROUTINE_CATALOG":                  false,
		"ROUTINE_NAME":                     false,
		"ROUTINE_SCHEMA":                   false,
		"ROW":                              true,
		"ROW_COUNT":                        false,
		"ROW_NUMBER":                       true,
		"ROWS":                             true,
		"RUNNING":                          true,
		"SAVEPOINT":                        true,
		"SCALAR":                           false,
		"SCALE":                            false,
		"SCHEMA":                           true,
		"SCHEMA_NAME":                      false,
		"SCOPE":                            true,
		"SCOPE_CATALOG":                    false,
		"SCOPE_NAME":                       false,
		"SCOPE_SCHEMA":                     false,
		"SCROLL":                           true,
		"SEARCH":                           true,
		"SECOND":                           true,
		"SECTION":                          true,
		"SECURITY":                         false,
		"SEEK":                             true,
		"SELECT":                           true,
		"SELECTIVE":                        false,
		"SELF":                             false,
		"SENSITIVE":                        true,
		"SEQUENCE":                         false,
		"SERIALIZABLE":                     false,
		"SERVER":                           false,
		"SERVER_NAME":                      false,
		"SESSION":                          true,
		"SESSION_USER":                     true,
		"SET":                              true,
		"SETS":                             false,
		"SHOW":                             true,
		"SIMILAR":                          true,
		"SIMPLE":                           false,
		"SIN":                              true,
		"SINH":                             true,
		"SIZE":                             true,
		"SKIP":                             true,
		"SMALLINT":                         true,
		"SOME":                             true,
		"SOURCE":                           false,
		"SPACE":                            true,
		"SPECIFIC":                         true,
		"SPECIFIC_NAME":                    false,
		"SPECIFICTYPE":                     true,
		"SQL":                              true,
		"SQLCODE":                          true,
		"SQLERROR":                         true,
		"SQLEXCEPTION":                     true,
		"SQLSTATE":                         true,
		"SQLWARNING":                       true,
		"SQRT":                             true,
		"STANDALONE":                       false,
		"START":                            true,
		"STATE":                            false,
		"STATEMENT":                        false,
		"STATIC":                           true,
		"STDDEV_POP":                       true,
		"STDDEV_SAMP":                      true,
		"STRING":                           false,
		"STRIP":                            false,
		"STRUCTURE":                        false,
		"STYLE":                            false,
		"SUBCLASS_ORIGIN":                  false,
		"SUBMULTISET":                      true,
		"SUBSET":                           true,
		"SUBSTRING":                        true,
		"SUBSTRING_REGEX":                  true,
		"SUCCEEDS":                         true,
		"SUM":                              true,
		"SYMMETRIC":                        true,
		"SYSTEM":                           true,
		"SYSTEM_TIME":                      true,
		"SYSTEM_USER":                      true,
		"TABLE":                            true,
		"TABLE_NAME":                       false,
		"TABLESAMPLE":                      true,
		"TAN":                              true,
		"TANH":                             true,
		"TEMPORARY":                        true,
		"THEN":                             true,
		"THROUGH":                          false,
		"TIES":                             false,
		"TIME":                             true,
		"TIMESTAMP":                        true,
		"TIMEZONE_HOUR":                    true,
		"TIMEZONE_MINUTE":                  true,
		"TO":                               true,
		"TOKEN":                            false,
		"TOP_LEVEL_COUNT":                  false,
		"TRAILING":                         true,
		"TRANSACTION":                      true,
		"TRANSACTION_ACTIVE":               false,
		"TRANSACTIONS_COMMITTED":           false,
		"TRANSACTIONS_ROLLED_BACK":         false,
		"TRANSFORM":                        false,
		"TRANSFORMS":                       false,
		"TRANSLATE":                        true,
		"TRANSLATE_REGEX":                  true,
		"TRANSLATION":                      true,
		"TREAT":                            true,
		"TRIGGER":                          true,
		"TRIGGER_CATALOG":                  false,
		"TRIGGER_NAME":                     false,
		"TRIGGER_SCHEMA":                   false,
		"TRIM":                             true,
		"TRIM_ARRAY":                       true,
		"TRUE":                             true,
		"TRUNCATE":                         true,
		"TYPE":                             false,
		"UESCAPE":                          true,
		"UNBOUNDED":                        false,
		"UNCOMMITTED":                      false,
		"UNCONDITIONAL":                    false,
		"UNDER":                            false,
		"UNION":                            true,
		"UNIQUE":                           true,
		"UNKNOWN":                          true,
		"UNLINK":                           false,
		"UNMATCHED":                        true,
		"UNNAMED":                          false,
		"UNNEST":                           true,
		"UNTYPED":                          false,
		"UPDATE":                           true,
		"UPPER":                            true,
		"URI":                              false,
		"USAGE":                            true,
		"USER":                             true,
		"USER_DEFINED_TYPE_CATALOG":        false,
		"USER_DEFINED_TYPE_CODE":           false,
		"USER_DEFINED_TYPE_NAME":           false,
		"USER_DEFINED_TYPE_SCHEMA":         false,
		"USING":                            true,
		"UTF16":                            false,
		"UTF32":                            false,
		"UTF8":                             false,
		"VALID":                            false,
		"VALUE":                            true,
		"VALUE_OF":                         true,
		"VALUES":                           true,
		"VAR_POP":                          true,
		"VAR_SAMP":                         true,
		"VARBINARY":                        true,
		"VARCHAR":                          true,
		"VARYING":                          true,
		"VERSION":                          false,
		"VERSIONING":                       true,
		"VIEW":                             true,
		"WHEN":                             true,
		"WHENEVER":                         true,
		"WHERE":                            true,
		"WHITESPACE":                       false,
		"WIDTH_BUCKET":                     true,
		"WINDOW":                           true,
		"WITH":                             true,
		"WITHIN":                           true,
		"WITHOUT":                          true,
		"WORK":                             true,
		"WRAPPER":                          false,
		"WRITE":                            true,
		"XML":                              true,
		"XMLAGG":                           true,
		"XMLATTRIBUTES":                    true,
		"XMLBINARY":                        true,
		"XMLCAST":                          true,
		"XMLCOMMENT":                       true,
		"XMLCONCAT":                        true,
		"XMLDECLARATION":                   false,
		"XMLDOCUMENT":                      true,
		"XMLELEMENT":                       true,
		"XMLEXISTS":                        true,
		"XMLFOREST":                        true,
		"XMLITERATE":                       true,
		"XMLNAMESPACES":                    true,
		"XMLPARSE":                         true,
		"XMLPI":                            true,
		"XMLQUERY":                         true,
		"XMLSCHEMA":                        false,
		"XMLSERIALIZE":                     true,
		"XMLTABLE":                         true,
		"XMLTEXT":                          true,
		"XMLVALIDATE":                      true,
		"YEAR":                             true,
		"YES":                              false,
		"ZONE":                             true,
	}

	v, ok := sqlStandardKeywords[strings.ToUpper(s)]

	return ok, v
}

// IsKeyword returns a boolean indicating if the supplied string
// is considered to be a keyword in ISO standared SQL
func (d StandardSQLDialect) IsKeyword(s string) bool {
	isKey, _ := d.keyword(s)
	return isKey
}

// IsReservedKeyword returns a boolean indicating if the supplied
// string is considered to be a reserved keyword in ISO standard SQL
func (d StandardSQLDialect) IsReservedKeyword(s string) bool {

	isKey, isReserved := d.keyword(s)

	if isKey {
		return isReserved
	}
	return false
}

// IsOperator returns a boolean indicating if the supplied string
// is considered to be an operator in ISO standard SQL
func (d StandardSQLDialect) IsOperator(s string) bool {

	sqlStandardOperators := map[string]bool{
		"+":  true,
		"-":  true,
		"*":  true,
		"/":  true,
		"%":  true,
		"=":  true,
		"!=": true,
		"<>": true,
		">":  true,
		"<":  true,
		">=": true,
		"<=": true,
		"!<": true,
		"!>": true,
	}
	_, ok := sqlStandardOperators[s]
	return ok
}

// IsLabel returns a boolean indicating if the supplied string
// is considered to be a label in ISO standard SQL
func (d StandardSQLDialect) IsLabel(s string) bool {
	return false
}

// IsIdentifier returns a boolean indicating if the supplied
// string is considered to be a non-quoted Standard SQL identifier.
func (d StandardSQLDialect) IsIdentifier(s string) bool {

	// not certain... but considering the PostgreSQL doumentation:
	//
	// "SQL identifiers and key words must begin with a letter (a-z, but
	// also letters with diacritical marks and non-Latin letters) or an
	// underscore (_). Subsequent characters in an identifier or key word
	// can be letters, underscores, digits (0-9), or dollar signs ($).
	// Note that dollar signs are not allowed in identifiers according to
	// the letter of the SQL standard, so their use might render
	// applications less portable. The SQL standard will not define a key
	// word that contains digits or starts or ends with an underscore, so
	// identifiers of this form are safe against possible conflict with
	// future extensions of the standard."

	const firstIdentChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const identChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

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
