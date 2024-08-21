package sqlparser

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/bits"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/bingoohuang/gg/pkg/reflector"
	"github.com/bingoohuang/gg/pkg/ss"
)

type DBType string

const (
	Mysql      DBType = "mysql"
	Sqlite3    DBType = "sqlite3"
	Dm         DBType = "dm"    // dm数据库
	Gbase      DBType = "gbase" // 南大通用
	Clickhouse DBType = "clickhouse"
	Postgresql DBType = "postgresql"
	Kingbase   DBType = "kingbase" // 金仓
	Shentong   DBType = "shentong" // 神通
	Mssql      DBType = "mssql"    // sqlserver 2012+
	Oracle     DBType = "oracle"   // oracle 12c+
)

// ToDBType converts driverName to different DBType.
func ToDBType(driverName string) DBType {
	switch strings.ToLower(driverName) {
	case "pgx", "opengauss":
		return Postgresql
	default:
		return DBType(driverName)
	}
}

var ErrUnsupportedDBType = errors.New("unsupported database type")

// Paging is pagination object.
type Paging struct {
	PageSeq   int  // Current page number, starting from 1
	PageSize  int  // How many items per page, 20 items by default
	RowsCount int  // Total number of data
	PageCount int  // How many pages
	FirstPage bool // Is it the first page
	HasPrev   bool // Whether there is a previous page
	HasNext   bool // Is there a next page
	LastPage  bool // Is it the last page
}

// NewPaging creates a Paging object.
func NewPaging() *Paging { return &Paging{PageSeq: 1, PageSize: 20} }

// SetRowsCount Set the total number of rows, calculate other values.
func (p *Paging) SetRowsCount(total int) {
	p.RowsCount = total
	p.PageCount = (p.RowsCount + p.PageSize - 1) / p.PageSize
	if p.PageSeq >= p.PageCount {
		p.LastPage = true
	} else {
		p.HasNext = true
	}
	if p.PageSeq > 1 {
		p.HasPrev = true
	} else {
		p.FirstPage = true
	}
}

// CompatibleLimit represents a LIMIT clause.
type CompatibleLimit struct {
	*Limit
	SwapArgs func(args []interface{})
	DBType
}

// Format formats the node.
func (n *CompatibleLimit) Format(buf *TrackedBuffer) {
	if n == nil {
		return
	}
	switch n.DBType {
	case Mysql, Sqlite3, Dm, Gbase, Clickhouse:
		buf.Myprintf(" limit ")
		if n.Offset != nil {
			buf.Myprintf("%v, ", n.Offset)
		}
		buf.Myprintf("%v", n.Rowcount)
		if n.Offset != nil && n.Rowcount != nil {
			offsetVar, ok1 := n.Offset.(*SQLVal)
			rowcount, ok2 := n.Rowcount.(*SQLVal)
			if ok1 && ok2 && offsetVar.Seq > rowcount.Seq {
				i := offsetVar.Seq - 1
				j := rowcount.Seq - 1
				n.SwapArgs = func(args []interface{}) {
					args[i], args[j] = args[j], args[i]
				}
			}
		}
	case Postgresql, Kingbase, Shentong:
		// https://www.postgresql.org/docs/9.3/queries-limit.html
		// SELECT select_list
		//    FROM table_expression
		//    [ ORDER BY ... ]
		//    [ LIMIT { number | ALL } ] [ OFFSET number ]
		buf.Myprintf(" limit %v", n.Rowcount)
		if n.Offset != nil {
			buf.Myprintf("offset %v", n.Offset)
		}
	case Mssql, Oracle:
		if n.Offset != nil {
			buf.Myprintf(" offset %v rows", n.Offset)
		}
		buf.Myprintf(" fetch next %v rows only", n.Rowcount)
		if n.Offset != nil && n.Rowcount != nil {
			offsetVar, ok1 := n.Offset.(*SQLVal)
			rowcount, ok2 := n.Rowcount.(*SQLVal)
			if ok1 && ok2 && offsetVar.Seq > rowcount.Seq {
				i := offsetVar.Seq - 1
				j := rowcount.Seq - 1
				n.SwapArgs = func(args []interface{}) {
					args[i], args[j] = args[j], args[i]
				}
			}
		}
	default:
		panic(ErrUnsupportedDBType)
	}
}

