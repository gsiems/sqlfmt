package dialect

import (
	"regexp"
	"strings"
)

type PostgreSQLDialect struct {
	dialect int
	name    string
}

func NewPostgreSQLDialect() PostgreSQLDialect {
	var d PostgreSQLDialect

	d.dialect = PostgreSQL
	d.name = "PostgreSQL"

	return d
}

func (d PostgreSQLDialect) Dialect() int {
	return d.dialect
}
func (d PostgreSQLDialect) DialectName() string {
	return d.name
}
func (d PostgreSQLDialect) CaseFolding() int {
	return FoldLower
}
func (d PostgreSQLDialect) IdentQuoteChar() string {
	return "\""
}
func (d PostgreSQLDialect) StringQuoteChar() string {
	return "'"
}

// MaxOperatorLength returns the length of the longest operator
// supported by the dialect
func (d PostgreSQLDialect) MaxOperatorLength() int {

	// Per https://www.postgresql.org/docs/current/sql-createoperator.html
	// NAMEDATALEN-1 (63 by default)
	return 63
}

// IsDatatype returns a boolean indicating if the supplied string
// (or string slice) is considered to be a datatype in PostgreSQL
func (d PostgreSQLDialect) IsDatatype(s ...string) bool {

	var pgDatatypes = map[string]bool{
		"bigint":                           true,
		"bigserial":                        true,
		"bit":                              true, // [(n)]
		"bit (n)":                          true, // [(n)]
		"bit varying":                      true, // [(n)]
		"bit varying (n)":                  true, // [(n)]
		"boolean":                          true,
		"bool":                             true, // alternate/abbreviated form
		"box":                              true,
		"bytea":                            true,
		"\"char\"":                         true, // datatype seen in pg_catalog
		"\"char\" (n)":                     true, // datatype seen in pg_catalog
		"char":                             true, // [(n)] alternate/abbreviated form
		"char (n)":                         true, // [(n)] alternate/abbreviated form
		"character":                        true, // [(n)]
		"character (n)":                    true, // [(n)]
		"character varying":                true, // [(n)]
		"character varying (n)":            true, // [(n)]
		"cidr":                             true,
		"circle":                           true,
		"datemultirange":                   true,
		"daterange":                        true,
		"date":                             true,
		"decimal":                          true, // [p[,s])
		"decimal (n)":                      true, // [p[,s])
		"decimal (n,n)":                    true, // [p[,s])
		"double precision":                 true,
		"float":                            true, // [(p)] alternate/abbreviated form
		"float4":                           true, // alternate/abbreviated form
		"float8":                           true, // alternate/abbreviated form
		"inet":                             true,
		"integer":                          true,
		"int":                              true, // alternate/abbreviated form
		"int2":                             true, // alternate/abbreviated form
		"int4multirange":                   true,
		"int4range":                        true,
		"int4":                             true, // alternate/abbreviated form
		"int8multirange":                   true,
		"int8range":                        true,
		"int8":                             true, // alternate/abbreviated form
		"interval":                         true, // [(p)]
		"interval (n)":                     true, // [(p)]
		"interval day to hour":             true, // expanded fields
		"interval day to minute":           true, // expanded fields
		"interval day to second":           true, // expanded fields
		"interval day":                     true, // expanded fields
		"interval hour to minute":          true, // expanded fields
		"interval hour to second":          true, // expanded fields
		"interval hour":                    true, // expanded fields
		"interval minute to second":        true, // expanded fields
		"interval minute":                  true, // expanded fields
		"interval month":                   true, // expanded fields
		"interval second":                  true, // expanded fields
		"interval year to month":           true, // expanded fields
		"interval year":                    true, // expanded fields
		"jsonb":                            true,
		"json":                             true,
		"line":                             true,
		"lseg":                             true,
		"macaddr8":                         true,
		"macaddr":                          true,
		"money":                            true,
		"name":                             true, // datatype seen in pg_catalog
		"numeric":                          true, // [p[,s])
		"numeric (n)":                      true, // [p[,s])
		"numeric (n,n)":                    true, // [p[,s])
		"nummultirange":                    true,
		"numrange":                         true,
		"path":                             true,
		"pg_lsn":                           true,
		"pg_snapshot":                      true,
		"point":                            true,
		"polygon":                          true,
		"real":                             true,
		"serial":                           true,
		"smallint":                         true,
		"smallserial":                      true,
		"text":                             true,
		"timestamp":                        true,
		"timestamp (n)":                    true,
		"timestamp without time zone":      true,
		"timestamp (n) without time zone":  true,
		"timestamp with time zone":         true,
		"timestamp (n) with time zone":     true,
		"timestamptz":                      true, // alternate/abbreviated form
		"timetz":                           true, // alternate/abbreviated form
		"time":                             true,
		"time (n)":                         true,
		"time without time zone":           true, // [(p)]
		"time (n) without time zone":       true, // [(p)]
		"time with time zone":              true, // [(p)]
		"time (n) with time zone":          true, // [(p)]
		"tsmultirange":                     true,
		"tsquery":                          true,
		"tsrange":                          true,
		"tstzmultirange":                   true,
		"tstzrange":                        true,
		"tsvector":                         true,
		"txid_snapshot":                    true,
		"uuid":                             true,
		"varchar":                          true, // [(n)]
		"varchar (n)":                      true, // [(n)]
		"xml":                              true,
		"oid":                              true, // object identifier types
		"regclass":                         true, // object identifier types
		"regcollation":                     true, // object identifier types
		"regconfig":                        true, // object identifier types
		"regdictionary":                    true, // object identifier types
		"regnamespace":                     true, // object identifier types
		"regoper":                          true, // object identifier types
		"regoperator":                      true, // object identifier types
		"regproc":                          true, // object identifier types
		"regprocedure":                     true, // object identifier types
		"regrole":                          true, // object identifier types
		"regtype":                          true, // object identifier types
		"box2d":                            true, // PostGIS extension
		"box3d":                            true, // PostGIS extension
		"geography":                        true, // PostGIS extension
		"geography (geometrycollection,n)": true, // PostGIS extension
		"geography (geometrycollection)":   true, // PostGIS extension
		"geography (linestring,n)":         true, // PostGIS extension
		"geography (linestring)":           true, // PostGIS extension
		"geography (multilinestring,n)":    true, // PostGIS extension
		"geography (multilinestring)":      true, // PostGIS extension
		"geography (multipoint,n)":         true, // PostGIS extension
		"geography (multipoint)":           true, // PostGIS extension
		"geography (multipolygon,n)":       true, // PostGIS extension
		"geography (multipolygon)":         true, // PostGIS extension
		"geography (point,n)":              true, // PostGIS extension
		"geography (point)":                true, // PostGIS extension
		"geography (polygon,n)":            true, // PostGIS extension
		"geography (polygon)":              true, // PostGIS extension
		"geometry_dump":                    true, // PostGIS extension
		"geometry":                         true, // PostGIS extension
		"geometry (geometrycollection,n)":  true, // PostGIS extension
		"geometry (geometrycollection)":    true, // PostGIS extension
		"geometry (linestring,n)":          true, // PostGIS extension
		"geometry (linestring)":            true, // PostGIS extension
		"geometry (multilinestring,n)":     true, // PostGIS extension
		"geometry (multilinestring)":       true, // PostGIS extension
		"geometry (multipoint,n)":          true, // PostGIS extension
		"geometry (multipoint)":            true, // PostGIS extension
		"geometry (multipolygon,n)":        true, // PostGIS extension
		"geometry (multipolygon)":          true, // PostGIS extension
		"geometry (point,n)":               true, // PostGIS extension
		"geometry (point)":                 true, // PostGIS extension
		"geometry (polygon,n)":             true, // PostGIS extension
		"geometry (polygon)":               true, // PostGIS extension
	}

	var z []string
	rn := regexp.MustCompile(`^[0-9]+$`)
	pv := ""

	for i, v := range s {
		switch v {
		case "(":
			z = append(z, " "+v)
		case ")", ",", "[", "]":
			z = append(z, v)
		default:
			switch {
			case rn.MatchString(v):
				z = append(z, "n")
			case i == 0:
				z = append(z, v)
			case pv == "(":
				z = append(z, v)
			default:
				z = append(z, " "+v)
			}
		}
		pv = v
	}

	k := strings.ToLower(strings.Join(z, ""))
	if _, ok := pgDatatypes[k]; ok {
		return true
	}

	// Check for an array of the datatype
	k = strings.TrimRight(k, "[]")
	if _, ok := pgDatatypes[k]; ok {
		return true
	}

	return false
}

