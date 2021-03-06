package sqx

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/bingoohuang/gg/pkg/mapstruct"
)

// QueryAsNumber executes a query which only returns number like count(*) sql.
func (s SQL) QueryAsNumber(db *sql.DB) (int64, error) {
	str, err := s.QueryAsString(db)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(str, 10, 64)
}

// QueryAsString executes a query which only returns number like count(*) sql.
func (s SQL) QueryAsString(db *sql.DB) (string, error) {
	row, err := s.QueryAsRow(db)
	if err != nil {
		return "", err
	}

	if len(row) == 0 {
		return "", nil
	}

	return row[0], nil
}

// Update executes an update/delete query and returns rows affected.
func (s SQL) Update(db *sql.DB) (int64, error) {
	if s.Log {
		log.Printf("I! execute [%s] with [%v]", s.Query, s.Vars)
	}

	var r sql.Result
	var err error
	if s.Ctx != nil {
		r, err = db.ExecContext(s.Ctx, s.Query, s.Vars...)
	} else {
		r, err = db.Exec(s.Query, s.Vars...)
	}
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

type RowScannerInit interface {
	InitRowScanner(columns []string)
}

type RowScanner interface {
	ScanRow(rows *sql.Rows, rowIndex int) (bool, error)
}

type ScanRowFn func(rows *sql.Rows, rowIndex int) (bool, error)

func (s ScanRowFn) ScanRow(rows *sql.Rows, rowIndex int) (bool, error) {
	return s(rows, rowIndex)
}

// QueryOption defines the query options.
type QueryOption struct {
	MaxRows          int
	TagNames         []string
	Scanner          RowScanner
	LowerColumnNames bool
}

// QueryOptionFn define the prototype function to set QueryOption.
type QueryOptionFn func(o *QueryOption)

// QueryOptionFns is the slice of QueryOptionFn.
type QueryOptionFns []QueryOptionFn

func (q QueryOptionFns) Options() *QueryOption {
	o := &QueryOption{
		TagNames: []string{"col", "db", "mapstruct", "field", "json", "yaml"},
	}

	for _, fn := range q {
		fn(o)
	}

	return o
}

// WithMaxRows set the max rows of QueryOption.
func WithMaxRows(maxRows int) QueryOptionFn {
	return func(o *QueryOption) { o.MaxRows = maxRows }
}

// WithLowerColumnNames set the LowerColumnNames of QueryOption.
func WithLowerColumnNames(v bool) QueryOptionFn {
	return func(o *QueryOption) { o.LowerColumnNames = v }
}

// WithTagNames set the tagNames for mapping struct fields to query Columns.
func WithTagNames(tagNames ...string) QueryOptionFn {
	return func(o *QueryOption) { o.TagNames = tagNames }
}

// WithOptions apply the query option directly.
func WithOptions(v *QueryOption) QueryOptionFn {
	return func(o *QueryOption) { *o = *v }
}

// WithScanRow set row scanner for the query result.
func WithScanRow(v ScanRowFn) QueryOptionFn {
	return func(o *QueryOption) { o.Scanner = v }
}

// WithRowScanner set row scanner for the query result.
func WithRowScanner(v RowScanner) QueryOptionFn {
	return func(o *QueryOption) { o.Scanner = v }
}

// allowRowNum test the current rowNum is allowed for MaxRows control.
func (o QueryOption) allowRowNum(rowNum int) bool {
	return o.MaxRows == 0 || rowNum <= o.MaxRows
}

// QueryAsBeans query return with result.
func (s SQL) QueryAsBeans(db *sql.DB, result interface{}, optionFns ...QueryOptionFn) error {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	option := QueryOptionFns(optionFns).Options()
	decoder, err := mapstruct.NewDecoder(&mapstruct.Config{
		Result:           result,
		TagNames:         option.TagNames,
		Squash:           true,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}

	var input interface{}

	if resultValue.Elem().Kind() == reflect.Struct {
		input, err = s.QueryAsMap(db, WithOptions(option))
	} else {
		input, err = s.QueryAsMaps(db, WithOptions(option))
	}

	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

type MapScanner struct {
	Data    []map[string]string
	Columns []string
	MaxRows int
}

func (r *MapScanner) Data0() map[string]string {
	if len(r.Data) == 0 {
		return nil
	}

	return r.Data[0]
}

func (m *MapScanner) InitRowScanner(columns []string) {
	m.Columns = append(m.Columns, columns...)
}

func (m *MapScanner) ScanRow(rows *sql.Rows, _ int) (bool, error) {
	if v, err := ScanMapRow(rows, m.Columns); err != nil {
		return false, err
	} else {
		m.Data = append(m.Data, v)
		return m.MaxRows == 0 || len(m.Data) < m.MaxRows, nil
	}
}

// QueryAsMaps query rows as map slice.
func (s SQL) QueryAsMaps(db *sql.DB, optionFns ...QueryOptionFn) ([]map[string]string, error) {
	scanner := &MapScanner{Data: make([]map[string]string, 0)}
	err := s.QueryRaw(db, append(optionFns, WithRowScanner(scanner))...)
	return scanner.Data, err
}

// QueryAsMap query a single row as a map return.
func (s SQL) QueryAsMap(db *sql.DB, optionFns ...QueryOptionFn) (map[string]string, error) {
	scanner := &MapScanner{Data: make([]map[string]string, 0), MaxRows: 1}
	err := s.QueryRaw(db, append(optionFns, WithRowScanner(scanner))...)
	return scanner.Data0(), err
}

func ScanMapRow(rows *sql.Rows, columns []string) (map[string]string, error) {
	holders, err := ScanRow(len(columns), rows)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for i, h := range holders {
		m[columns[i]] = h.String
	}

	return m, nil
}

type StringRowScanner struct {
	Data    [][]string
	Columns []string
	MaxRows int
}

func (r *StringRowScanner) InitRowScanner(columns []string) {
	r.Columns = append(r.Columns, columns...)
}

func (r *StringRowScanner) ScanRow(rows *sql.Rows, _ int) (bool, error) {
	if m, err := ScanStringRow(rows, r.Columns); err != nil {
		return false, err
	} else {
		r.Data = append(r.Data, m)
		return r.MaxRows == 0 || len(r.Data) < r.MaxRows, nil
	}
}

func (r *StringRowScanner) Data0() []string {
	if len(r.Data) == 0 {
		return nil
	}

	return r.Data[0]
}

// QueryAsRow query a single row as a string slice return.
func (s SQL) QueryAsRow(db *sql.DB, optionFns ...QueryOptionFn) ([]string, error) {
	f := &StringRowScanner{MaxRows: 1}
	if err := s.QueryRaw(db, append(optionFns, WithRowScanner(f))...); err != nil {
		return nil, err
	}

	return f.Data0(), nil
}

// QueryAsRows query rows as [][]string.
func (s SQL) QueryAsRows(db *sql.DB, optionFns ...QueryOptionFn) ([][]string, error) {
	f := &StringRowScanner{}
	if err := s.QueryRaw(db, append(optionFns, WithRowScanner(f))...); err != nil {
		return nil, err
	}

	return f.Data, nil
}

func ScanStringRow(rows *sql.Rows, columns []string) ([]string, error) {
	holders, err := ScanRow(len(columns), rows)
	if err != nil {
		return nil, err
	}

	m := make([]string, len(columns))
	for i, h := range holders {
		m[i] = h.String
	}
	return m, nil
}

// QueryRaw query rows for customized row scanner.
func (s SQL) QueryRaw(db *sql.DB, optionFns ...QueryOptionFn) error {
	option, r, columns, err := s.prepareQuery(db, optionFns...)
	if err != nil {
		return err
	}

	defer r.Close()

	if initial, ok := option.Scanner.(RowScannerInit); ok {
		initial.InitRowScanner(columns)
	}

	for rn := 0; r.Next() && option.allowRowNum(rn+1); rn++ {
		if continued, err := option.Scanner.ScanRow(r, rn); err != nil {
			return err
		} else if !continued {
			break
		}
	}

	return nil
}

func ScanRow(columnSize int, r *sql.Rows) ([]sql.NullString, error) {
	holders := make([]sql.NullString, columnSize)
	pointers := make([]interface{}, columnSize)
	for i := 0; i < columnSize; i++ {
		pointers[i] = &holders[i]
	}

	if err := r.Scan(pointers...); err != nil {
		return nil, err
	}

	return holders, nil
}

func (s SQL) prepareQuery(db *sql.DB, optionFns ...QueryOptionFn) (*QueryOption, *sql.Rows, []string, error) {
	option := QueryOptionFns(optionFns).Options()

	if s.Log {
		log.Printf("I! execute [%s] with [%v]", s.Query, s.Vars)
	}
	var r *sql.Rows
	var err error
	if s.Ctx != nil {
		r, err = db.QueryContext(s.Ctx, s.Query, s.Vars...)
	} else {
		r, err = db.Query(s.Query, s.Vars...)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, nil, nil, err
	}

	if option.LowerColumnNames {
		for i, col := range columns {
			columns[i] = strings.ToLower(col)
		}
	}

	return option, r, columns, nil
}