func (t DBType) createPagingClause(plFormatter PlaceholderFormatter, p *Paging, placeholder bool) (page string, bindArgs []interface{}) {
	var s strings.Builder
	start := p.PageSize * (p.PageSeq - 1)
	plf := plFormatter.FormatPlaceholder
	switch t {
	case Mysql, Sqlite3, Dm, Gbase, Clickhouse:
		if placeholder {
			s.WriteString(fmt.Sprintf("limit %s,%s", plf(), plf()))
			bindArgs = []interface{}{start, p.PageSize}
		} else {
			s.WriteString(fmt.Sprintf("limit %d,%d", start, p.PageSize))
		}
	case Postgresql, Kingbase, Shentong:
		if placeholder {
			s.WriteString(fmt.Sprintf("limit %s offset %s", plf(), plf()))
			bindArgs = []interface{}{p.PageSize, start}
		} else {
			s.WriteString(fmt.Sprintf("limit %d offset %d", p.PageSize, start))
		}
	case Mssql, Oracle:
		if placeholder {
			s.WriteString(fmt.Sprintf("offset %s rows fetch next %s rows only", plf(), plf()))
			bindArgs = []interface{}{start, p.PageSize}
		} else {
			s.WriteString(fmt.Sprintf("offset %d rows fetch next %d rows only", start, p.PageSize))
		}
	default:
		panic(ErrUnsupportedDBType)
	}

	page = s.String()
	return
}

type IdQuoter interface {
	Quote(string) string
}

const (
	quote  = '\''
	escape = '\\'
)

// SingleQuote returns a single-quoted Go string literal representing s. But, nothing else escapes.
func SingleQuote(s string) string {
	out := []rune{quote}
	for _, r := range s {
		switch r {
		case quote:
			out = append(out, escape, r)
		default:
			out = append(out, r)
		}
	}
	out = append(out, quote)
	return string(out)
}

type KeywordQuoter struct {
	DBType
	Upper bool
}

func (kq KeywordQuoter) Quote(s string) string {
	namesMap := quoteNames[kq.DBType]
	if len(namesMap) > 0 && namesMap[strings.ToUpper(s)] {
		return strconv.Quote(kq.upper(s))
	}

	return s
}

func (kq KeywordQuoter) upper(s string) string {
	if kq.Upper {
		return strings.ToUpper(s)
	}

	return s
}

type MySQLIdQuoter struct{}

func (kq MySQLIdQuoter) Quote(s string) string {
	b := new(bytes.Buffer)
	b.WriteByte('`')
	for _, c := range s {
		b.WriteRune(c)
		if c == '`' {
			b.WriteRune('`')
		}
	}
	b.WriteByte('`')
	return b.String()
}

type DoubleQuoteIdQuoter struct{}

func (kq DoubleQuoteIdQuoter) Quote(s string) string {
	return strconv.Quote(s)
}

type PlaceholderFormatter interface {
	FormatPlaceholder() string
	ResetPlaceholder()
}

type QuestionPlaceholderFormatter struct{}

func (QuestionPlaceholderFormatter) FormatPlaceholder() string { return "?" }
func (QuestionPlaceholderFormatter) ResetPlaceholder()         {}

type PrefixPlaceholderFormatter struct {
	Prefix string
	Pos    int // 1-based
}

func (p *PrefixPlaceholderFormatter) ResetPlaceholder() { p.Pos = 0 }
func (p *PrefixPlaceholderFormatter) FormatPlaceholder() string {
	p.Pos++
	return fmt.Sprintf("%s%d", p.Prefix, p.Pos)
}

type ConvertConfig struct {
	Paging             *Paging
	AutoIncrementField string
}

type ConvertOption func(*ConvertConfig)

func WithLimit(v int) ConvertOption {
	return func(c *ConvertConfig) { c.Paging = &Paging{PageSeq: 1, PageSize: v} }
}
func WithPaging(v *Paging) ConvertOption { return func(c *ConvertConfig) { c.Paging = v } }
func WithAutoIncrement(v string) ConvertOption {
	return func(c *ConvertConfig) { c.AutoIncrementField = v }
}

