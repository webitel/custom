package postgres

import (
	"strings"
	"unicode/utf8"
)

// according to ./service/custom/internal/types/utils.(CodeName) method
// [field] name(s) comform next constraints:
// - cannot start with digits ; (0-9)
// - consists of ASCII chars  ; [_0-9A-Za-z]
//
// [see]: strconv.Quote(s) for more details
func CustomSqlIdentifier(s string) (ident string) {
	// https://www.postgresql.org/docs/17/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
	//
	// Quoting an identifier also makes it case-sensitive,
	// whereas unquoted names are always folded to lower case.
	// For example, the identifiers FOO, foo, and "foo" are considered the same by PostgreSQL,
	// but "Foo" and "FOO" are different from these three and each other.
	const (
		QUOTE = '"'
		DELIM = string(QUOTE)
	)
	if n := len(s); n > 2 && s[0] == QUOTE && s[n-1] == QUOTE {
		s = s[1 : n-1] // unquote
	}
	// Is an SQL Standard KEY WORD ? Needs escaping ?
	escape := sqlstdKeyWordMap[strings.ToUpper(s)]
	for i := 0; !escape && i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			// [TODO]: correct escaping ...
			escape = true
			break
		}
		escape = escape || ('A' <= c && c <= 'Z') || (c == QUOTE) // || (c == '-')
	}
	if escape {
		// To include the escape character in the identifier literally, write it twice.
		s = strings.ReplaceAll(
			s, DELIM, strings.Repeat(DELIM, 2),
		)
		// Delimited -OR- Quoted !
		return DELIM + s + DELIM
	}
	// as is ..
	return s
}