func (d PostgreSQLDialect) keyword(s string) (bool, bool) {

	/*
	   PostgreSQL keywords

	   https://www.postgresql.org/docs/current/sql-keywords-appendix.html

	*/

	// map[keyword]isReserved
	var pgKeywords = map[string]bool{
		"ABORT":                         false,
		"ACCESS":                        false,
		"AGGREGATE":                     false,
		"ALSO":                          false,
		"ANALYSE":                       true,
		"ANALYZE":                       true,
		"ATTACH":                        false,
		"BACKWARD":                      false,
		"BIT":                           false,
		"BIT_LENGTH":                    false,
		"CACHE":                         false,
		"CHECKPOINT":                    false,
		"CLASS":                         false,
		"CLUSTER":                       false,
		"COMMENT":                       false,
		"COMMENTS":                      false,
		"COMPRESSION":                   false,
		"CONCURRENTLY":                  true,
		"CONFIGURATION":                 false,
		"CONFLICT":                      false,
		"CONVERSION":                    false,
		"COST":                          false,
		"CSV":                           false,
		"DATABASE":                      false,
		"DELIMITER":                     false,
		"DELIMITERS":                    false,
		"DEPENDS":                       false,
		"DETACH":                        false,
		"DICTIONARY":                    false,
		"DISABLE":                       false,
		"DISCARD":                       false,
		"DO":                            true,
		"ENABLE":                        false,
		"ENCRYPTED":                     false,
		"ENUM":                          false,
		"EVENT":                         false,
		"EXCEPTION":                     false,
		"EXCLUSIVE":                     false,
		"EXPLAIN":                       false,
		"EXTENSION":                     false,
		"FAMILY":                        false,
		"FINALIZE":                      false,
		"FORCE":                         false,
		"FORWARD":                       false,
		"FREEZE":                        true,
		"FUNCTIONS":                     false,
		"HANDLER":                       false,
		"HEADER":                        false,
		"IF":                            false,
		"ILIKE":                         true,
		"IMMUTABLE":                     false,
		"IMPLICIT":                      false,
		"INCLUDE":                       false,
		"INDEX":                         false,
		"INDEXES":                       false,
		"INHERIT":                       false,
		"INHERITS":                      false,
		"INLINE":                        false,
		"ISNULL":                        true,
		"LABEL":                         false,
		"LEAKPROOF":                     false,
		"LISTEN":                        false,
		"LOAD":                          false,
		"LOCK":                          false,
		"LOCKED":                        false,
		"LOGGED":                        false,
		"MATERIALIZED":                  false,
		"MODE":                          false,
		"MOVE":                          false,
		"NOTHING":                       false,
		"NOTIFY":                        false,
		"NOTNULL":                       true,
		"NOWAIT":                        false,
		"OIDS":                          false,
		"OPERATOR":                      false,
		"OWNED":                         false,
		"OWNER":                         false,
		"PARALLEL":                      false,
		"PARSER":                        false,
		"PASSWORD":                      false,
		"PLANS":                         false,
		"POLICY":                        false,
		"PREPARED":                      false,
		"PROCEDURAL":                    false,
		"PROCEDURES":                    false,
		"PROGRAM":                       false,
		"PUBLICATION":                   false,
		"QUOTE":                         false,
		"REASSIGN":                      false,
		"RECHECK":                       false,
		"REFRESH":                       false,
		"REINDEX":                       false,
		"RENAME":                        false,
		"REPLACE":                       false,
		"REPLICA":                       false,
		"RESET":                         false,
		"ROUTINES":                      false,
		"RULE":                          false,
		"SCHEMAS":                       false,
		"SEQUENCES":                     false,
		"SETOF":                         false,
		"SHARE":                         false,
		"SNAPSHOT":                      false,
		"SQLCODE":                       false,
		"SQLERROR":                      false,
		"STABLE":                        false,
		"STATISTICS":                    false,
		"STDIN":                         false,
		"STDOUT":                        false,
		"STORAGE":                       false,
		"STORED":                        false,
		"STRICT":                        false,
		"SUBSCRIPTION":                  false,
		"SUPPORT":                       false,
		"SYSID":                         false,
		"TABLES":                        false,
		"TABLESPACE":                    false,
		"TEMP":                          false,
		"TEMPLATE":                      false,
		"TEXT":                          false,
		"TRUSTED":                       false,
		"TYPES":                         false,
		"UNENCRYPTED":                   false,
		"UNLISTEN":                      false,
		"UNLOGGED":                      false,
		"UNTIL":                         false,
		"VACUUM":                        false,
		"VALIDATE":                      false,
		"VALIDATOR":                     false,
		"VARIADIC":                      true,
		"VERBOSE":                       true,
		"VIEWS":                         false,
		"VOLATILE":                      false,
		"XMLROOT":                       false,
		"A":                             false,
		"ABSOLUTE":                      false,
		"ACCORDING":                     false,
		"ACTION":                        false,
		"ADA":                           false,
		"ADD":                           false,
		"ADMIN":                         false,
		"AFTER":                         false,
		"ALWAYS":                        false,
		"ASC":                           true,
		"ASSERTION":                     false,
		"ASSIGNMENT":                    false,
		"ATTRIBUTE":                     false,
		"ATTRIBUTES":                    false,
		"BASE64":                        false,
		"BEFORE":                        false,
		"BERNOULLI":                     false,
		"BLOCKED":                       false,
		"BOM":                           false,
		"BREADTH":                       false,
		"C":                             false,
		"CASCADE":                       false,
		"CATALOG":                       false,
		"CATALOG_NAME":                  false,
		"CHAIN":                         false,
		"CHAINING":                      false,
		"CHARACTER_​SET_​CATALOG":       false,
		"CHARACTER_SET_NAME":            false,
		"CHARACTER_SET_SCHEMA":          false,
		"CHARACTERISTICS":               false,
		"CHARACTERS":                    false,
		"CLASS_ORIGIN":                  false,
		"COBOL":                         false,
		"COLLATION":                     true,
		"COLLATION_CATALOG":             false,
		"COLLATION_NAME":                false,
		"COLLATION_SCHEMA":              false,
		"COLUMN_NAME":                   false,
		"COLUMNS":                       false,
		"COMMAND_FUNCTION":              false,
		"COMMAND_​FUNCTION_​CODE":       false,
		"COMMITTED":                     false,
		"CONDITION_NUMBER":              false,
		"CONDITIONAL":                   false,
		"CONNECTION":                    false,
		"CONNECTION_NAME":               false,
		"CONSTRAINT_CATALOG":            false,
		"CONSTRAINT_NAME":               false,
		"CONSTRAINT_SCHEMA":             false,
		"CONSTRAINTS":                   false,
		"CONSTRUCTOR":                   false,
		"CONTENT":                       false,
		"CONTINUE":                      false,
		"CONTROL":                       false,
		"COPARTITION":                   false,
		"CURSOR_NAME":                   false,
		"DATA":                          false,
		"DATETIME_​INTERVAL_​CODE":      false,
		"DATETIME_​INTERVAL_​PRECISION": false,
		"DB":                            false,
		"DEFAULTS":                      false,
		"DEFERRABLE":                    true,
		"DEFERRED":                      false,
		"DEFINED":                       false,
		"DEFINER":                       false,
		"DEGREE":                        false,
		"DEPTH":                         false,
		"DERIVED":                       false,
		"DESC":                          true,
		"DESCRIPTOR":                    false,
		"DIAGNOSTICS":                   false,
		"DISPATCH":                      false,
		"DOCUMENT":                      false,
		"DOMAIN":                        false,
		"DYNAMIC_FUNCTION":              false,
		"DYNAMIC_​FUNCTION_​CODE":       false,
		"ENCODING":                      false,
		"ENFORCED":                      false,
		"ERROR":                         false,
		"EXCLUDE":                       false,
		"EXCLUDING":                     false,
		"EXPRESSION":                    false,
		"FILE":                          false,
		"FINAL":                         false,
		"FINISH":                        false,
		"FIRST":                         false,
		"FLAG":                          false,
		"FOLLOWING":                     false,
		"FORMAT":                        false,
		"FORTRAN":                       false,
		"FOUND":                         false,
		"FS":                            false,
		"FULFILL":                       false,
		"G":                             false,
		"GENERAL":                       false,
		"GENERATED":                     false,
		"GO":                            false,
		"GOTO":                          false,
		"GRANTED":                       false,
		"HEX":                           false,
		"HIERARCHY":                     false,
		"ID":                            false,
		"IGNORE":                        false,
		"IMMEDIATE":                     false,
		"IMMEDIATELY":                   false,
		"IMPLEMENTATION":                false,
		"INCLUDING":                     false,
		"INCREMENT":                     false,
		"INDENT":                        false,
		"INITIALLY":                     true,
		"INPUT":                         false,
		"INSTANCE":                      false,
		"INSTANTIABLE":                  false,
		"INSTEAD":                       false,
		"INTEGRITY":                     false,
		"INVOKER":                       false,
		"ISOLATION":                     false,
		"K":                             false,
		"KEEP":                          false,
		"KEY":                           false,
		"KEY_MEMBER":                    false,
		"KEY_TYPE":                      false,
		"KEYS":                          false,
		"LAST":                          false,
		"LENGTH":                        false,
		"LEVEL":                         false,
		"LIBRARY":                       false,
		"LIMIT":                         true,
		"LINK":                          false,
		"LOCATION":                      false,
		"LOCATOR":                       false,
		"M":                             false,
		"MAP":                           false,
		"MAPPING":                       false,
		"MATCHED":                       false,
		"MAXVALUE":                      false,
		"MEASURES":                      false,
		"MESSAGE_LENGTH":                false,
		"MESSAGE_OCTET_LENGTH":          false,
		"MESSAGE_TEXT":                  false,
		"MINVALUE":                      false,
		"MORE":                          false,
		"MUMPS":                         false,
		"NAME":                          false,
		"NAMES":                         false,
		"NAMESPACE":                     false,
		"NESTED":                        false,
		"NESTING":                       false,
		"NEXT":                          false,
		"NFC":                           false,
		"NFD":                           false,
		"NFKC":                          false,
		"NFKD":                          false,
		"NIL":                           false,
		"NORMALIZED":                    false,
		"NULL_ORDERING":                 false,
		"NULLABLE":                      false,
		"NULLS":                         false,
		"NUMBER":                        false,
		"OBJECT":                        false,
		"OCCURRENCE":                    false,
		"OCTETS":                        false,
		"OFF":                           false,
		"OPTION":                        false,
		"OPTIONS":                       false,
		"ORDERING":                      false,
		"ORDINALITY":                    false,
		"OTHERS":                        false,
		"OUTPUT":                        false,
		"OVERFLOW":                      false,
		"OVERRIDING":                    false,
		"P":                             false,
		"PAD":                           false,
		"PARAMETER_MODE":                false,
		"PARAMETER_NAME":                false,
		"PARAMETER_​ORDINAL_​POSITION":  false,
		"PARAMETER_​SPECIFIC_​CATALOG":  false,
		"PARAMETER_​SPECIFIC_​NAME":     false,
		"PARAMETER_​SPECIFIC_​SCHEMA":   false,
		"PARTIAL":                       false,
		"PASCAL":                        false,
		"PASS":                          false,
		"PASSING":                       false,
		"PASSTHROUGH":                   false,
		"PAST":                          false,
		"PATH":                          false,
		"PERMISSION":                    false,
		"PERMUTE":                       false,
		"PIPE":                          false,
		"PLACING":                       true,
		"PLAN":                          false,
		"PLI":                           false,
		"PRECEDING":                     false,
		"PRESERVE":                      false,
		"PREV":                          false,
		"PRIOR":                         false,
		"PRIVATE":                       false,
		"PRIVILEGES":                    false,
		"PRUNE":                         false,
		"PUBLIC":                        false,
		"QUOTES":                        false,
		"READ":                          false,
		"RECOVERY":                      false,
		"RELATIVE":                      false,
		"REPEATABLE":                    false,
		"REQUIRING":                     false,
		"RESPECT":                       false,
		"RESTART":                       false,
		"RESTORE":                       false,
		"RESTRICT":                      false,
		"RETURNED_CARDINALITY":          false,
		"RETURNED_LENGTH":               false,
		"RETURNED_​OCTET_​LENGTH":       false,
		"RETURNED_SQLSTATE":             false,
		"RETURNING":                     true,
		"ROLE":                          false,
		"ROUTINE":                       false,
		"ROUTINE_CATALOG":               false,
		"ROUTINE_NAME":                  false,
		"ROUTINE_SCHEMA":                false,
		"ROW_COUNT":                     false,
		"SCALAR":                        false,
		"SCALE":                         false,
		"SCHEMA":                        false,
		"SCHEMA_NAME":                   false,
		"SCOPE_CATALOG":                 false,
		"SCOPE_NAME":                    false,
		"SCOPE_SCHEMA":                  false,
		"SECTION":                       false,
		"SECURITY":                      false,
		"SELECTIVE":                     false,
		"SELF":                          false,
		"SEMANTICS":                     false,
		"SEQUENCE":                      false,
		"SERIALIZABLE":                  false,
		"SERVER":                        false,
		"SERVER_NAME":                   false,
		"SESSION":                       false,
		"SETS":                          false,
		"SIMPLE":                        false,
		"SIZE":                          false,
		"SORT_DIRECTION":                false,
		"SOURCE":                        false,
		"SPACE":                         false,
		"SPECIFIC_NAME":                 false,
		"STANDALONE":                    false,
		"STATE":                         false,
		"STATEMENT":                     false,
		"STRING":                        false,
		"STRIP":                         false,
		"STRUCTURE":                     false,
		"STYLE":                         false,
		"SUBCLASS_ORIGIN":               false,
		"T":                             false,
		"TABLE_NAME":                    false,
		"TEMPORARY":                     false,
		"THROUGH":                       false,
		"TIES":                          false,
		"TOKEN":                         false,
		"TOP_LEVEL_COUNT":               false,
		"TRANSACTION":                   false,
		"TRANSACTION_ACTIVE":            false,
		"TRANSACTIONS_​COMMITTED":       false,
		"TRANSACTIONS_​ROLLED_​BACK":    false,
		"TRANSFORM":                     false,
		"TRANSFORMS":                    false,
		"TRIGGER_CATALOG":               false,
		"TRIGGER_NAME":                  false,
		"TRIGGER_SCHEMA":                false,
		"TYPE":                          false,
		"UNBOUNDED":                     false,
		"UNCOMMITTED":                   false,
		"UNCONDITIONAL":                 false,
		"UNDER":                         false,
		"UNLINK":                        false,
		"UNMATCHED":                     false,
		"UNNAMED":                       false,
		"UNTYPED":                       false,
		"URI":                           false,
		"USAGE":                         false,
		"USER_​DEFINED_​TYPE_​CATALOG":  false,
		"USER_​DEFINED_​TYPE_​CODE":     false,
		"USER_​DEFINED_​TYPE_​NAME":     false,
		"USER_​DEFINED_​TYPE_​SCHEMA":   false,
		"UTF16":                         false,
		"UTF32":                         false,
		"UTF8":                          false,
		"VALID":                         false,
		"VERSION":                       false,
		"VIEW":                          false,
		"WHITESPACE":                    false,
		"WORK":                          false,
		"WRAPPER":                       false,
		"WRITE":                         false,
		"XMLDECLARATION":                false,
		"XMLSCHEMA":                     false,
		"YES":                           false,
		"ZONE":                          false,
		"FALSE":                         true,
		"TRUE":                          true,
		"ABS":                           true,
		"ABSENT":                        true,
		"ACOS":                          true,
		"ALL":                           true,
		"ALLOCATE":                      true,
		"ALTER":                         true,
		"AND":                           true,
		"ANY":                           true,
		"ANY_VALUE":                     true,
		"ARE":                           true,
		"ARRAY":                         true,
		"ARRAY_AGG":                     true,
		"ARRAY_​MAX_​CARDINALITY":       true,
		"AS":                            true,
		"ASENSITIVE":                    true,
		"ASIN":                          true,
		"ASYMMETRIC":                    true,
		"AT":                            true,
		"ATAN":                          true,
		"ATOMIC":                        true,
		"AUTHORIZATION":                 true,
		"AVG":                           true,
		"BEGIN":                         true,
		"BEGIN_FRAME":                   true,
		"BEGIN_PARTITION":               true,
		"BETWEEN":                       true,
		"BIGINT":                        true,
		"BINARY":                        true,
		"BLOB":                          true,
		"BOOLEAN":                       true,
		"BOTH":                          true,
		"BTRIM":                         true,
		"BY":                            true,
		"CALL":                          true,
		"CALLED":                        true,
		"CARDINALITY":                   true,
		"CASCADED":                      true,
		"CASE":                          true,
		"CAST":                          true,
		"CEIL":                          true,
		"CEILING":                       true,
		"CHAR":                          true,
		"CHAR_LENGTH":                   true,
		"CHARACTER":                     true,
		"CHARACTER_LENGTH":              true,
		"CHECK":                         true,
		"CLASSIFIER":                    true,
		"CLOB":                          true,
		"CLOSE":                         true,
		"COALESCE":                      true,
		"COLLATE":                       true,
		"COLLECT":                       true,
		"COLUMN":                        true,
		"COMMIT":                        true,
		"CONDITION":                     true,
		"CONNECT":                       true,
		"CONSTRAINT":                    true,
		"CONTAINS":                      true,
		"CONVERT":                       true,
		"COPY":                          true,
		"CORR":                          true,
		"CORRESPONDING":                 true,
		"COS":                           true,
		"COSH":                          true,
		"COUNT":                         true,
		"COVAR_POP":                     true,
		"COVAR_SAMP":                    true,
		"CREATE":                        true,
		"CROSS":                         true,
		"CUBE":                          true,
		"CUME_DIST":                     true,
		"CURRENT":                       true,
		"CURRENT_CATALOG":               true,
		"CURRENT_DATE":                  true,
		"CURRENT_​DEFAULT_​TRANSFORM_​GROUP": true,
		"CURRENT_PATH":      true,
		"CURRENT_ROLE":      true,
		"CURRENT_ROW":       true,
		"CURRENT_SCHEMA":    true,
		"CURRENT_TIME":      true,
		"CURRENT_TIMESTAMP": true,
		"CURRENT_​TRANSFORM_​GROUP_​FOR_​TYPE": true,
		"CURRENT_USER":         true,
		"CURSOR":               true,
		"CYCLE":                true,
		"DATALINK":             true,
		"DATE":                 true,
		"DAY":                  true,
		"DEALLOCATE":           true,
		"DEC":                  true,
		"DECFLOAT":             true,
		"DECIMAL":              true,
		"DECLARE":              true,
		"DEFAULT":              true,
		"DEFINE":               true,
		"DELETE":               true,
		"DENSE_RANK":           true,
		"DEREF":                true,
		"DESCRIBE":             true,
		"DETERMINISTIC":        true,
		"DISCONNECT":           true,
		"DISTINCT":             true,
		"DLNEWCOPY":            true,
		"DLPREVIOUSCOPY":       true,
		"DLURLCOMPLETE":        true,
		"DLURLCOMPLETEONLY":    true,
		"DLURLCOMPLETEWRITE":   true,
		"DLURLPATH":            true,
		"DLURLPATHONLY":        true,
		"DLURLPATHWRITE":       true,
		"DLURLSCHEME":          true,
		"DLURLSERVER":          true,
		"DLVALUE":              true,
		"DOUBLE":               true,
		"DROP":                 true,
		"DYNAMIC":              true,
		"EACH":                 true,
		"ELEMENT":              true,
		"ELSE":                 true,
		"EMPTY":                true,
		"END":                  true,
		"END_FRAME":            true,
		"END_PARTITION":        true,
		"END-EXEC":             true,
		"EQUALS":               true,
		"ESCAPE":               true,
		"EVERY":                true,
		"EXCEPT":               true,
		"EXEC":                 true,
		"EXECUTE":              true,
		"EXISTS":               true,
		"EXP":                  true,
		"EXTERNAL":             true,
		"EXTRACT":              true,
		"FETCH":                true,
		"FILTER":               true,
		"FIRST_VALUE":          true,
		"FLOAT":                true,
		"FLOOR":                true,
		"FOR":                  true,
		"FOREIGN":              true,
		"FRAME_ROW":            true,
		"FREE":                 true,
		"FROM":                 true,
		"FULL":                 true,
		"FUNCTION":             true,
		"FUSION":               true,
		"GET":                  true,
		"GLOBAL":               true,
		"GRANT":                true,
		"GREATEST":             true,
		"GROUP":                true,
		"GROUPING":             true,
		"GROUPS":               true,
		"HAVING":               true,
		"HOLD":                 true,
		"HOUR":                 true,
		"IDENTITY":             true,
		"IMPORT":               true,
		"IN":                   true,
		"INDICATOR":            true,
		"INITIAL":              true,
		"INNER":                true,
		"INOUT":                true,
		"INSENSITIVE":          true,
		"INSERT":               true,
		"INT":                  true,
		"INTEGER":              true,
		"INTERSECT":            true,
		"INTERSECTION":         true,
		"INTERVAL":             true,
		"INTO":                 true,
		"IS":                   true,
		"JOIN":                 true,
		"JSON":                 true,
		"JSON_ARRAY":           true,
		"JSON_ARRAYAGG":        true,
		"JSON_EXISTS":          true,
		"JSON_OBJECT":          true,
		"JSON_OBJECTAGG":       true,
		"JSON_QUERY":           true,
		"JSON_SCALAR":          true,
		"JSON_SERIALIZE":       true,
		"JSON_TABLE":           true,
		"JSON_TABLE_PRIMITIVE": true,
		"JSON_VALUE":           true,
		"LAG":                  true,
		"LANGUAGE":             true,
		"LARGE":                true,
		"LAST_VALUE":           true,
		"LATERAL":              true,
		"LEAD":                 true,
		"LEADING":              true,
		"LEAST":                true,
		"LEFT":                 true,
		"LIKE":                 true,
		"LIKE_REGEX":           true,
		"LISTAGG":              true,
		"LN":                   true,
		"LOCAL":                true,
		"LOCALTIME":            true,
		"LOCALTIMESTAMP":       true,
		"LOG":                  true,
		"LOG10":                true,
		"LOWER":                true,
		"LPAD":                 true,
		"LTRIM":                true,
		"MATCH":                true,
		"MATCH_NUMBER":         true,
		"MATCH_RECOGNIZE":      true,
		"MATCHES":              true,
		"MAX":                  true,
		"MEMBER":               true,
		"MERGE":                true,
		"METHOD":               true,
		"MIN":                  true,
		"MINUTE":               true,
		"MOD":                  true,
		"MODIFIES":             true,
		"MODULE":               true,
		"MONTH":                true,
		"MULTISET":             true,
		"NATIONAL":             true,
		"NATURAL":              true,
		"NCHAR":                true,
		"NCLOB":                true,
		"NEW":                  true,
		"NO":                   true,
		"NONE":                 true,
		"NORMALIZE":            true,
		"NOT":                  true,
		"NTH_VALUE":            true,
		"NTILE":                true,
		"NULL":                 true,
		"NULLIF":               true,
		"NUMERIC":              true,
		"OCCURRENCES_REGEX":    true,
		"OCTET_LENGTH":         true,
		"OF":                   true,
		"OFFSET":               true,
		"OLD":                  true,
		"OMIT":                 true,
		"ON":                   true,
		"ONE":                  true,
		"ONLY":                 true,
		"OPEN":                 true,
		"OR":                   true,
		"ORDER":                true,
		"OUT":                  true,
		"OUTER":                true,
		"OVER":                 true,
		"OVERLAPS":             true,
		"OVERLAY":              true,
		"PARAMETER":            true,
		"PARTITION":            true,
		"PATTERN":              true,
		"PER":                  true,
		"PERCENT":              true,
		"PERCENT_RANK":         true,
		"PERCENTILE_CONT":      true,
		"PERCENTILE_DISC":      true,
		"PERIOD":               true,
		"PORTION":              true,
		"POSITION":             true,
		"POSITION_REGEX":       true,
		"POWER":                true,
		"PRECEDES":             true,
		"PRECISION":            true,
		"PREPARE":              true,
		"PRIMARY":              true,
		"PROCEDURE":            true,
		"PTF":                  true,
		"RANGE":                true,
		"RANK":                 true,
		"READS":                true,
		"REAL":                 true,
		"RECURSIVE":            true,
		"REF":                  true,
		"REFERENCES":           true,
		"REFERENCING":          true,
		"REGR_AVGX":            true,
		"REGR_AVGY":            true,
		"REGR_COUNT":           true,
		"REGR_INTERCEPT":       true,
		"REGR_R2":              true,
		"REGR_SLOPE":           true,
		"REGR_SXX":             true,
		"REGR_SXY":             true,
		"REGR_SYY":             true,
		"RELEASE":              true,
		"RESULT":               true,
		"RETURN":               true,
		"RETURNS":              true,
		"REVOKE":               true,
		"RIGHT":                true,
		"ROLLBACK":             true,
		"ROLLUP":               true,
		"ROW":                  true,
		"ROW_NUMBER":           true,
		"ROWS":                 true,
		"RPAD":                 true,
		"RTRIM":                true,
		"RUNNING":              true,
		"SAVEPOINT":            true,
		"SCOPE":                true,
		"SCROLL":               true,
		"SEARCH":               true,
		"SECOND":               true,
		"SEEK":                 true,
		"SELECT":               true,
		"SENSITIVE":            true,
		"SESSION_USER":         true,
		"SET":                  true,
		"SHOW":                 true,
		"SIMILAR":              true,
		"SIN":                  true,
		"SINH":                 true,
		"SKIP":                 true,
		"SMALLINT":             true,
		"SOME":                 true,
		"SPECIFIC":             true,
		"SPECIFICTYPE":         true,
		"SQL":                  true,
		"SQLEXCEPTION":         true,
		"SQLSTATE":             true,
		"SQLWARNING":           true,
		"SQRT":                 true,
		"START":                true,
		"STATIC":               true,
		"STDDEV_POP":           true,
		"STDDEV_SAMP":          true,
		"SUBMULTISET":          true,
		"SUBSET":               true,
		"SUBSTRING":            true,
		"SUBSTRING_REGEX":      true,
		"SUCCEEDS":             true,
		"SUM":                  true,
		"SYMMETRIC":            true,
		"SYSTEM":               true,
		"SYSTEM_TIME":          true,
		"SYSTEM_USER":          true,
		"TABLE":                true,
		"TABLESAMPLE":          true,
		"TAN":                  true,
		"TANH":                 true,
		"THEN":                 true,
		"TIME":                 true,
		"TIMESTAMP":            true,
		"TIMEZONE_HOUR":        true,
		"TIMEZONE_MINUTE":      true,
		"TO":                   true,
		"TRAILING":             true,
		"TRANSLATE":            true,
		"TRANSLATE_REGEX":      true,
		"TRANSLATION":          true,
		"TREAT":                true,
		"TRIGGER":              true,
		"TRIM":                 true,
		"TRIM_ARRAY":           true,
		"TRUNCATE":             true,
		"UESCAPE":              true,
		"UNION":                true,
		"UNIQUE":               true,
		"UNKNOWN":              true,
		"UNNEST":               true,
		"UPDATE":               true,
		"UPPER":                true,
		"USER":                 true,
		"USING":                true,
		"VALUE":                true,
		"VALUE_OF":             true,
		"VALUES":               true,
		"VAR_POP":              true,
		"VAR_SAMP":             true,
		"VARBINARY":            true,
		"VARCHAR":              true,
		"VARYING":              true,
		"VERSIONING":           true,
		"WHEN":                 true,
		"WHENEVER":             true,
		"WHERE":                true,
		"WIDTH_BUCKET":         true,
		"WINDOW":               true,
		"WITH":                 true,
		"WITHIN":               true,
		"WITHOUT":              true,
		"XML":                  true,
		"XMLAGG":               true,
		"XMLATTRIBUTES":        true,
		"XMLBINARY":            true,
		"XMLCAST":              true,
		"XMLCOMMENT":           true,
		"XMLCONCAT":            true,
		"XMLDOCUMENT":          true,
		"XMLELEMENT":           true,
		"XMLEXISTS":            true,
		"XMLFOREST":            true,
		"XMLITERATE":           true,
		"XMLNAMESPACES":        true,
		"XMLPARSE":             true,
		"XMLPI":                true,
		"XMLQUERY":             true,
		"XMLSERIALIZE":         true,
		"XMLTABLE":             true,
		"XMLTEXT":              true,
		"XMLVALIDATE":          true,
		"YEAR":                 true,
		// additions not in ref
		"ELSIF":   false,
		"FOREACH": false,
		"LOOP":    false,
	}

	v, ok := pgKeywords[strings.ToUpper(s)]

	return ok, v
}