type ConvertResult struct {
	ExtraArgs     []interface{}
	CountingQuery string
	ScanValues    []interface{}
	VarPoses      []int // []var pos)
	BindMode      BindMode
	VarNames      []string

	InPlaceholder *InPlaceholder
	Placeholders  int

	ConvertQuery func() string
	SwapArgs     func(args []interface{})
}

type BindMode uint

const (
	ByPlaceholder BindMode = 1 << iota
	BySeq
	ByName
)

func (r *ConvertResult) PickArgs(args []interface{}) (q string, bindArgs []interface{}) {
	switch r.BindMode {
	case ByName:
		arg := args[0]
		if IsStructOrPtrToStruct(arg) {
			obj := reflector.New(arg)
			for _, name := range r.VarNames {
				name2 := strings.ToLower(ss.Strip(name, func(r rune) bool { return r == '-' || r == '_' }))
				v, err := obj.Field(name2).Get()
				if err != nil {
					v, err = obj.FieldByTag("db", name).Get()
				}

				if err != nil {
					panic(err)
				}

				bindArgs = append(bindArgs, v)
			}

		} else if IsMap(arg) {
			f := func(s string) string {
				return strings.ToLower(ss.Strip(s, func(r rune) bool { return r == '-' || r == '_' }))
			}
			vmap := reflect.ValueOf(arg)
			for _, name := range r.VarNames {
				if v := vmap.MapIndex(reflect.ValueOf(name)); v.IsValid() {
					bindArgs = append(bindArgs, v.Interface())
				} else {
					bindArg, _ := findInMap(vmap, name, f)
					bindArgs = append(bindArgs, bindArg)
				}
			}
		} else {
			bindArgs = args
		}

	case BySeq:
		for _, p := range r.VarPoses {
			bindArgs = append(bindArgs, args[p-1])
		}
	default:
		if r.IsInPlaceholders() {
			if len(args) == 1 && IsSlice(args[0]) {
				r.ResetInVars(SliceLen(args[0]))
				bindArgs = CreateSlice(args[0])
			} else {
				r.ResetInVars(len(args))
				bindArgs = args
			}
		} else {
			bindArgs = append(bindArgs, args...)
		}
	}

	resultArgs := append(bindArgs, r.ExtraArgs...)
	query := r.ConvertQuery()
	if r.SwapArgs != nil {
		r.SwapArgs(resultArgs)
	}
	return query, resultArgs
}

func CreateSlice(i interface{}) []interface{} {
	ti := reflect.ValueOf(i)
	elements := make([]interface{}, ti.Len())

	for i := 0; i < ti.Len(); i++ {
		elements[i] = ti.Index(i).Interface()
	}

	return elements
}

func SliceLen(i interface{}) int {
	ti := reflect.ValueOf(i)
	return ti.Len()
}

func IsSlice(i interface{}) bool {
	return reflect.TypeOf(i).Kind() == reflect.Slice
}

func (r ConvertResult) IsInPlaceholders() bool {
	return r.InPlaceholder != nil && r.Placeholders == r.InPlaceholder.Num
}

func (r ConvertResult) ResetInVars(varsNum int) {
	if varsNum == r.InPlaceholder.Num {
		return
	}

	var exprs ValTuple

	for i := 0; i < varsNum; i++ {
		exprs = append(exprs, &SQLVal{Type: ValArg, Val: []byte("?")})
	}

	if varsNum == 0 {
		exprs = append(exprs, &NullVal{})
	}

	r.InPlaceholder.Expr.Right = exprs
}

func findInMap(vmap reflect.Value, name string, f func(s string) string) (interface{}, bool) {
	name = f(name)

	for iter := vmap.MapRange(); iter.Next(); {
		k := iter.Key().Interface()
		if kk, ok := k.(string); ok {
			if f(kk) == name {
				return iter.Value().Interface(), true
			}
		}
	}

	return nil, false
}

