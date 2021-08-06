package sqlparser

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
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

// CreatePagingClause SQL statement for wrapping paging.
func (t DBType) CreatePagingClause(p *Paging, placeholder bool) (page string, bindArgs []interface{}) {
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
			b.WriteByte('`')
		}
	}
	b.WriteByte('`')
	return b.String()
}

type DoubleQuoteIdQuoter struct{}

func (DoubleQuoteIdQuoter) Quote(s string) string { return strconv.Quote(s) }

type PlaceholderFormatter interface {
	FormatPlaceholder() string
}

type QuestionPlaceholderFormatter struct{}

func (QuestionPlaceholderFormatter) FormatPlaceholder() string { return "?" }

type PrefixPlaceholderFormatter struct {
	Prefix string
	Pos    int // 1-based
}

func (p PrefixPlaceholderFormatter) FormatPlaceholder() string {
	return fmt.Sprintf("%s%d", p.Prefix, p.Pos)
}

type ConvertConfig struct {
	Paging          *Paging
	AutoIncrementPK string
}

type ConvertOption func(*ConvertConfig)

func WithPaging(v *Paging) ConvertOption { return func(c *ConvertConfig) { c.Paging = v } }
func WithAutoIncrementPK(v string) ConvertOption {
	return func(c *ConvertConfig) { c.AutoIncrementPK = v }
}

type ConvertResult struct {
	ExtraArgs     []interface{}
	CountingQuery string
	ScanValues    []interface{}
}

var errFound = errors.New("found")

const CreateCountingQuery = -1

// Convert converts query to target db type.
// 1. adjust the SQL variable symbols by different type, such as ?,? $1,$2.
// 1. quote table name, field names.
func (t DBType) Convert(query string, options ...ConvertOption) (string, *ConvertResult, error) {
	config := &ConvertConfig{}
	for _, f := range options {
		f(config)
	}

	stmt, err := Parse(query)
	if err != nil {
		return "", nil, err
	}

	hasPlaceholder := false
	_ = stmt.WalkSubtree(func(node SQLNode) (kontinue bool, err error) {
		if v, ok := (node).(*SQLVal); ok && v.Type == ValArg {
			hasPlaceholder = true
			return false, errFound
		}
		return true, nil
	})

	cr := &ConvertResult{}
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

	p := config.Paging
	if p != nil {
		pagingClause, bindArgs := t.CreatePagingClause(p, hasPlaceholder)
		cr.ExtraArgs = append(cr.ExtraArgs, bindArgs...)
		if p.RowsCount == CreateCountingQuery {
			cr.CountingQuery = t.createCountingQuery(stmt, buf, q)
		}
		q += " " + pagingClause
	}

	if config.AutoIncrementPK != "" {
		q += " " + t.createAutoIncrementPK(cr, config.AutoIncrementPK)
	}

	return q, cr, nil
}

func (t DBType) createCountingQuery(stmt Statement, buf *TrackedBuffer, q string) string {
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
		query := "select count(*) cnt from (" + q + ") t"
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

	buf.Myprintf("%v", stmt)
	return buf.String()
}

func (t DBType) createAutoIncrementPK(cr *ConvertResult, pk string) string {
	switch t {
	case Postgresql, Kingbase:
		// https://gist.github.com/miguelmota/d54814683346c4c98cec432cf99506c0
		return "returning " + pk
	case Oracle, Shentong:
		// https://forum.golangbridge.org/t/returning-values-with-insert-query-using-oracle-database-in-golang/13099/5
		var p int64 = 0
		cr.ScanValues = append(cr.ScanValues, sql.Named("ggReturningID", sql.Out{Dest: &p}))
		return "returning " + pk + " into :ggReturningID "
	default:
		return ""
	}
}
