package sqx

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"

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

// QueryAsNumber executes a query which only returns number like count(*) sql.
func (s SQL) QueryAsString(db *sql.DB) (string, error) {
	row, err := s.QueryAsRow(db)
	if err != nil {
		return "", err
	}

	return row[0], err
}

// Update executes an update/delete query and returns rows affected.
func (s SQL) Update(db *sql.DB) (int64, error) {
	log.Printf("I! execute [%s] with [%v]", s.Query, s.Vars)
	r, err := db.Exec(s.Query, s.Vars...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

type RowScanner interface {
	ScanRow(rows *sql.Rows, rowIndex int, columns []string) (bool, error)
}

type ScanRowFn func(rows *sql.Rows, rowIndex int, columns []string) (bool, error)

func (s ScanRowFn) ScanRow(rows *sql.Rows, rowIndex int, columns []string) (bool, error) {
	return s(rows, rowIndex, columns)
}

// QueryOption defines the query options.
type QueryOption struct {
	MaxRows  int
	TagNames []string
	Scanner  RowScanner
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

// WithTagNames set the tagNames for mapping struct fields to query columns.
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

// QueryAsMap query a single row as a map return.
func (s SQL) QueryAsMap(db *sql.DB, optionFns ...QueryOptionFn) (map[string]string, error) {
	var m map[string]string
	f := func(rows *sql.Rows, rowIndex int, columns []string) (bool, error) {
		if r, err := scanMapRow(rows, columns); err != nil {
			return false, err
		} else {
			m = r
			return false, nil
		}
	}

	err := s.QueryRaw(db, append(optionFns, WithScanRow(f))...)
	return m, err
}

type mapScanner struct {
	data []map[string]string
}

func (m *mapScanner) ScanRow(rows *sql.Rows, _ int, columns []string) (bool, error) {
	if v, err := scanMapRow(rows, columns); err != nil {
		return false, err
	} else {
		m.data = append(m.data, v)
		return true, nil
	}
}

// QueryAsMaps query rows as map slice.
func (s SQL) QueryAsMaps(db *sql.DB, optionFns ...QueryOptionFn) ([]map[string]string, error) {
	scanner := &mapScanner{data: make([]map[string]string, 0)}
	err := s.QueryRaw(db, append(optionFns, WithRowScanner(scanner))...)
	return scanner.data, err
}

func scanMapRow(rows *sql.Rows, columns []string) (map[string]string, error) {
	holders, err := scanRow(len(columns), rows)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for i, h := range holders {
		m[columns[i]] = h.String
	}

	return m, nil
}

// QueryAsRow query a single row as a string slice return.
func (s SQL) QueryAsRow(db *sql.DB, optionFns ...QueryOptionFn) ([]string, error) {
	var row []string
	f := func(rows *sql.Rows, rowIndex int, columns []string) (bool, error) {
		if m, err := scanStringRow(rows, columns); err != nil {
			return false, err
		} else {
			row = m
			return false, nil
		}
	}

	err := s.QueryRaw(db, append(optionFns, WithScanRow(f))...)
	return row, err
}

// QueryAsRows query rows as [][]string.
func (s SQL) QueryAsRows(db *sql.DB, optionFns ...QueryOptionFn) ([][]string, error) {
	rowsData := make([][]string, 0)
	f := func(rows *sql.Rows, rowIndex int, columns []string) (bool, error) {
		if m, err := scanStringRow(rows, columns); err != nil {
			return false, err
		} else {
			rowsData = append(rowsData, m)
			return true, nil
		}
	}

	if err := s.QueryRaw(db, append(optionFns, WithScanRow(f))...); err != nil {
		return nil, err
	}

	return rowsData, nil
}

func scanStringRow(rows *sql.Rows, columns []string) ([]string, error) {
	holders, err := scanRow(len(columns), rows)
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

	for rn := 0; r.Next() && option.allowRowNum(rn+1); rn++ {
		if continued, err := option.Scanner.ScanRow(r, rn, columns); err != nil {
			return err
		} else if !continued {
			break
		}
	}

	return nil
}

func scanRow(columnSize int, r *sql.Rows) ([]sql.NullString, error) {
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

	log.Printf("I! execute [%s] with [%v]", s.Query, s.Vars)
	r, err := db.Query(s.Query, s.Vars...)
	if err != nil {
		return nil, nil, nil, err
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, nil, nil, err
	}

	return option, r, columns, nil
}