func IsMap(arg interface{}) bool {
	t := reflect.TypeOf(arg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Map
}

func IsStructOrPtrToStruct(arg interface{}) bool {
	t := reflect.TypeOf(arg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

var ErrSyntax = errors.New("syntax not supported")

const CreateCountingQuery = -1

var numReg = regexp.MustCompile(`^[1-9]\d*$`)

// 在Oracle数据库中，使用双引号（"）来引用列名通常有以下几个特定含义：
// 区分大小写：Oracle默认情况下列名是大小写不敏感的，但是当你使用双引号将列名括起来时，它将变得大小写敏感。这意味着，如果列名在数据库中是大写，你在使用时也需要使用大写。
// 保留字：如果列名是SQL保留字，例如 SELECT、FROM 等，使用双引号可以避免语法错误，因为双引号告诉Oracle这个保留字是列名而不是SQL关键字。
// 特殊字符：如果列名中包含特殊字符或空格，使用双引号可以正确引用这些列名。
// 跨数据库兼容性：在某些情况下，使用双引号引用列名可以提高SQL语句在不同数据库系统间的兼容性，因为有些数据库系统默认使用双引号来区分标识符。
// 例如，以下是一个使用双引号的Oracle查询语句示例：
// SELECT "ColumnName" FROM "TableName";
// 在这个例子中，如果ColumnName是一个保留字或者包含特殊字符，使用双引号可以确保Oracle正确解析它为列名。同时，如果列名的大小写与数据库中实际的大小写不一致，使用双引号也可以确保正确引用。

var quoteNames = map[DBType]map[string]bool{
	Oracle: toMap([]string{
		// https://docs.oracle.com/cd/B10501_01/appdev.920/a42525/apb.htm

		"NUMBER", "BINARY_INTEGER", "BINARY_FLOAT", "BINARY_DOUBLE",
		"CHAR", "VARCHAR2", "CLOB", "DATE",
		"TIMESTAMP", "INTERVAL", "BLOB", "NCLOB",
		"ROWID", "UROWID", "BOOLEAN", "XMLTYPE",
		"RAW", "LONG", "BFILE", "VARRAY",

		// Oracle Reserved Words
		"ACCESS", "ELSE", "MODIFY", "START",
		"ADD", "EXCLUSIVE", "NOAUDIT", "SELECT",
		"ALL", "EXISTS", "NOCOMPRESS", "SESSION",
		"ALTER", "FILE", "NOT", "SET",
		"AND", "FLOAT", "NOTFOUND", "SHARE",
		"ANY", "FOR", "NOWAIT", "SIZE",
		"ARRAYLEN", "FROM", "NULL", "SMALLINT",
		"AS", "GRANT", "NUMBER", "SQLBUF",
		"ASC", "GROUP", "OF", "SUCCESSFUL",
		"AUDIT", "HAVING", "OFFLINE", "SYNONYM",
		"BETWEEN", "IDENTIFIED", "ON", "SYSDATE",
		"BY", "IMMEDIATE", "ONLINE", "TABLE",
		"CHAR", "IN", "OPTION", "THEN",
		"CHECK", "INCREMENT", "OR", "TO",
		"CLUSTER", "INDEX", "ORDER", "TRIGGER",
		"COLUMN", "INITIAL", "PCTFREE", "UID",
		"COMMENT", "INSERT", "PRIOR", "UNION",
		"COMPRESS", "INTEGER", "PRIVILEGES", "UNIQUE",
		"CONNECT", "INTERSECT", "PUBLIC", "UPDATE",
		"CREATE", "INTO", "RAW", "USER",
		"CURRENT", "IS", "RENAME", "VALIDATE",
		"DATE", "LEVEL", "RESOURCE", "VALUES",
		"DECIMAL", "LIKE", "REVOKE", "VARCHAR",
		"DEFAULT", "LOCK", "ROW", "VARCHAR2",
		"DELETE", "LONG", "ROWID", "VIEW",
		"DESC", "MAXEXTENTS", "ROWLABEL", "WHENEVER",
		"DISTINCT", "MINUS", "ROWNUM", "WHERE",
		"DROP", "MODE", "ROWS", "WITH",

		// Oracle Keywords
		"ADMIN", "CURSOR", "FOUND", "MOUNT",
		"AFTER", "CYCLE", "FUNCTION", "NEXT",
		"ALLOCATE", "DATABASE", "GO", "NEW",
		"ANALYZE", "DATAFILE", "GOTO", "NOARCHIVELOG",
		"ARCHIVE", "DBA", "GROUPS", "NOCACHE",
		"ARCHIVELOG", "DEC", "INCLUDING", "NOCYCLE",
		"AUTHORIZATION", "DECLARE", "INDICATOR", "NOMAXVALUE",
		"AVG", "DISABLE", "INITRANS", "NOMINVALUE",
		"BACKUP", "DISMOUNT", "INSTANCE", "NONE",
		"BEGIN", "DOUBLE", "INT", "NOORDER",
		"BECOME", "DUMP", "KEY", "NORESETLOGS",
		"BEFORE", "EACH", "LANGUAGE", "NORMAL",
		"BLOCK", "ENABLE", "LAYER", "NOSORT",
		"BODY", "END", "LINK", "NUMERIC",
		"CACHE", "ESCAPE", "LISTS", "OFF",
		"CANCEL", "EVENTS", "LOGFILE", "OLD",
		"CASCADE", "EXCEPT", "MANAGE", "ONLY",
		"CHANGE", "EXCEPTIONS", "MANUAL", "OPEN",
		"CHARACTER", "EXEC", "MAX", "OPTIMAL",
		"CHECKPOINT", "EXPLAIN", "MAXDATAFILES", "OWN",
		"CLOSE", "EXECUTE", "MAXINSTANCES", "PACKAGE",
		"COBOL", "EXTENT", "MAXLOGFILES", "PARALLEL",
		"COMMIT", "EXTERNALLY", "MAXLOGHISTORY", "PCTINCREASE",
		"COMPILE", "FETCH", "MAXLOGMEMBERS", "PCTUSED",
		"CONSTRAINT", "FLUSH", "MAXTRANS", "PLAN",
		"CONSTRAINTS", "FREELIST", "MAXVALUE", "PLI",
		"CONTENTS", "FREELISTS", "MIN", "PRECISION",
		"CONTINUE", "FORCE", "MINEXTENTS", "PRIMARY",
		"CONTROLFILE", "FOREIGN", "MINVALUE", "PRIVATE",
		"COUNT", "FORTRAN", "MODULE", "PROCEDURE",

		// Oracle Keywords (continued):
		"PROFILE", "SAVEPOINT", "SQLSTATE", "TRACING",
		"QUOTA", "SCHEMA", "STATEMENT_ID", "TRANSACTION",
		"READ", "SCN", "STATISTICS", "TRIGGERS",
		"REAL", "SECTION", "STOP", "TRUNCATE",
		"RECOVER", "SEGMENT", "STORAGE", "UNDER",
		"REFERENCES", "SEQUENCE", "SUM", "UNLIMITED",
		"REFERENCING", "SHARED", "SWITCH", "UNTIL",
		"RESETLOGS", "SNAPSHOT", "SYSTEM", "USE",
		"RESTRICTED", "SOME", "TABLES", "USING",
		"REUSE", "SORT", "TABLESPACE", "WHEN",
		"ROLE", "SQL", "TEMPORARY", "WRITE",
		"ROLES", "SQLCODE", "THREAD", "WORK",
		"ROLLBACK", "SQLERROR", "TIME",

		// PL/SQL Reserved Words
		"ABORT", "BETWEEN", "CRASH", "DIGITS",
		"ACCEPT", "BINARY_INTEGER", "CREATE", "DISPOSE",
		"ACCESS", "BODY", "CURRENT", "DISTINCT",
		"ADD", "BOOLEAN", "CURRVAL", "DO",
		"ALL", "BY", "CURSOR", "DROP",
		"ALTER", "CASE", "DATABASE", "ELSE",
		"AND", "CHAR", "DATA_BASE", "ELSIF",
		"ANY", "CHAR_BASE", "DATE", "END",
		"ARRAY", "CHECK", "DBA", "ENTRY",
		"ARRAYLEN", "CLOSE", "DEBUGOFF", "EXCEPTION",
		"AS", "CLUSTER", "DEBUGON", "EXCEPTION_INIT",
		"ASC", "CLUSTERS", "DECLARE", "EXISTS",
		"ASSERT", "COLUMNS", "DEFAULT", "FALSE",
		"ASSIGN", "AT", "COMMIT", "DEFINITION", "FETCH",
		"AUTHORIZATION", "COMPRESS", "DELAY", "FLOAT",
		"AVG", "CONNECT", "CONSTANT", "DESC", "FROM",

		"FUNCTION", "NEW", "RELEASE", "SUM",
		"GENERIC", "NEXTVAL", "REMR", "TABAUTH",
		"GOTO", "NOCOMPRESS", "RENAME", "TABLE",
		"GRANT", "NOT", "RESOURCE", "TABLES",
		"GROUP", "NULL", "RETURN", "TASK",
		"HAVING", "NUMBER", "REVERSE", "TERMINATE",
		"IDENTIFIED", "NUMBER_BASE", "REVOKE", "THEN",
		"IF", "OF", "ROLLBACK", "TO",
		"IN", "ON", "ROWID", "TRUE",
		"INDEX", "OPEN", "ROWLABEL", "TYPE",
		"INDEXES", "OPTION", "ROWNUM", "UNION",
		"INDICATOR", "OR", "ROWTYPE", "UNIQUE",
		"INSERT", "ORDER", "RUN", "UPDATE",
		"INTEGER", "OTHERS", "SAVEPOINT", "USE",
		"INTERSECT", "OUT", "SCHEMA", "VALUES",
		"INTO", "PACKAGE", "SELECT", "VARCHAR",
		"IS", "PARTITION", "SEPARATE", "VARCHAR2",
		"LEVEL", "PCTFREE", "SET", "VARIANCE",
		"LIKE", "POSITIVE", "SIZE", "VIEW",
		"LIMITED", "PRAGMA", "SMALLINT", "VIEWS",
		"LOOP", "PRIOR", "SPACE", "WHEN",
		"MAX", "PRIVATE", "SQL", "WHERE",
		"MIN", "PROCEDURE", "SQLCODE", "WHILE",
		"MINUS", "PUBLIC", "SQLERRM", "WITH",
		"MLSLABEL", "RAISE", "START", "WORK",
		"MOD", "RANGE", "STATEMENT", "XOR",
		"MODE", "REAL", "STDDEV", "NATURAL",
		"RECORD", "SUBTYPE",
	}),
}

func toMap(arr []string) map[string]bool {
	m := make(map[string]bool)
	for _, el := range arr {
		m[strings.ToUpper(el)] = true
	}

	return m
}

// RegisterQuoteNames 根据数据库类型 dbType，注册 需要添加双引号的字段名称
func RegisterQuoteNames(dbType DBType, names []string) {
	namesMap := quoteNames[dbType]
	if len(namesMap) == 0 {
		namesMap = make(map[string]bool)
	}

	for _, name := range names {
		namesMap[strings.ToUpper(name)] = true
	}

	quoteNames[dbType] = namesMap
}

// Convert converts query to target db type.
// 1. adjust the SQL variable symbols by different type, such as ?,? $1,$2.
// 1. quote table name, field names.
func (t DBType) Convert(query string, options ...ConvertOption) (*ConvertResult, error) {
	stmt, err := Parse(query)
	if err != nil {
		return nil, err
	}

	insertStmt, _ := stmt.(*Insert)
	if err := t.checkMySQLOnDuplicateKey(insertStmt); err != nil {
		return nil, fmt.Errorf("on duplicate key is not supported directly in SQL, error %w", ErrSyntax)
	}

	fixInsertPlaceholders(insertStmt)
	cr := &ConvertResult{}

	insertPos := -1
	lastColName := ""

	_ = stmt.WalkSubtree(func(node SQLNode) (kontinue bool, err error) {
		if cr.InPlaceholder == nil {
			cr.InPlaceholder = ParseInPlaceholder(node)
		}

		if cn, cnOk := node.(*ColName); cnOk {
			lastColName = cn.Name.Lowered()
			return true, nil
		}
		if _, ok := node.(*Limit); ok {
			return true, err
		}

		if v, ok := node.(*SQLVal); ok {
			switch v.Type {
			case ValArg, StrVal: // 转换 :a :b :c 或者 :1 :2 :3的占位符形式
				if string(v.Val) == "?" {
					cr.Placeholders++
				} else {
					convertCustomBinding(insertStmt, &insertPos, &lastColName, v, cr)
				}
			}
		}

		return true, nil
	})

	if len(cr.VarPoses) > 0 {
		cr.BindMode |= BySeq
	}
	if len(cr.VarNames) > 0 {
		cr.BindMode |= ByName
	}
	if cr.Placeholders > 0 {
		cr.BindMode |= ByPlaceholder
	}
	if bits.OnesCount(uint(cr.BindMode)) > 1 {
		return nil, fmt.Errorf("mixed bind modes are not supported, error %w", ErrSyntax)
	}

	buf := &TrackedBuffer{Buffer: new(bytes.Buffer)}

	switch t {
	case Postgresql, Kingbase:
		buf.PlaceholderFormatter = &PrefixPlaceholderFormatter{Prefix: "$"}
	case Mssql:
		buf.PlaceholderFormatter = &PrefixPlaceholderFormatter{Prefix: "@p"}
	case Oracle, Shentong:
		buf.PlaceholderFormatter = &PrefixPlaceholderFormatter{Prefix: ":"}
	default:
		buf.PlaceholderFormatter = &QuestionPlaceholderFormatter{}
	}

	switch t {
	case Mysql, Sqlite3, Gbase, Clickhouse:
		// https://www.sqlite.org/lang_keywords.html
		buf.IdQuoter = &MySQLIdQuoter{}
	case Oracle:
		buf.IdQuoter = &KeywordQuoter{DBType: Oracle, Upper: true}
	default:
		buf.IdQuoter = &DoubleQuoteIdQuoter{}
	}

	config := &ConvertConfig{}
	for _, f := range options {
		f(config)
	}

	selectStmt, _ := stmt.(*Select)
	var limit *Limit
	if selectStmt != nil {
		limit = selectStmt.Limit
		selectStmt.SetLimit(nil)
	}

	var compatibleLimit *CompatibleLimit
	p := config.Paging
	isPaging := selectStmt != nil && p != nil
	if !isPaging && limit != nil {
		compatibleLimit = &CompatibleLimit{Limit: limit, DBType: t}
		selectStmt.SetLimitSQLNode(compatibleLimit)
	}

	cr.ConvertQuery = func() string {
		buf.Myprintf("%v", stmt)
		q := buf.String()
		buf.Reset()

		if compatibleLimit != nil {
			cr.SwapArgs = compatibleLimit.SwapArgs
		}

		if isPaging {
			pagingClause, bindArgs := t.createPagingClause(buf.PlaceholderFormatter, p, cr.BindMode > 0)
			cr.ExtraArgs = append(cr.ExtraArgs, bindArgs...)
			if p.RowsCount == CreateCountingQuery {
				cr.CountingQuery = t.createCountingQuery(stmt, buf, q)
			}
			q += " " + pagingClause
		}

		if f := config.AutoIncrementField; f != "" {
			q += " " + t.createAutoIncrementPK(cr, f)
		}

		return q
	}

	return cr, nil
}

type InPlaceholder struct {
	Expr *ComparisonExpr
	Num  int
}

func ParseInPlaceholder(node SQLNode) *InPlaceholder {
	v, ok := node.(*ComparisonExpr)
	if !ok {
		return nil
	}

	if v.Operator != "in" {
		return nil
	}

	t, tOk := v.Right.(ValTuple)
	if !tOk {
		return nil
	}

	for _, tv := range t {
		if tw, twOK := tv.(*SQLVal); twOK {
			if !(tw.Type == ValArg && bytes.Equal(tw.Val, []byte("?"))) {
				return nil
			}
		} else {
			return nil
		}
	}

	return &InPlaceholder{Expr: v, Num: len(t)}
}

func convertCustomBinding(insert *Insert, insertPos *int, lastColName *string, v *SQLVal, cr *ConvertResult) {
	if len(v.Val) == 0 || !bytes.HasPrefix(v.Val, []byte(":")) {
		return
	}
	name := strings.TrimSpace(string(v.Val[1:]))
	if name == "" {
		return
	}

	if numReg.MatchString(name) {
		num, _ := strconv.Atoi(name)
		cr.VarPoses = append(cr.VarPoses, num)
	} else {
		if name == "?" { // 从上下文推断变量名称
			if insert != nil {
				*insertPos++
				col := insert.Columns[*insertPos]
				name = col.Lowered()
			} else if *lastColName != "" {
				name = *lastColName
				*lastColName = ""
			}
		}
		cr.VarNames = append(cr.VarNames, name)
	}

	v.Type = ValArg
	v.Val = []byte("?")
}

func (t DBType) checkMySQLOnDuplicateKey(insertStmt *Insert) error {
	// 只有MySQL 的 ON DUPLICATE KEY被支持
	// eg. INSERT INTO table (a,b,c) VALUES (1,2,3),(4,5,6) ON DUPLICATE KEY UPDATE c=VALUES(a)+VALUES(b);
	if insertStmt != nil && len(insertStmt.OnDup) > 0 {
		switch t {
		case Mysql:
		default:
			return ErrSyntax
		}
	}
	return nil
}

func fixInsertPlaceholders(insertStmt *Insert) {
	if insertStmt == nil {
		return
	}

	// 是insert into t(a,b,c) values(...)的格式
	insertRows, ok := insertStmt.Rows.(Values)
	if !ok {
		return
	}

	// 只有一个values列表
	if len(insertRows) != 1 {
		return
	}

	questionVals := 0
	inferVals := 0
	others := 0
	insertRow := insertRows[0]
	for _, node := range insertRow {
		if v, ok := (node).(*SQLVal); ok && v.Type == ValArg {
			switch vs := string(v.Val); vs {
			case "?":
				questionVals++
				continue
			case ":?":
				inferVals++
				continue
			}
		}

		others++
		break
	}

	// 不全是?占位符
	if others > 0 || questionVals > 0 && inferVals > 0 {
		return
	}

	diff := len(insertStmt.Columns) - ss.Ifi(questionVals > 0, questionVals, inferVals)
	if diff == 0 {
		return
	}

	if diff > 0 {
		pl := []byte(ss.If(questionVals > 0, "?", ":?"))
		appendVarArgs := make([]Expr, 0, diff)
		for i := 0; i < diff; i++ {
			appendVarArgs = append(appendVarArgs, NewValArg(pl))
		}
		insertRows[0] = append(insertRows[0], appendVarArgs...)
	} else {
		insertRows[0] = insertRows[0][:len(insertStmt.Columns)]
	}
}

func (t DBType) createCountingQuery(stmt Statement, buf *TrackedBuffer, q string) string {
	buf.PlaceholderFormatter.ResetPlaceholder()

	countWrapRequired := func() bool {
		if _, ok := stmt.(*Union); ok {
			return true
		}
		if v, ok := stmt.(*Select); ok && v.Distinct != "" || len(v.GroupBy) > 0 {
			v.OrderBy = nil
			return true
		}
		return false
	}

	if countWrapRequired() {
		query := "select count(*) cnt from (" + q + ") t_gg_cnt"
		cr, err := t.Convert(query)
		if err != nil {
			log.Printf("failed to convert query %s, err: %v", query, err)
			return ""
		}

		buf.Myprintf("%v", cr.ConvertQuery())
		return buf.String()
	}

	s, ok := stmt.(*Select)
	if !ok {
		return ""
	}

	s.OrderBy = nil
	p, _ := Parse(`select count(*) cnt`)
	s.SelectExprs = p.(*Select).SelectExprs
	buf.Myprintf("%v", s)
	return buf.String()
}

func (t DBType) createAutoIncrementPK(cr *ConvertResult, autoIncrementField string) string {
	switch t {
	case Postgresql, Kingbase:
		// https://gist.github.com/miguelmota/d54814683346c4c98cec432cf99506c0
		return "returning " + autoIncrementField
	case Oracle, Shentong:
		// https://forum.golangbridge.org/t/returning-values-with-insert-query-using-oracle-database-in-golang/13099/5
		var p int64 = 0
		cr.ScanValues = append(cr.ScanValues, sql.Named(autoIncrementField, sql.Out{Dest: &p}))
		return "returning " + autoIncrementField + " into :" + autoIncrementField
	default:
		return ""
	}
}
