package sqlparser

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type DBType string

const (
	Mysql      DBType = "mysql"
	Sqlite     DBType = "sqlite"
	Dm         DBType = "dm"    // dm数据库
	Gbase      DBType = "gbase" // 南大通用
	Clickhouse DBType = "clickhouse"
	Postgresql DBType = "postgresql"
	Kingbase   DBType = "kingbase" // 金仓
	Shentong   DBType = "shentong" // 神通
	Mssql      DBType = "mssql"    // sqlserver 2012+
	Oracle     DBType = "oracle"   // oracle 12c+
)

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

// createPagingClause SQL statement for wrapping paging.
func (t DBType) createPagingClause(p *Paging, placeholder bool) (page string, bindArgs []interface{}) {
	var s strings.Builder
	start := p.PageSize * (p.PageSeq - 1)
	switch t {
	case Mysql, Sqlite, Dm, Gbase, Clickhouse:
		if placeholder {
			s.WriteString("limit ?,?")
			bindArgs = []interface{}{start, p.PageSize}
		} else {
			s.WriteString(fmt.Sprintf("limit %d,%d", start, p.PageSize))
		}
	case Postgresql, Kingbase, Shentong:
		if placeholder {
			s.WriteString("limit ? offset ?")
			bindArgs = []interface{}{p.PageSize, start}
		} else {
			s.WriteString(fmt.Sprintf("limit %d offset %d", p.PageSize, start))
		}
	case Mssql, Oracle:
		if placeholder {
			s.WriteString("offset ? rows fetch next ? rows only")
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

type MySQLIdQuoter struct{}

func (MySQLIdQuoter) Quote(s string) string {
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

func (DoubleQuoteIdQuoter) Quote(s string) string { return strconv.Quote(s) }

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
	return fmt.Sprintf("%s%d", p.Prefix, p.Pos)
}

type ConvertConfig struct {
	Paging             *Paging
	AutoIncrementField string
}

type ConvertOption func(*ConvertConfig)

func WithPaging(v *Paging) ConvertOption { return func(c *ConvertConfig) { c.Paging = v } }
func WithAutoIncrement(v string) ConvertOption {
	return func(c *ConvertConfig) { c.AutoIncrementField = v }
}

type ConvertResult struct {
	ExtraArgs     []interface{}
	CountingQuery string
	ScanValues    []interface{}
	VarPosMap     map[string]int
	PosVarMap     map[int]string
	PosPosMap     map[int]int // real pos -> special pos
}

var errFound = errors.New("found")

var ErrSyntax = errors.New("syntax not supported")

const CreateCountingQuery = -1

var numReg = regexp.MustCompile(`^[1-9]\d*$`)

// Convert converts query to target db type.
// 1. adjust the SQL variable symbols by different type, such as ?,? $1,$2.
// 1. quote table name, field names.
func (t DBType) Convert(query string, options ...ConvertOption) (string, *ConvertResult, error) {
	stmt, err := Parse(query)
	if err != nil {
		return "", nil, err
	}

	insertStmt, _ := stmt.(*Insert)
	if err := t.checkMySQLOnDuplicateKey(insertStmt); err != nil {
		return "", nil, err
	}

	fixPlaceholders(insertStmt)
	cr := &ConvertResult{VarPosMap: make(map[string]int), PosVarMap: make(map[int]string), PosPosMap: make(map[int]int)}
	posIncr := 0
	purePlaceholders := 0

	_ = stmt.WalkSubtree(func(node SQLNode) (kontinue bool, err error) {
		if v, ok := (node).(*SQLVal); ok {
			switch v.Type {
			case ValArg, StrVal: // 转换 :a :b :c 或者 :1 :2 :3的占位符形式
				if string(v.Val) == "?" {
					purePlaceholders++
				} else {
					convertCustomBinding(v, &posIncr, cr)
				}
			}
		}
		return true, nil
	})

	modeUsed := 0
	if len(cr.PosPosMap) > 0 {
		modeUsed++
	}
	if len(cr.PosVarMap) > 0 {
		modeUsed++
	}
	if purePlaceholders > 0 {
		modeUsed++
	}
	if modeUsed > 1 {
		return "", nil, ErrSyntax
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
	case Mysql, Sqlite, Dm, Gbase, Clickhouse:
		buf.IdQuoter = &MySQLIdQuoter{}
	default:
		buf.IdQuoter = &DoubleQuoteIdQuoter{}
	}

	buf.Myprintf("%v", stmt)
	q := buf.String()
	buf.Reset()

	config := &ConvertConfig{}
	for _, f := range options {
		f(config)
	}

	if p := config.Paging; p != nil {
		hasPlaceholder := modeUsed > 0
		pagingClause, bindArgs := t.createPagingClause(p, hasPlaceholder)
		cr.ExtraArgs = append(cr.ExtraArgs, bindArgs...)
		if p.RowsCount == CreateCountingQuery {
			cr.CountingQuery = t.createCountingQuery(stmt, buf, q)
		}
		q += " " + pagingClause
	}

	if f := config.AutoIncrementField; f != "" {
		q += " " + t.createAutoIncrementPK(cr, f)
	}

	return q, cr, nil
}

func convertCustomBinding(v *SQLVal, posIncr *int, cr *ConvertResult) {
	if len(v.Val) == 0 || !bytes.HasPrefix(v.Val, []byte(":")) {
		return
	}
	name := strings.TrimSpace(string(v.Val[1:]))
	if name == "" {
		return
	}

	*posIncr++
	if numReg.MatchString(name) {
		num, _ := strconv.Atoi(name)
		cr.PosPosMap[*posIncr] = num
	} else {
		pos, exists := cr.VarPosMap[name]
		if exists {
			cr.PosVarMap[pos] = name
		} else {
			cr.VarPosMap[name] = *posIncr
		}
	}

	v.Type = ValArg
	v.Val = []byte("?")
}

func (t DBType) checkMySQLOnDuplicateKey(insertStmt *Insert) error {
	// MySQL 的 ON DUPLICATE KEY 不被支持
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

func fixPlaceholders(insertStmt *Insert) {
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
	others := 0
	insertRow := insertRows[0]
	for _, node := range insertRow {
		if v, ok := (node).(*SQLVal); ok && v.Type == ValArg && string(v.Val) == "?" {
			questionVals++
			continue
		} else {
			others++
			break
		}
	}

	// 不全是?占位符
	if others > 0 {
		return
	}

	diff := len(insertStmt.Columns) - questionVals
	if diff == 0 {
		return
	}

	if diff > 0 {
		var appendVarArgs = make([]Expr, 0, diff)
		for i := 0; i < diff; i++ {
			appendVarArgs = append(appendVarArgs, NewValArg([]byte("?")))
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
		countStmt, _, err := t.Convert(query)
		if err != nil {
			log.Printf("failed to convert query %s, err: %v", query, err)
			return ""
		}

		buf.Myprintf("%v", countStmt)
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