// IsKeyword returns a boolean indicating if the supplied string
// is considered to be a keyword in PostgreSQL
func (d PostgreSQLDialect) IsKeyword(s string) bool {
	isKey, _ := d.keyword(s)
	return isKey
}

// IsReservedKeyword returns a boolean indicating if the supplied
// string is considered to be a reserved keyword in PostgreSQL
func (d PostgreSQLDialect) IsReservedKeyword(s string) bool {
	isKey, isReserved := d.keyword(s)

	if isKey {
		return isReserved
	}
	return false
}

// IsOperator returns a boolean indicating if the supplied string
// is considered to be an operator in PostgreSQL
func (d PostgreSQLDialect) IsOperator(s string) bool {

	var pgOperators = map[string]bool{
		"^":   true,
		"~":   true,
		"~*":  true,
		"<<":  true,
		"<=":  true,
		"<>":  true,
		"<":   true,
		"=":   true,
		">=":  true,
		">>":  true,
		">":   true,
		"||/": true,
		"||":  true,
		"|/":  true,
		"|":   true,
		"-":   true,
		":=":  true,
		"::":  true,
		"!~":  true,
		"!~*": true,
		"!=":  true,
		"!!":  true,
		"!":   true,
		"/":   true,
		"@":   true,
		"*":   true,
		"&":   true,
		"#":   true,
		"%":   true,
		"+":   true,
		"=>":  true, // Added fat comma for function/procedure calls
		"..":  true, // Added for loop ranges
	}

	// For valid operators that come with the system
	if _, ok := pgOperators[s]; ok {
		return true
	}

	/*
		For custom operators, per https://www.postgresql.org/docs/current/sql-createoperator.html:

		The operator name is a sequence of up to NAMEDATALEN-1 (63 by default)
		characters from the following list:

		+ - * / < > = ~ ! @ # % ^ & | ` ?

		There are a few restrictions on your choice of name:

		-- and /* cannot appear anywhere in an operator name, since they will
		be taken as the start of a comment.

		A multicharacter operator name cannot end in + or -, unless the name
		also contains at least one of these characters:

		~ ! @ # % ^ & | ` ?

		For example, @- is an allowed operator name, but *- is not. This
		restriction allows PostgreSQL to parse SQL-compliant commands without
		requiring spaces between tokens.

		The symbol => is reserved by the SQL grammar, so it cannot be used as
		an operator name.
	*/
	var pB string
	bs := []byte(s)
	idxMax := len(bs) - 1
	hasValidMultichar := false

	for idx := 0; idx <= idxMax; idx++ {
		b := string(bs[idx])
		switch b {
		case "+":
			if idx == idxMax && !hasValidMultichar {
				return false
			}
		case "-":
			if pB == "-" {
				// "--" is a comment start
				return false
			}
			if idx == idxMax && !hasValidMultichar {
				return false
			}
		case "*":
			if pB == "/" {
				// "/*" is a comment start
				return false
			}
		case "~", "!", "@", "#", "%", "^", "&", "|", "`", "?":
			hasValidMultichar = true
		case "/", "<", ">", "=":
		// nada
		default:
			return false
		}

		pB = b
	}
	return true
}

// IsLabel returns a boolean indicating if the supplied string
// is considered to be a label in PostgreSQL
func (d PostgreSQLDialect) IsLabel(s string) bool {
	if len(s) < 5 {
		return false
	}
	if s[0:2] != "<<" {
		return false
	}
	if s[len(s)-2:len(s)] != ">>" {
		return false
	}
	if !d.IsIdentifier(s[2 : len(s)-2]) {
		return false
	}
	return true
}

// IsIdentifier returns a boolean indicating if the supplied
// string is considered to be a non-quoted PostgreSQL identifier.
func (d PostgreSQLDialect) IsIdentifier(s string) bool {

	// "SQL identifiers and key words must begin with a letter (a-z, but
	// also letters with diacritical marks and non-Latin letters) or an
	// underscore (_). Subsequent characters in an identifier or key word
	// can be letters, underscores, digits (0-9), or dollar signs ($).
	// Note that dollar signs are not allowed in identifiers according to
	// the letter of the SQL standard, so their use might render
	// applications less portable."

	const firstIdentChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	const identChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$"

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