// https://www.postgresql.org/docs/17/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
//
// Key words and identifiers have the same lexical structure, meaning that one cannot know
// whether a token is an identifier or a key word without knowing the language.
//
// A complete list of key words can be found in [Appendix C].
// https://www.postgresql.org/docs/17/sql-keywords-appendix.html
//
// MAP[KEYWORD]RESERVED?
var sqlstdKeyWordMap = map[string]bool{
	"A":                                false, //  	non-reserved	non-reserved
	"ABORT":                            false, // non-reserved
	"ABS":                              false, //  	reserved	reserved
	"ABSENT":                           false, // non-reserved	reserved	reserved
	"ABSOLUTE":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"ACCESS":                           false, // non-reserved
	"ACCORDING":                        false, //  	non-reserved	non-reserved
	"ACOS":                             false, //  	reserved	reserved
	"ACTION":                           false, // non-reserved	non-reserved	non-reserved	reserved
	"ADA":                              false, //  	non-reserved	non-reserved	non-reserved
	"ADD":                              false, // non-reserved	non-reserved	non-reserved	reserved
	"ADMIN":                            false, // non-reserved	non-reserved	non-reserved
	"AFTER":                            false, // non-reserved	non-reserved	non-reserved
	"AGGREGATE":                        false, // non-reserved
	"ALL":                              true,  // reserved	reserved	reserved	reserved
	"ALLOCATE":                         false, //  	reserved	reserved	reserved
	"ALSO":                             false, // non-reserved
	"ALTER":                            false, // non-reserved	reserved	reserved	reserved
	"ALWAYS":                           false, // non-reserved	non-reserved	non-reserved
	"ANALYSE":                          true,  // reserved
	"ANALYZE":                          true,  // reserved
	"AND":                              true,  // reserved	reserved	reserved	reserved
	"ANY":                              true,  // reserved	reserved	reserved	reserved
	"ANY_VALUE":                        false, //  	reserved
	"ARE":                              false, //  	reserved	reserved	reserved
	"ARRAY":                            true,  // reserved, requires AS	reserved	reserved
	"ARRAY_AGG":                        false, //  	reserved	reserved
	"ARRAY_MAX_CARDINALITY":            false, //  	reserved	reserved
	"AS":                               true,  // reserved, requires AS	reserved	reserved	reserved
	"ASC":                              true,  // reserved	non-reserved	non-reserved	reserved
	"ASENSITIVE":                       false, // non-reserved	reserved	reserved
	"ASIN":                             false, //  	reserved	reserved
	"ASSERTION":                        false, // non-reserved	non-reserved	non-reserved	reserved
	"ASSIGNMENT":                       false, // non-reserved	non-reserved	non-reserved
	"ASYMMETRIC":                       true,  // reserved	reserved	reserved
	"AT":                               false, // non-reserved	reserved	reserved	reserved
	"ATAN":                             false, //  	reserved	reserved
	"ATOMIC":                           false, // non-reserved	reserved	reserved
	"ATTACH":                           false, // non-reserved
	"ATTRIBUTE":                        false, // non-reserved	non-reserved	non-reserved
	"ATTRIBUTES":                       false, //  	non-reserved	non-reserved
	"AUTHORIZATION":                    true,  // reserved (can be function or type)	reserved	reserved	reserved
	"AVG":                              false, //  	reserved	reserved	reserved
	"BACKWARD":                         false, // non-reserved
	"BASE64":                           false, //  	non-reserved	non-reserved
	"BEFORE":                           false, // non-reserved	non-reserved	non-reserved
	"BEGIN":                            false, // non-reserved	reserved	reserved	reserved
	"BEGIN_FRAME":                      false, //  	reserved	reserved
	"BEGIN_PARTITION":                  false, //  	reserved	reserved
	"BERNOULLI":                        false, //  	non-reserved	non-reserved
	"BETWEEN":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"BIGINT":                           false, // non-reserved (cannot be function or type)	reserved	reserved
	"BINARY":                           true,  // reserved (can be function or type)	reserved	reserved
	"BIT":                              false, // non-reserved (cannot be function or type)	 	 	reserved
	"BIT_LENGTH":                       false, //  	 	 	reserved
	"BLOB":                             false, //  	reserved	reserved
	"BLOCKED":                          false, //  	non-reserved	non-reserved
	"BOM":                              false, //  	non-reserved	non-reserved
	"BOOLEAN":                          false, // non-reserved (cannot be function or type)	reserved	reserved
	"BOTH":                             true,  // reserved	reserved	reserved	reserved
	"BREADTH":                          false, // non-reserved	non-reserved	non-reserved
	"BTRIM":                            false, //  	reserved
	"BY":                               false, // non-reserved	reserved	reserved	reserved
	"C":                                false, //  	non-reserved	non-reserved	non-reserved
	"CACHE":                            false, // non-reserved
	"CALL":                             false, // non-reserved	reserved	reserved
	"CALLED":                           false, // non-reserved	reserved	reserved
	"CARDINALITY":                      false, //  	reserved	reserved
	"CASCADE":                          false, // non-reserved	non-reserved	non-reserved	reserved
	"CASCADED":                         false, // non-reserved	reserved	reserved	reserved
	"CASE":                             true,  // reserved	reserved	reserved	reserved
	"CAST":                             true,  // reserved	reserved	reserved	reserved
	"CATALOG":                          false, // non-reserved	non-reserved	non-reserved	reserved
	"CATALOG_NAME":                     false, //  	non-reserved	non-reserved	non-reserved
	"CEIL":                             false, //  	reserved	reserved
	"CEILING":                          false, //  	reserved	reserved
	"CHAIN":                            false, // non-reserved	non-reserved	non-reserved
	"CHAINING":                         false, //  	non-reserved	non-reserved
	"CHAR":                             false, // non-reserved (cannot be function or type), requires AS	reserved	reserved	reserved
	"CHARACTER":                        false, // non-reserved (cannot be function or type), requires AS	reserved	reserved	reserved
	"CHARACTERISTICS":                  false, // non-reserved	non-reserved	non-reserved
	"CHARACTERS":                       false, //  	non-reserved	non-reserved
	"CHARACTER_LENGTH":                 false, //  	reserved	reserved	reserved
	"CHARACTER_SET_CATALOG":            false, //  	non-reserved	non-reserved	non-reserved
	"CHARACTER_SET_NAME":               false, //  	non-reserved	non-reserved	non-reserved
	"CHARACTER_SET_SCHEMA":             false, //  	non-reserved	non-reserved	non-reserved
	"CHAR_LENGTH":                      false, //  	reserved	reserved	reserved
	"CHECK":                            true,  // reserved	reserved	reserved	reserved
	"CHECKPOINT":                       false, // non-reserved
	"CLASS":                            false, // non-reserved
	"CLASSIFIER":                       false, //  	reserved	reserved
	"CLASS_ORIGIN":                     false, //  	non-reserved	non-reserved	non-reserved
	"CLOB":                             false, //  	reserved	reserved
	"CLOSE":                            false, // non-reserved	reserved	reserved	reserved
	"CLUSTER":                          false, // non-reserved
	"COALESCE":                         false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"COBOL":                            false, //  	non-reserved	non-reserved	non-reserved
	"COLLATE":                          true,  // reserved	reserved	reserved	reserved
	"COLLATION":                        true,  // reserved (can be function or type)	non-reserved	non-reserved	reserved
	"COLLATION_CATALOG":                false, //  	non-reserved	non-reserved	non-reserved
	"COLLATION_NAME":                   false, //  	non-reserved	non-reserved	non-reserved
	"COLLATION_SCHEMA":                 false, //  	non-reserved	non-reserved	non-reserved
	"COLLECT":                          false, //  	reserved	reserved
	"COLUMN":                           true,  // reserved	reserved	reserved	reserved
	"COLUMNS":                          false, // non-reserved	non-reserved	non-reserved
	"COLUMN_NAME":                      false, //  	non-reserved	non-reserved	non-reserved
	"COMMAND_FUNCTION":                 false, //  	non-reserved	non-reserved	non-reserved
	"COMMAND_FUNCTION_CODE":            false, //  	non-reserved	non-reserved
	"COMMENT":                          false, // non-reserved
	"COMMENTS":                         false, // non-reserved
	"COMMIT":                           false, // non-reserved	reserved	reserved	reserved
	"COMMITTED":                        false, // non-reserved	non-reserved	non-reserved	non-reserved
	"COMPRESSION":                      false, // non-reserved
	"CONCURRENTLY":                     true,  // reserved (can be function or type)
	"CONDITION":                        false, //  	reserved	reserved
	"CONDITIONAL":                      false, // non-reserved	non-reserved	non-reserved
	"CONDITION_NUMBER":                 false, //  	non-reserved	non-reserved	non-reserved
	"CONFIGURATION":                    false, // non-reserved
	"CONFLICT":                         false, // non-reserved
	"CONNECT":                          false, //  	reserved	reserved	reserved
	"CONNECTION":                       false, // non-reserved	non-reserved	non-reserved	reserved
	"CONNECTION_NAME":                  false, //  	non-reserved	non-reserved	non-reserved
	"CONSTRAINT":                       true,  // reserved	reserved	reserved	reserved
	"CONSTRAINTS":                      false, // non-reserved	non-reserved	non-reserved	reserved
	"CONSTRAINT_CATALOG":               false, //  	non-reserved	non-reserved	non-reserved
	"CONSTRAINT_NAME":                  false, //  	non-reserved	non-reserved	non-reserved
	"CONSTRAINT_SCHEMA":                false, //  	non-reserved	non-reserved	non-reserved
	"CONSTRUCTOR":                      false, //  	non-reserved	non-reserved
	"CONTAINS":                         false, //  	reserved	reserved
	"CONTENT":                          false, // non-reserved	non-reserved	non-reserved
	"CONTINUE":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"CONTROL":                          false, //  	non-reserved	non-reserved
	"CONVERSION":                       false, // non-reserved
	"CONVERT":                          false, //  	reserved	reserved	reserved
	"COPARTITION":                      false, //  	non-reserved
	"COPY":                             false, // non-reserved	reserved	reserved
	"CORR":                             false, //  	reserved	reserved
	"CORRESPONDING":                    false, //  	reserved	reserved	reserved
	"COS":                              false, //  	reserved	reserved
	"COSH":                             false, //  	reserved	reserved
	"COST":                             false, // non-reserved
	"COUNT":                            false, //  	reserved	reserved	reserved
	"COVAR_POP":                        false, //  	reserved	reserved
	"COVAR_SAMP":                       false, //  	reserved	reserved
	"CREATE":                           true,  // reserved, requires AS	reserved	reserved	reserved
	"CROSS":                            true,  // reserved (can be function or type)	reserved	reserved	reserved
	"CSV":                              false, // non-reserved
	"CUBE":                             false, // non-reserved	reserved	reserved
	"CUME_DIST":                        false, //  	reserved	reserved
	"CURRENT":                          false, // non-reserved	reserved	reserved	reserved
	"CURRENT_CATALOG":                  true,  // reserved	reserved	reserved
	"CURRENT_DATE":                     true,  // reserved	reserved	reserved	reserved
	"CURRENT_DEFAULT_TRANSFORM_GROUP":  false, //  	reserved	reserved
	"CURRENT_PATH":                     false, //  	reserved	reserved
	"CURRENT_ROLE":                     true,  // reserved	reserved	reserved
	"CURRENT_ROW":                      false, //  	reserved	reserved
	"CURRENT_SCHEMA":                   true,  // reserved (can be function or type)	reserved	reserved
	"CURRENT_TIME":                     true,  // reserved	reserved	reserved	reserved
	"CURRENT_TIMESTAMP":                true,  // reserved	reserved	reserved	reserved
	"CURRENT_TRANSFORM_GROUP_FOR_TYPE": false, //  	reserved	reserved
	"CURRENT_USER":                     true,  // reserved	reserved	reserved	reserved
	"CURSOR":                           false, // non-reserved	reserved	reserved	reserved
	"CURSOR_NAME":                      false, //  	non-reserved	non-reserved	non-reserved
	"CYCLE":                            false, // non-reserved	reserved	reserved
	"DATA":                             false, // non-reserved	non-reserved	non-reserved	non-reserved
	"DATABASE":                         false, // non-reserved
	"DATALINK":                         false, //  	reserved	reserved
	"DATE":                             false, //  	reserved	reserved	reserved
	"DATETIME_INTERVAL_CODE":           false, //  	non-reserved	non-reserved	non-reserved
	"DATETIME_INTERVAL_PRECISION":      false, //  	non-reserved	non-reserved	non-reserved
	"DAY":                              false, // non-reserved, requires AS	reserved	reserved	reserved
	"DB":                               false, //  	non-reserved	non-reserved
	"DEALLOCATE":                       false, // non-reserved	reserved	reserved	reserved
	"DEC":                              false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"DECFLOAT":                         false, //  	reserved	reserved
	"DECIMAL":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"DECLARE":                          false, // non-reserved	reserved	reserved	reserved
	"DEFAULT":                          true,  // reserved	reserved	reserved	reserved
	"DEFAULTS":                         false, // non-reserved	non-reserved	non-reserved
	"DEFERRABLE":                       true,  // reserved	non-reserved	non-reserved	reserved
	"DEFERRED":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"DEFINE":                           false, //  	reserved	reserved
	"DEFINED":                          false, //  	non-reserved	non-reserved
	"DEFINER":                          false, // non-reserved	non-reserved	non-reserved
	"DEGREE":                           false, //  	non-reserved	non-reserved
	"DELETE":                           false, // non-reserved	reserved	reserved	reserved
	"DELIMITER":                        false, // non-reserved
	"DELIMITERS":                       false, // non-reserved
	"DENSE_RANK":                       false, //  	reserved	reserved
	"DEPENDS":                          false, // non-reserved
	"DEPTH":                            false, // non-reserved	non-reserved	non-reserved
	"DEREF":                            false, //  	reserved	reserved
	"DERIVED":                          false, //  	non-reserved	non-reserved
	"DESC":                             true,  // reserved	non-reserved	non-reserved	reserved
	"DESCRIBE":                         false, //  	reserved	reserved	reserved
	"DESCRIPTOR":                       false, //  	non-reserved	non-reserved	reserved
	"DETACH":                           false, // non-reserved
	"DETERMINISTIC":                    false, //  	reserved	reserved
	"DIAGNOSTICS":                      false, //  	non-reserved	non-reserved	reserved
	"DICTIONARY":                       false, // non-reserved
	"DISABLE":                          false, // non-reserved
	"DISCARD":                          false, // non-reserved
	"DISCONNECT":                       false, //  	reserved	reserved	reserved
	"DISPATCH":                         false, //  	non-reserved	non-reserved
	"DISTINCT":                         true,  // reserved	reserved	reserved	reserved
	"DLNEWCOPY":                        false, //  	reserved	reserved
	"DLPREVIOUSCOPY":                   false, //  	reserved	reserved
	"DLURLCOMPLETE":                    false, //  	reserved	reserved
	"DLURLCOMPLETEONLY":                false, //  	reserved	reserved
	"DLURLCOMPLETEWRITE":               false, //  	reserved	reserved
	"DLURLPATH":                        false, //  	reserved	reserved
	"DLURLPATHONLY":                    false, //  	reserved	reserved
	"DLURLPATHWRITE":                   false, //  	reserved	reserved
	"DLURLSCHEME":                      false, //  	reserved	reserved
	"DLURLSERVER":                      false, //  	reserved	reserved
	"DLVALUE":                          false, //  	reserved	reserved
	"DO":                               true,  // reserved
	"DOCUMENT":                         false, // non-reserved	non-reserved	non-reserved
	"DOMAIN":                           false, // non-reserved	non-reserved	non-reserved	reserved
	"DOUBLE":                           false, // non-reserved	reserved	reserved	reserved
	"DROP":                             false, // non-reserved	reserved	reserved	reserved
	"DYNAMIC":                          false, //  	reserved	reserved
	"DYNAMIC_FUNCTION":                 false, //  	non-reserved	non-reserved	non-reserved
	"DYNAMIC_FUNCTION_CODE":            false, //  	non-reserved	non-reserved
	"EACH":                             false, // non-reserved	reserved	reserved
	"ELEMENT":                          false, //  	reserved	reserved
	"ELSE":                             true,  // reserved	reserved	reserved	reserved
	"EMPTY":                            false, // non-reserved	reserved	reserved
	"ENABLE":                           false, // non-reserved
	"ENCODING":                         false, // non-reserved	non-reserved	non-reserved
	"ENCRYPTED":                        false, // non-reserved
	"END":                              true,  // reserved	reserved	reserved	reserved
	"END-EXEC":                         true,  // reserved	reserved	reserved
	"END_FRAME":                        false, //  	reserved	reserved
	"END_PARTITION":                    false, //  	reserved	reserved
	"ENFORCED":                         false, //  	non-reserved	non-reserved
	"ENUM":                             false, // non-reserved
	"EQUALS":                           false, //  	reserved	reserved
	"ERROR":                            false, // non-reserved	non-reserved	non-reserved
	"ESCAPE":                           false, // non-reserved	reserved	reserved	reserved
	"EVENT":                            false, // non-reserved
	"EVERY":                            false, //  	reserved	reserved
	"EXCEPT":                           true,  // reserved, requires AS	reserved	reserved	reserved
	"EXCEPTION":                        false, //  	 	 	reserved
	"EXCLUDE":                          false, // non-reserved	non-reserved	non-reserved
	"EXCLUDING":                        false, // non-reserved	non-reserved	non-reserved
	"EXCLUSIVE":                        false, // non-reserved
	"EXEC":                             false, //  	reserved	reserved	reserved
	"EXECUTE":                          false, // non-reserved	reserved	reserved	reserved
	"EXISTS":                           false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"EXP":                              false, //  	reserved	reserved
	"EXPLAIN":                          false, // non-reserved
	"EXPRESSION":                       false, // non-reserved	non-reserved	non-reserved
	"EXTENSION":                        false, // non-reserved
	"EXTERNAL":                         false, // non-reserved	reserved	reserved	reserved
	"EXTRACT":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"FALSE":                            true,  // reserved	reserved	reserved	reserved
	"FAMILY":                           false, // non-reserved
	"FETCH":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"FILE":                             false, //  	non-reserved	non-reserved
	"FILTER":                           false, // non-reserved, requires AS	reserved	reserved
	"FINAL":                            false, //  	non-reserved	non-reserved
	"FINALIZE":                         false, // non-reserved
	"FINISH":                           false, //  	non-reserved	non-reserved
	"FIRST":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"FIRST_VALUE":                      false, //  	reserved	reserved
	"FLAG":                             false, //  	non-reserved	non-reserved
	"FLOAT":                            false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"FLOOR":                            false, //  	reserved	reserved
	"FOLLOWING":                        false, // non-reserved	non-reserved	non-reserved
	"FOR":                              true,  // reserved, requires AS	reserved	reserved	reserved
	"FORCE":                            false, // non-reserved
	"FOREIGN":                          true,  // reserved	reserved	reserved	reserved
	"FORMAT":                           false, // non-reserved	non-reserved	non-reserved
	"FORTRAN":                          false, //  	non-reserved	non-reserved	non-reserved
	"FORWARD":                          false, // non-reserved
	"FOUND":                            false, //  	non-reserved	non-reserved	reserved
	"FRAME_ROW":                        false, //  	reserved	reserved
	"FREE":                             false, //  	reserved	reserved
	"FREEZE":                           true,  // reserved (can be function or type)
	"FROM":                             true,  // reserved, requires AS	reserved	reserved	reserved
	"FS":                               false, //  	non-reserved	non-reserved
	"FULFILL":                          false, //  	non-reserved	non-reserved
	"FULL":                             true,  // reserved (can be function or type)	reserved	reserved	reserved
	"FUNCTION":                         false, // non-reserved	reserved	reserved
	"FUNCTIONS":                        false, // non-reserved
	"FUSION":                           false, //  	reserved	reserved
	"G":                                false, //  	non-reserved	non-reserved
	"GENERAL":                          false, //  	non-reserved	non-reserved
	"GENERATED":                        false, // non-reserved	non-reserved	non-reserved
	"GET":                              false, //  	reserved	reserved	reserved
	"GLOBAL":                           false, // non-reserved	reserved	reserved	reserved
	"GO":                               false, //  	non-reserved	non-reserved	reserved
	"GOTO":                             false, //  	non-reserved	non-reserved	reserved
	"GRANT":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"GRANTED":                          false, // non-reserved	non-reserved	non-reserved
	"GREATEST":                         false, // non-reserved (cannot be function or type)	reserved
	"GROUP":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"GROUPING":                         false, // non-reserved (cannot be function or type)	reserved	reserved
	"GROUPS":                           false, // non-reserved	reserved	reserved
	"HANDLER":                          false, // non-reserved
	"HAVING":                           true,  // reserved, requires AS	reserved	reserved	reserved
	"HEADER":                           false, // non-reserved
	"HEX":                              false, //  	non-reserved	non-reserved
	"HIERARCHY":                        false, //  	non-reserved	non-reserved
	"HOLD":                             false, // non-reserved	reserved	reserved
	"HOUR":                             false, // non-reserved, requires AS	reserved	reserved	reserved
	"ID":                               false, //  	non-reserved	non-reserved
	"IDENTITY":                         false, // non-reserved	reserved	reserved	reserved
	"IF":                               false, // non-reserved
	"IGNORE":                           false, //  	non-reserved	non-reserved
	"ILIKE":                            true,  // reserved (can be function or type)
	"IMMEDIATE":                        false, // non-reserved	non-reserved	non-reserved	reserved
	"IMMEDIATELY":                      false, //  	non-reserved	non-reserved
	"IMMUTABLE":                        false, // non-reserved
	"IMPLEMENTATION":                   false, //  	non-reserved	non-reserved
	"IMPLICIT":                         false, // non-reserved
	"IMPORT":                           false, // non-reserved	reserved	reserved
	"IN":                               true,  // reserved	reserved	reserved	reserved
	"INCLUDE":                          false, // non-reserved
	"INCLUDING":                        false, // non-reserved	non-reserved	non-reserved
	"INCREMENT":                        false, // non-reserved	non-reserved	non-reserved
	"INDENT":                           false, // non-reserved	non-reserved	non-reserved
	"INDEX":                            false, // non-reserved
	"INDEXES":                          false, // non-reserved
	"INDICATOR":                        false, //  	reserved	reserved	reserved
	"INHERIT":                          false, // non-reserved
	"INHERITS":                         false, // non-reserved
	"INITIAL":                          false, //  	reserved	reserved
	"INITIALLY":                        true,  // reserved	non-reserved	non-reserved	reserved
	"INLINE":                           false, // non-reserved
	"INNER":                            true,  // reserved (can be function or type)	reserved	reserved	reserved
	"INOUT":                            false, // non-reserved (cannot be function or type)	reserved	reserved
	"INPUT":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"INSENSITIVE":                      false, // non-reserved	reserved	reserved	reserved
	"INSERT":                           false, // non-reserved	reserved	reserved	reserved
	"INSTANCE":                         false, //  	non-reserved	non-reserved
	"INSTANTIABLE":                     false, //  	non-reserved	non-reserved
	"INSTEAD":                          false, // non-reserved	non-reserved	non-reserved
	"INT":                              false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"INTEGER":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"INTEGRITY":                        false, //  	non-reserved	non-reserved
	"INTERSECT":                        true,  // reserved, requires AS	reserved	reserved	reserved
	"INTERSECTION":                     false, //  	reserved	reserved
	"INTERVAL":                         false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"INTO":                             true,  // reserved, requires AS	reserved	reserved	reserved
	"INVOKER":                          false, // non-reserved	non-reserved	non-reserved
	"IS":                               true,  // reserved (can be function or type)	reserved	reserved	reserved
	"ISNULL":                           true,  // reserved (can be function or type), requires AS
	"ISOLATION":                        false, // non-reserved	non-reserved	non-reserved	reserved
	"JOIN":                             true,  // reserved (can be function or type)	reserved	reserved	reserved
	"JSON":                             false, // non-reserved (cannot be function or type)	reserved
	"JSON_ARRAY":                       false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_ARRAYAGG":                    false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_EXISTS":                      false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_OBJECT":                      false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_OBJECTAGG":                   false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_QUERY":                       false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_SCALAR":                      false, // non-reserved (cannot be function or type)	reserved
	"JSON_SERIALIZE":                   false, // non-reserved (cannot be function or type)	reserved
	"JSON_TABLE":                       false, // non-reserved (cannot be function or type)	reserved	reserved
	"JSON_TABLE_PRIMITIVE":             false, //  	reserved	reserved
	"JSON_VALUE":                       false, // non-reserved (cannot be function or type)	reserved	reserved
	"K":                                false, //  	non-reserved	non-reserved
	"KEEP":                             false, // non-reserved	non-reserved	non-reserved
	"KEY":                              false, // non-reserved	non-reserved	non-reserved	reserved
	"KEYS":                             false, // non-reserved	non-reserved	non-reserved
	"KEY_MEMBER":                       false, //  	non-reserved	non-reserved
	"KEY_TYPE":                         false, //  	non-reserved	non-reserved
	"LABEL":                            false, // non-reserved
	"LAG":                              false, //  	reserved	reserved
	"LANGUAGE":                         false, // non-reserved	reserved	reserved	reserved
	"LARGE":                            false, // non-reserved	reserved	reserved
	"LAST":                             false, // non-reserved	non-reserved	non-reserved	reserved
	"LAST_VALUE":                       false, //  	reserved	reserved
	"LATERAL":                          true,  // reserved	reserved	reserved
	"LEAD":                             false, //  	reserved	reserved
	"LEADING":                          true,  // reserved	reserved	reserved	reserved
	"LEAKPROOF":                        false, // non-reserved
	"LEAST":                            false, // non-reserved (cannot be function or type)	reserved
	"LEFT":                             true,  // reserved (can be function or type)	reserved	reserved	reserved
	"LENGTH":                           false, //  	non-reserved	non-reserved	non-reserved
	"LEVEL":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"LIBRARY":                          false, //  	non-reserved	non-reserved
	"LIKE":                             true,  // reserved (can be function or type)	reserved	reserved	reserved
	"LIKE_REGEX":                       false, //  	reserved	reserved
	"LIMIT":                            true,  // reserved, requires AS	non-reserved	non-reserved
	"LINK":                             false, //  	non-reserved	non-reserved
	"LISTAGG":                          false, //  	reserved	reserved
	"LISTEN":                           false, // non-reserved
	"LN":                               false, //  	reserved	reserved
	"LOAD":                             false, // non-reserved
	"LOCAL":                            false, // non-reserved	reserved	reserved	reserved
	"LOCALTIME":                        true,  // reserved	reserved	reserved
	"LOCALTIMESTAMP":                   true,  // reserved	reserved	reserved
	"LOCATION":                         false, // non-reserved	non-reserved	non-reserved
	"LOCATOR":                          false, //  	non-reserved	non-reserved
	"LOCK":                             false, // non-reserved
	"LOCKED":                           false, // non-reserved
	"LOG":                              false, //  	reserved	reserved
	"LOG10":                            false, //  	reserved	reserved
	"LOGGED":                           false, // non-reserved
	"LOWER":                            false, //  	reserved	reserved	reserved
	"LPAD":                             false, //  	reserved
	"LTRIM":                            false, //  	reserved
	"M":                                false, //  	non-reserved	non-reserved
	"MAP":                              false, //  	non-reserved	non-reserved
	"MAPPING":                          false, // non-reserved	non-reserved	non-reserved
	"MATCH":                            false, // non-reserved	reserved	reserved	reserved
	"MATCHED":                          false, // non-reserved	non-reserved	non-reserved
	"MATCHES":                          false, //  	reserved	reserved
	"MATCH_NUMBER":                     false, //  	reserved	reserved
	"MATCH_RECOGNIZE":                  false, //  	reserved	reserved
	"MATERIALIZED":                     false, // non-reserved
	"MAX":                              false, //  	reserved	reserved	reserved
	"MAXVALUE":                         false, // non-reserved	non-reserved	non-reserved
	"MEASURES":                         false, //  	non-reserved	non-reserved
	"MEMBER":                           false, //  	reserved	reserved
	"MERGE":                            false, // non-reserved	reserved	reserved
	"MERGE_ACTION":                     false, // non-reserved (cannot be function or type)
	"MESSAGE_LENGTH":                   false, //  	non-reserved	non-reserved	non-reserved
	"MESSAGE_OCTET_LENGTH":             false, //  	non-reserved	non-reserved	non-reserved
	"MESSAGE_TEXT":                     false, //  	non-reserved	non-reserved	non-reserved
	"METHOD":                           false, // non-reserved	reserved	reserved
	"MIN":                              false, //  	reserved	reserved	reserved
	"MINUTE":                           false, // non-reserved, requires AS	reserved	reserved	reserved
	"MINVALUE":                         false, // non-reserved	non-reserved	non-reserved
	"MOD":                              false, //  	reserved	reserved
	"MODE":                             false, // non-reserved
	"MODIFIES":                         false, //  	reserved	reserved
	"MODULE":                           false, //  	reserved	reserved	reserved
	"MONTH":                            false, // non-reserved, requires AS	reserved	reserved	reserved
	"MORE":                             false, //  	non-reserved	non-reserved	non-reserved
	"MOVE":                             false, // non-reserved
	"MULTISET":                         false, //  	reserved	reserved
	"MUMPS":                            false, //  	non-reserved	non-reserved	non-reserved
	"NAME":                             false, // non-reserved	non-reserved	non-reserved	non-reserved
	"NAMES":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"NAMESPACE":                        false, //  	non-reserved	non-reserved
	"NATIONAL":                         false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"NATURAL":                          true,  // reserved (can be function or type)	reserved	reserved	reserved
	"NCHAR":                            false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"NCLOB":                            false, //  	reserved	reserved
	"NESTED":                           false, // non-reserved	non-reserved	non-reserved
	"NESTING":                          false, //  	non-reserved	non-reserved
	"NEW":                              false, // non-reserved	reserved	reserved
	"NEXT":                             false, // non-reserved	non-reserved	non-reserved	reserved
	"NFC":                              false, // non-reserved	non-reserved	non-reserved
	"NFD":                              false, // non-reserved	non-reserved	non-reserved
	"NFKC":                             false, // non-reserved	non-reserved	non-reserved
	"NFKD":                             false, // non-reserved	non-reserved	non-reserved
	"NIL":                              false, //  	non-reserved	non-reserved
	"NO":                               false, // non-reserved	reserved	reserved	reserved
	"NONE":                             false, // non-reserved (cannot be function or type)	reserved	reserved
	"NORMALIZE":                        false, // non-reserved (cannot be function or type)	reserved	reserved
	"NORMALIZED":                       false, // non-reserved	non-reserved	non-reserved
	"NOT":                              true,  // reserved	reserved	reserved	reserved
	"NOTHING":                          false, // non-reserved
	"NOTIFY":                           false, // non-reserved
	"NOTNULL":                          true,  // reserved (can be function or type), requires AS
	"NOWAIT":                           false, // non-reserved
	"NTH_VALUE":                        false, //  	reserved	reserved
	"NTILE":                            false, //  	reserved	reserved
	"NULL":                             true,  // reserved	reserved	reserved	reserved
	"NULLABLE":                         false, //  	non-reserved	non-reserved	non-reserved
	"NULLIF":                           false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"NULLS":                            false, // non-reserved	non-reserved	non-reserved
	"NULL_ORDERING":                    false, //  	non-reserved	non-reserved
	"NUMBER":                           false, //  	non-reserved	non-reserved	non-reserved
	"NUMERIC":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"OBJECT":                           false, // non-reserved	non-reserved	non-reserved
	"OCCURRENCE":                       false, //  	non-reserved	non-reserved
	"OCCURRENCES_REGEX":                false, //  	reserved	reserved
	"OCTETS":                           false, //  	non-reserved	non-reserved
	"OCTET_LENGTH":                     false, //  	reserved	reserved	reserved
	"OF":                               false, // non-reserved	reserved	reserved	reserved
	"OFF":                              false, // non-reserved	non-reserved	non-reserved
	"OFFSET":                           true,  // reserved, requires AS	reserved	reserved
	"OIDS":                             false, // non-reserved
	"OLD":                              false, // non-reserved	reserved	reserved
	"OMIT":                             false, // non-reserved	reserved	reserved
	"ON":                               true,  // reserved, requires AS	reserved	reserved	reserved
	"ONE":                              false, //  	reserved	reserved
	"ONLY":                             true,  // reserved	reserved	reserved	reserved
	"OPEN":                             false, //  	reserved	reserved	reserved
	"OPERATOR":                         false, // non-reserved
	"OPTION":                           false, // non-reserved	non-reserved	non-reserved	reserved
	"OPTIONS":                          false, // non-reserved	non-reserved	non-reserved
	"OR":                               true,  // reserved	reserved	reserved	reserved
	"ORDER":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"ORDERING":                         false, //  	non-reserved	non-reserved
	"ORDINALITY":                       false, // non-reserved	non-reserved	non-reserved
	"OTHERS":                           false, // non-reserved	non-reserved	non-reserved
	"OUT":                              false, // non-reserved (cannot be function or type)	reserved	reserved
	"OUTER":                            true,  // reserved (can be function or type)	reserved	reserved	reserved
	"OUTPUT":                           false, //  	non-reserved	non-reserved	reserved
	"OVER":                             false, // non-reserved, requires AS	reserved	reserved
	"OVERFLOW":                         false, //  	non-reserved	non-reserved
	"OVERLAPS":                         true,  // reserved (can be function or type), requires AS	reserved	reserved	reserved
	"OVERLAY":                          false, // non-reserved (cannot be function or type)	reserved	reserved
	"OVERRIDING":                       false, // non-reserved	non-reserved	non-reserved
	"OWNED":                            false, // non-reserved
	"OWNER":                            false, // non-reserved
	"P":                                false, //  	non-reserved	non-reserved
	"PAD":                              false, //  	non-reserved	non-reserved	reserved
	"PARALLEL":                         false, // non-reserved
	"PARAMETER":                        false, // non-reserved	reserved	reserved
	"PARAMETER_MODE":                   false, //  	non-reserved	non-reserved
	"PARAMETER_NAME":                   false, //  	non-reserved	non-reserved
	"PARAMETER_ORDINAL_POSITION":       false, //  	non-reserved	non-reserved
	"PARAMETER_SPECIFIC_CATALOG":       false, //  	non-reserved	non-reserved
	"PARAMETER_SPECIFIC_NAME":          false, //  	non-reserved	non-reserved
	"PARAMETER_SPECIFIC_SCHEMA":        false, //  	non-reserved	non-reserved
	"PARSER":                           false, // non-reserved
	"PARTIAL":                          false, // non-reserved	non-reserved	non-reserved	reserved
	"PARTITION":                        false, // non-reserved	reserved	reserved
	"PASCAL":                           false, //  	non-reserved	non-reserved	non-reserved
	"PASS":                             false, //  	non-reserved	non-reserved
	"PASSING":                          false, // non-reserved	non-reserved	non-reserved
	"PASSTHROUGH":                      false, //  	non-reserved	non-reserved
	"PASSWORD":                         false, // non-reserved
	"PAST":                             false, //  	non-reserved	non-reserved
	"PATH":                             false, // non-reserved	non-reserved	non-reserved
	"PATTERN":                          false, //  	reserved	reserved
	"PER":                              false, //  	reserved	reserved
	"PERCENT":                          false, //  	reserved	reserved
	"PERCENTILE_CONT":                  false, //  	reserved	reserved
	"PERCENTILE_DISC":                  false, //  	reserved	reserved
	"PERCENT_RANK":                     false, //  	reserved	reserved
	"PERIOD":                           false, //  	reserved	reserved
	"PERMISSION":                       false, //  	non-reserved	non-reserved
	"PERMUTE":                          false, //  	non-reserved	non-reserved
	"PIPE":                             false, //  	non-reserved	non-reserved
	"PLACING":                          true,  // reserved	non-reserved	non-reserved
	"PLAN":                             false, // non-reserved	non-reserved	non-reserved
	"PLANS":                            false, // non-reserved
	"PLI":                              false, //  	non-reserved	non-reserved	non-reserved
	"POLICY":                           false, // non-reserved
	"PORTION":                          false, //  	reserved	reserved
	"POSITION":                         false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"POSITION_REGEX":                   false, //  	reserved	reserved
	"POWER":                            false, //  	reserved	reserved
	"PRECEDES":                         false, //  	reserved	reserved
	"PRECEDING":                        false, // non-reserved	non-reserved	non-reserved
	"PRECISION":                        false, // non-reserved (cannot be function or type), requires AS	reserved	reserved	reserved
	"PREPARE":                          false, // non-reserved	reserved	reserved	reserved
	"PREPARED":                         false, // non-reserved
	"PRESERVE":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"PREV":                             false, //  	non-reserved	non-reserved
	"PRIMARY":                          true,  // reserved	reserved	reserved	reserved
	"PRIOR":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"PRIVATE":                          false, //  	non-reserved	non-reserved
	"PRIVILEGES":                       false, // non-reserved	non-reserved	non-reserved	reserved
	"PROCEDURAL":                       false, // non-reserved
	"PROCEDURE":                        false, // non-reserved	reserved	reserved	reserved
	"PROCEDURES":                       false, // non-reserved
	"PROGRAM":                          false, // non-reserved
	"PRUNE":                            false, //  	non-reserved	non-reserved
	"PTF":                              false, //  	reserved	reserved
	"PUBLIC":                           false, //  	non-reserved	non-reserved	reserved
	"PUBLICATION":                      false, // non-reserved
	"QUOTE":                            false, // non-reserved
	"QUOTES":                           false, // non-reserved	non-reserved	non-reserved
	"RANGE":                            false, // non-reserved	reserved	reserved
	"RANK":                             false, //  	reserved	reserved
	"READ":                             false, // non-reserved	non-reserved	non-reserved	reserved
	"READS":                            false, //  	reserved	reserved
	"REAL":                             false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"REASSIGN":                         false, // non-reserved
	"RECHECK":                          false, // non-reserved
	"RECOVERY":                         false, //  	non-reserved	non-reserved
	"RECURSIVE":                        false, // non-reserved	reserved	reserved
	"REF":                              false, // non-reserved	reserved	reserved
	"REFERENCES":                       true,  // reserved	reserved	reserved	reserved
	"REFERENCING":                      false, // non-reserved	reserved	reserved
	"REFRESH":                          false, // non-reserved
	"REGR_AVGX":                        false, //  	reserved	reserved
	"REGR_AVGY":                        false, //  	reserved	reserved
	"REGR_COUNT":                       false, //  	reserved	reserved
	"REGR_INTERCEPT":                   false, //  	reserved	reserved
	"REGR_R2":                          false, //  	reserved	reserved
	"REGR_SLOPE":                       false, //  	reserved	reserved
	"REGR_SXX":                         false, //  	reserved	reserved
	"REGR_SXY":                         false, //  	reserved	reserved
	"REGR_SYY":                         false, //  	reserved	reserved
	"REINDEX":                          false, // non-reserved
	"RELATIVE":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"RELEASE":                          false, // non-reserved	reserved	reserved
	"RENAME":                           false, // non-reserved
	"REPEATABLE":                       false, // non-reserved	non-reserved	non-reserved	non-reserved
	"REPLACE":                          false, // non-reserved
	"REPLICA":                          false, // non-reserved
	"REQUIRING":                        false, //  	non-reserved	non-reserved
	"RESET":                            false, // non-reserved
	"RESPECT":                          false, //  	non-reserved	non-reserved
	"RESTART":                          false, // non-reserved	non-reserved	non-reserved
	"RESTORE":                          false, //  	non-reserved	non-reserved
	"RESTRICT":                         false, // non-reserved	non-reserved	non-reserved	reserved
	"RESULT":                           false, //  	reserved	reserved
	"RETURN":                           false, // non-reserved	reserved	reserved
	"RETURNED_CARDINALITY":             false, //  	non-reserved	non-reserved
	"RETURNED_LENGTH":                  false, //  	non-reserved	non-reserved	non-reserved
	"RETURNED_OCTET_LENGTH":            false, //  	non-reserved	non-reserved	non-reserved
	"RETURNED_SQLSTATE":                false, //  	non-reserved	non-reserved	non-reserved
	"RETURNING":                        true,  // reserved, requires AS	non-reserved	non-reserved
	"RETURNS":                          false, // non-reserved	reserved	reserved
	"REVOKE":                           false, // non-reserved	reserved	reserved	reserved
	"RIGHT":                            true,  // reserved (can be function or type)	reserved	reserved	reserved
	"ROLE":                             false, // non-reserved	non-reserved	non-reserved
	"ROLLBACK":                         false, // non-reserved	reserved	reserved	reserved
	"ROLLUP":                           false, // non-reserved	reserved	reserved
	"ROUTINE":                          false, // non-reserved	non-reserved	non-reserved
	"ROUTINES":                         false, // non-reserved
	"ROUTINE_CATALOG":                  false, //  	non-reserved	non-reserved
	"ROUTINE_NAME":                     false, //  	non-reserved	non-reserved
	"ROUTINE_SCHEMA":                   false, //  	non-reserved	non-reserved
	"ROW":                              false, // non-reserved (cannot be function or type)	reserved	reserved
	"ROWS":                             false, // non-reserved	reserved	reserved	reserved
	"ROW_COUNT":                        false, //  	non-reserved	non-reserved	non-reserved
	"ROW_NUMBER":                       false, //  	reserved	reserved
	"RPAD":                             false, //  	reserved
	"RTRIM":                            false, //  	reserved
	"RULE":                             false, // non-reserved
	"RUNNING":                          false, //  	reserved	reserved
	"SAVEPOINT":                        false, // non-reserved	reserved	reserved
	"SCALAR":                           false, // non-reserved	non-reserved	non-reserved
	"SCALE":                            false, //  	non-reserved	non-reserved	non-reserved
	"SCHEMA":                           false, // non-reserved	non-reserved	non-reserved	reserved
	"SCHEMAS":                          false, // non-reserved
	"SCHEMA_NAME":                      false, //  	non-reserved	non-reserved	non-reserved
	"SCOPE":                            false, //  	reserved	reserved
	"SCOPE_CATALOG":                    false, //  	non-reserved	non-reserved
	"SCOPE_NAME":                       false, //  	non-reserved	non-reserved
	"SCOPE_SCHEMA":                     false, //  	non-reserved	non-reserved
	"SCROLL":                           false, // non-reserved	reserved	reserved	reserved
	"SEARCH":                           false, // non-reserved	reserved	reserved
	"SECOND":                           false, // non-reserved, requires AS	reserved	reserved	reserved
	"SECTION":                          false, //  	non-reserved	non-reserved	reserved
	"SECURITY":                         false, // non-reserved	non-reserved	non-reserved
	"SEEK":                             false, //  	reserved	reserved
	"SELECT":                           true,  // reserved	reserved	reserved	reserved
	"SELECTIVE":                        false, //  	non-reserved	non-reserved
	"SELF":                             false, //  	non-reserved	non-reserved
	"SEMANTICS":                        false, //  	non-reserved	non-reserved
	"SENSITIVE":                        false, //  	reserved	reserved
	"SEQUENCE":                         false, // non-reserved	non-reserved	non-reserved
	"SEQUENCES":                        false, // non-reserved
	"SERIALIZABLE":                     false, // non-reserved	non-reserved	non-reserved	non-reserved
	"SERVER":                           false, // non-reserved	non-reserved	non-reserved
	"SERVER_NAME":                      false, //  	non-reserved	non-reserved	non-reserved
	"SESSION":                          false, // non-reserved	non-reserved	non-reserved	reserved
	"SESSION_USER":                     true,  // reserved	reserved	reserved	reserved
	"SET":                              false, // non-reserved	reserved	reserved	reserved
	"SETOF":                            false, // non-reserved (cannot be function or type)
	"SETS":                             false, // non-reserved	non-reserved	non-reserved
	"SHARE":                            false, // non-reserved
	"SHOW":                             false, // non-reserved	reserved	reserved
	"SIMILAR":                          true,  // reserved (can be function or type)	reserved	reserved
	"SIMPLE":                           false, // non-reserved	non-reserved	non-reserved
	"SIN":                              false, //  	reserved	reserved
	"SINH":                             false, //  	reserved	reserved
	"SIZE":                             false, //  	non-reserved	non-reserved	reserved
	"SKIP":                             false, // non-reserved	reserved	reserved
	"SMALLINT":                         false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"SNAPSHOT":                         false, // non-reserved
	"SOME":                             true,  // reserved	reserved	reserved	reserved
	"SORT_DIRECTION":                   false, //  	non-reserved	non-reserved
	"SOURCE":                           false, // non-reserved	non-reserved	non-reserved
	"SPACE":                            false, //  	non-reserved	non-reserved	reserved
	"SPECIFIC":                         false, //  	reserved	reserved
	"SPECIFICTYPE":                     false, //  	reserved	reserved
	"SPECIFIC_NAME":                    false, //  	non-reserved	non-reserved
	"SQL":                              false, // non-reserved	reserved	reserved	reserved
	"SQLCODE":                          false, //  	 	 	reserved
	"SQLERROR":                         false, //  	 	 	reserved
	"SQLEXCEPTION":                     false, //  	reserved	reserved
	"SQLSTATE":                         false, //  	reserved	reserved	reserved
	"SQLWARNING":                       false, //  	reserved	reserved
	"SQRT":                             false, //  	reserved	reserved
	"STABLE":                           false, // non-reserved
	"STANDALONE":                       false, // non-reserved	non-reserved	non-reserved
	"START":                            false, // non-reserved	reserved	reserved
	"STATE":                            false, //  	non-reserved	non-reserved
	"STATEMENT":                        false, // non-reserved	non-reserved	non-reserved
	"STATIC":                           false, //  	reserved	reserved
	"STATISTICS":                       false, // non-reserved
	"STDDEV_POP":                       false, //  	reserved	reserved
	"STDDEV_SAMP":                      false, //  	reserved	reserved
	"STDIN":                            false, // non-reserved
	"STDOUT":                           false, // non-reserved
	"STORAGE":                          false, // non-reserved
	"STORED":                           false, // non-reserved
	"STRICT":                           false, // non-reserved
	"STRING":                           false, // non-reserved	non-reserved	non-reserved
	"STRIP":                            false, // non-reserved	non-reserved	non-reserved
	"STRUCTURE":                        false, //  	non-reserved	non-reserved
	"STYLE":                            false, //  	non-reserved	non-reserved
	"SUBCLASS_ORIGIN":                  false, //  	non-reserved	non-reserved	non-reserved
	"SUBMULTISET":                      false, //  	reserved	reserved
	"SUBSCRIPTION":                     false, // non-reserved
	"SUBSET":                           false, //  	reserved	reserved
	"SUBSTRING":                        false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"SUBSTRING_REGEX":                  false, //  	reserved	reserved
	"SUCCEEDS":                         false, //  	reserved	reserved
	"SUM":                              false, //  	reserved	reserved	reserved
	"SUPPORT":                          false, // non-reserved
	"SYMMETRIC":                        true,  // reserved	reserved	reserved
	"SYSID":                            false, // non-reserved
	"SYSTEM":                           false, // non-reserved	reserved	reserved
	"SYSTEM_TIME":                      false, //  	reserved	reserved
	"SYSTEM_USER":                      true,  // reserved	reserved	reserved	reserved
	"T":                                false, //  	non-reserved	non-reserved
	"TABLE":                            true,  // reserved	reserved	reserved	reserved
	"TABLES":                           false, // non-reserved
	"TABLESAMPLE":                      true,  // reserved (can be function or type)	reserved	reserved
	"TABLESPACE":                       false, // non-reserved
	"TABLE_NAME":                       false, //  	non-reserved	non-reserved	non-reserved
	"TAN":                              false, //  	reserved	reserved
	"TANH":                             false, //  	reserved	reserved
	"TARGET":                           false, // non-reserved
	"TEMP":                             false, // non-reserved
	"TEMPLATE":                         false, // non-reserved
	"TEMPORARY":                        false, // non-reserved	non-reserved	non-reserved	reserved
	"TEXT":                             false, // non-reserved
	"THEN":                             true,  // reserved	reserved	reserved	reserved
	"THROUGH":                          false, //  	non-reserved	non-reserved
	"TIES":                             false, // non-reserved	non-reserved	non-reserved
	"TIME":                             false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"TIMESTAMP":                        false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"TIMEZONE_HOUR":                    false, //  	reserved	reserved	reserved
	"TIMEZONE_MINUTE":                  false, //  	reserved	reserved	reserved
	"TO":                               true,  // reserved, requires AS	reserved	reserved	reserved
	"TOKEN":                            false, //  	non-reserved	non-reserved
	"TOP_LEVEL_COUNT":                  false, //  	non-reserved	non-reserved
	"TRAILING":                         true,  // reserved	reserved	reserved	reserved
	"TRANSACTION":                      false, // non-reserved	non-reserved	non-reserved	reserved
	"TRANSACTIONS_COMMITTED":           false, //  	non-reserved	non-reserved
	"TRANSACTIONS_ROLLED_BACK":         false, //  	non-reserved	non-reserved
	"TRANSACTION_ACTIVE":               false, //  	non-reserved	non-reserved
	"TRANSFORM":                        false, // non-reserved	non-reserved	non-reserved
	"TRANSFORMS":                       false, //  	non-reserved	non-reserved
	"TRANSLATE":                        false, //  	reserved	reserved	reserved
	"TRANSLATE_REGEX":                  false, //  	reserved	reserved
	"TRANSLATION":                      false, //  	reserved	reserved	reserved
	"TREAT":                            false, // non-reserved (cannot be function or type)	reserved	reserved
	"TRIGGER":                          false, // non-reserved	reserved	reserved
	"TRIGGER_CATALOG":                  false, //  	non-reserved	non-reserved
	"TRIGGER_NAME":                     false, //  	non-reserved	non-reserved
	"TRIGGER_SCHEMA":                   false, //  	non-reserved	non-reserved
	"TRIM":                             false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"TRIM_ARRAY":                       false, //  	reserved	reserved
	"TRUE":                             true,  // reserved	reserved	reserved	reserved
	"TRUNCATE":                         false, // non-reserved	reserved	reserved
	"TRUSTED":                          false, // non-reserved
	"TYPE":                             false, // non-reserved	non-reserved	non-reserved	non-reserved
	"TYPES":                            false, // non-reserved
	"UESCAPE":                          false, // non-reserved	reserved	reserved
	"UNBOUNDED":                        false, // non-reserved	non-reserved	non-reserved
	"UNCOMMITTED":                      false, // non-reserved	non-reserved	non-reserved	non-reserved
	"UNCONDITIONAL":                    false, // non-reserved	non-reserved	non-reserved
	"UNDER":                            false, //  	non-reserved	non-reserved
	"UNENCRYPTED":                      false, // non-reserved
	"UNION":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"UNIQUE":                           true,  // reserved	reserved	reserved	reserved
	"UNKNOWN":                          false, // non-reserved	reserved	reserved	reserved
	"UNLINK":                           false, //  	non-reserved	non-reserved
	"UNLISTEN":                         false, // non-reserved
	"UNLOGGED":                         false, // non-reserved
	"UNMATCHED":                        false, //  	non-reserved	non-reserved
	"UNNAMED":                          false, //  	non-reserved	non-reserved	non-reserved
	"UNNEST":                           false, //  	reserved	reserved
	"UNTIL":                            false, // non-reserved
	"UNTYPED":                          false, //  	non-reserved	non-reserved
	"UPDATE":                           false, // non-reserved	reserved	reserved	reserved
	"UPPER":                            false, //  	reserved	reserved	reserved
	"URI":                              false, //  	non-reserved	non-reserved
	"USAGE":                            false, //  	non-reserved	non-reserved	reserved
	"USER":                             true,  // reserved	reserved	reserved	reserved
	"USER_DEFINED_TYPE_CATALOG":        false, //  	non-reserved	non-reserved
	"USER_DEFINED_TYPE_CODE":           false, //  	non-reserved	non-reserved
	"USER_DEFINED_TYPE_NAME":           false, //  	non-reserved	non-reserved
	"USER_DEFINED_TYPE_SCHEMA":         false, //  	non-reserved	non-reserved
	"USING":                            true,  // reserved	reserved	reserved	reserved
	"UTF16":                            false, //  	non-reserved	non-reserved
	"UTF32":                            false, //  	non-reserved	non-reserved
	"UTF8":                             false, //  	non-reserved	non-reserved
	"VACUUM":                           false, // non-reserved
	"VALID":                            false, // non-reserved	non-reserved	non-reserved
	"VALIDATE":                         false, // non-reserved
	"VALIDATOR":                        false, // non-reserved
	"VALUE":                            false, // non-reserved	reserved	reserved	reserved
	"VALUES":                           false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"VALUE_OF":                         false, //  	reserved	reserved
	"VARBINARY":                        false, //  	reserved	reserved
	"VARCHAR":                          false, // non-reserved (cannot be function or type)	reserved	reserved	reserved
	"VARIADIC":                         true,  // reserved
	"VARYING":                          false, // non-reserved, requires AS	reserved	reserved	reserved
	"VAR_POP":                          false, //  	reserved	reserved
	"VAR_SAMP":                         false, //  	reserved	reserved
	"VERBOSE":                          true,  // reserved (can be function or type)
	"VERSION":                          false, // non-reserved	non-reserved	non-reserved
	"VERSIONING":                       false, //  	reserved	reserved
	"VIEW":                             false, // non-reserved	non-reserved	non-reserved	reserved
	"VIEWS":                            false, // non-reserved
	"VOLATILE":                         false, // non-reserved
	"WHEN":                             true,  // reserved	reserved	reserved	reserved
	"WHENEVER":                         false, //  	reserved	reserved	reserved
	"WHERE":                            true,  // reserved, requires AS	reserved	reserved	reserved
	"WHITESPACE":                       false, // non-reserved	non-reserved	non-reserved
	"WIDTH_BUCKET":                     false, //  	reserved	reserved
	"WINDOW":                           true,  // reserved, requires AS	reserved	reserved
	"WITH":                             true,  // reserved, requires AS	reserved	reserved	reserved
	"WITHIN":                           false, // non-reserved, requires AS	reserved	reserved
	"WITHOUT":                          false, // non-reserved, requires AS	reserved	reserved
	"WORK":                             false, // non-reserved	non-reserved	non-reserved	reserved
	"WRAPPER":                          false, // non-reserved	non-reserved	non-reserved
	"WRITE":                            false, // non-reserved	non-reserved	non-reserved	reserved
	"XML":                              false, // non-reserved	reserved	reserved
	"XMLAGG":                           false, //  	reserved	reserved
	"XMLATTRIBUTES":                    false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLBINARY":                        false, //  	reserved	reserved
	"XMLCAST":                          false, //  	reserved	reserved
	"XMLCOMMENT":                       false, //  	reserved	reserved
	"XMLCONCAT":                        false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLDECLARATION":                   false, //  	non-reserved	non-reserved
	"XMLDOCUMENT":                      false, //  	reserved	reserved
	"XMLELEMENT":                       false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLEXISTS":                        false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLFOREST":                        false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLITERATE":                       false, //  	reserved	reserved
	"XMLNAMESPACES":                    false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLPARSE":                         false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLPI":                            false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLQUERY":                         false, //  	reserved	reserved
	"XMLROOT":                          false, // non-reserved (cannot be function or type)
	"XMLSCHEMA":                        false, //  	non-reserved	non-reserved
	"XMLSERIALIZE":                     false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLTABLE":                         false, // non-reserved (cannot be function or type)	reserved	reserved
	"XMLTEXT":                          false, //  	reserved	reserved
	"XMLVALIDATE":                      false, //  	reserved	reserved
	"YEAR":                             false, // non-reserved, requires AS	reserved	reserved	reserved
	"YES":                              false, // non-reserved	non-reserved	non-reserved
	"ZONE":                             false, // non-reserved	non-reserved	non-reserved	reserved
}
