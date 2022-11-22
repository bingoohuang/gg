package sqx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"reflect"
	"strconv"
	"strings"

	"github.com/bingoohuang/gg/pkg/mapstruct"
)

// QueryAsNumber executes a query which only returns number like count(*) sql.
func (s SQL) QueryAsNumber(db SqxDB) (int64, error) {
	str, err := s.QueryAsString(db)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(str, 10, 64)
}

// QueryAsString executes a query which only returns number like count(*) sql.
func (s SQL) QueryAsString(db SqxDB) (string, error) {
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
func (s SQL) Update(db SqxDB) (int64, error) {
	r, err := s.UpdateRaw(db)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (s SQL) UpdateRaw(db SqxDB) (sql.Result, error) {
	if dbTypeAware, ok := db.(DBTypeAware); ok {
		dbType := dbTypeAware.GetDBType()
		cr, err := dbType.Convert(s.Q, s.ConvertOptions...)
		if err != nil {
			return nil, err
		}

		s.Q, s.Vars = cr.PickArgs(s.Vars)
	}

	if !s.NoLog {
		logQuery(s.Name, s.Q, s.Vars)
	}

	ctx, cancel := s.prepareContext()
	defer cancel()

	result, err := db.ExecContext(ctx, s.Q, s.Vars...)
	logQueryError(s.NoLog, s.Name, result, err)
	return result, err
}

type RowScannerInit interface {
	InitRowScanner(columns []string)
}

type RowScanner interface {
	ScanRow(columns []string, rows *sql.Rows, rowIndex int) (bool, error)
}

type ScanRowFn func(columns []string, rows *sql.Rows, rowIndex int) (bool, error)

func (s ScanRowFn) ScanRow(columns []string, rows *sql.Rows, rowIndex int) (bool, error) {
	return s(columns, rows, rowIndex)
}

// QueryOption defines the query options.
type QueryOption struct {
	MaxRows          int
	TagNames         []string
	Scanner          RowScanner
	LowerColumnNames bool

	ConvertOptionOptions []sqlparser.ConvertOption
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

// Query queries return with result.
func (s SQL) Query(db SqxDB, result interface{}, optionFns ...QueryOptionFn) error {
	err := s.query(db, result, optionFns...)
	logQueryError(s.NoLog, s.Name, nil, err)
	logRows(s.Name, GetQueryRows(result))
	return err
}

func GetQueryRows(dest interface{}) int {
	if dest == nil {
		return 0
	}

	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return v.Len()
	default:
		return 1
	}
}

func (s SQL) query(db SqxDB, result interface{}, optionFns ...QueryOptionFn) error {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	elem := resultValue.Elem()
	elemKind := elem.Kind()
	if elemKind == reflect.Ptr { // 如果依然是指针
		typ := elem.Type().Elem() // 获取二级指针底层类型
		val := reflect.New(typ)   // 创新底层类型对象
		err := s.Query(db, val.Interface(), optionFns...)
		if err == nil {
			elem.Set(val) // 赋予一级指针新对象地址
		}
		return err
	}

	option := QueryOptionFns(optionFns).Options()

	var err error
	var input interface{}

	options := WithOptions(option)
	switch elemKind {
	case reflect.Struct:
		input, err = s.QueryAsMap(db, options)
	case reflect.Slice:
		sliceElemType := elem.Type().Elem()
		switch sliceElemType.Kind() {
		case reflect.Struct:
			input, err = s.QueryAsMaps(db, options)
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			scanner := &Col1Scanner{}
			err = s.QueryRaw(db, options, WithRowScanner(scanner))
			input = scanner.Data
		default:
			return ErrNotSupported
		}
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		scanner := &Col1Scanner{MaxRows: 1}
		err = s.QueryRaw(db, options, WithRowScanner(scanner))
		if len(scanner.Data) > 0 {
			input = scanner.Data[0]
		}
	default:
		return ErrNotSupported
	}

	if err != nil {
		return err
	}

	decoder, err := mapstruct.NewDecoder(&mapstruct.Config{
		Result:   result,
		TagNames: option.TagNames,
		Squash:   true,
		WeakType: true,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

var ErrNotSupported = errors.New("sqx: Unsupported result type")

type Col1Scanner struct {
	Data    []string
	MaxRows int
}

func (s *Col1Scanner) ScanRow(columns []string, rows *sql.Rows, _ int) (bool, error) {
	if v, err := ScanSliceRow(rows, columns); err != nil {
		return false, err
	} else {
		s.Data = append(s.Data, v[0])
		return s.MaxRows == 0 || len(s.Data) < s.MaxRows, nil
	}
}

type MapScanner struct {
	Data    []map[string]string
	MaxRows int
}

func (s *MapScanner) Data0() map[string]string {
	if len(s.Data) == 0 {
		return nil
	}

	return s.Data[0]
}

func (s *MapScanner) ScanRow(columns []string, rows *sql.Rows, _ int) (bool, error) {
	if v, err := ScanMapRow(rows, columns); err != nil {
		return false, err
	} else {
		s.Data = append(s.Data, v)
		return s.MaxRows == 0 || len(s.Data) < s.MaxRows, nil
	}
}

// QueryAsMaps query rows as map slice.
func (s SQL) QueryAsMaps(db SqxDB, optionFns ...QueryOptionFn) ([]map[string]string, error) {
	scanner := &MapScanner{Data: make([]map[string]string, 0)}
	err := s.QueryRaw(db, append(optionFns, WithRowScanner(scanner))...)
	return scanner.Data, err
}

// QueryAsMap query a single row as a map return.
func (s SQL) QueryAsMap(db SqxDB, optionFns ...QueryOptionFn) (map[string]string, error) {
	scanner := &MapScanner{Data: make([]map[string]string, 0), MaxRows: 1}
	err := s.QueryRaw(db, append(optionFns, WithRowScanner(scanner))...)
	return scanner.Data0(), err
}

func ScanSliceRow(rows *sql.Rows, columns []string) ([]string, error) {
	holders, err := ScanRow(len(columns), rows)
	if err != nil {
		return nil, err
	}

	m := make([]string, len(columns))
	for i, h := range holders {
		m[i] = h.String()
	}

	return m, nil
}

func ScanMapRow(rows *sql.Rows, columns []string) (map[string]string, error) {
	holders, err := ScanRow(len(columns), rows)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for i, h := range holders {
		m[columns[i]] = h.String()
	}

	return m, nil
}

type StringRowScanner struct {
	Data    [][]string
	MaxRows int
}

func (r *StringRowScanner) ScanRow(columns []string, rows *sql.Rows, _ int) (bool, error) {
	if m, err := ScanStringRow(rows, columns); err != nil {
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
func (s SQL) QueryAsRow(db SqxDB, optionFns ...QueryOptionFn) ([]string, error) {
	f := &StringRowScanner{MaxRows: 1}
	if err := s.QueryRaw(db, append(optionFns, WithRowScanner(f))...); err != nil {
		return nil, err
	}

	return f.Data0(), nil
}

// QueryAsRows query rows as [][]string.
func (s SQL) QueryAsRows(db SqxDB, optionFns ...QueryOptionFn) ([][]string, error) {
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
		m[i] = h.String()
	}
	return m, nil
}

// QueryRaw query rows for customized row scanner.
func (s SQL) QueryRaw(db SqxDB, optionFns ...QueryOptionFn) error {
	option, r, columns, err := s.prepareQuery(db, optionFns...)
	if err != nil {
		return err
	}

	defer r.Close()

	if initial, ok := option.Scanner.(RowScannerInit); ok {
		initial.InitRowScanner(columns)
	}

	rows := 0
	for rn := 0; r.Next() && option.allowRowNum(rn+1); rn++ {
		rows++
		if continued, err := option.Scanner.ScanRow(columns, r, rn); err != nil {
			return err
		} else if !continued {
			break
		}
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func ScanRowValues(rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	row, err := ScanRow(len(cols), rows)
	if err != nil {
		return nil, err
	}

	rowValues := make([]interface{}, len(cols))
	for i := range rowValues {
		rowValues[i] = row[i].Get()
	}

	return rowValues, nil
}

func ScanRow(columnSize int, r *sql.Rows) ([]NullAny, error) {
	holders := make([]NullAny, columnSize)
	pointers := make([]interface{}, columnSize)
	for i := 0; i < columnSize; i++ {
		pointers[i] = &holders[i]
	}

	if err := r.Scan(pointers...); err != nil {
		return nil, err
	}

	return holders, nil
}

func (s SQL) prepareContext() (ctx context.Context, cancel func()) {
	ctx = s.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if s.Timeout > 0 {
		return context.WithTimeout(ctx, s.Timeout)
	}

	return ctx, func() {}
}
func (s *SQL) prepareQuery(db SqxDB, optionFns ...QueryOptionFn) (*QueryOption, *sql.Rows, []string, error) {
	if err := s.adaptQuery(db); err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := s.prepareContext()
	defer cancel()
	ctx = context.WithValue(ctx, AdaptedKey, s.adapted)
	r, err := db.QueryContext(ctx, s.Q, s.Vars...)

	if err != nil {
		return nil, nil, nil, err
	}

	columns, err := r.Columns()
	if err != nil {
		return nil, nil, nil, err
	}

	option := QueryOptionFns(optionFns).Options()
	if option.LowerColumnNames {
		for i, col := range columns {
			columns[i] = strings.ToLower(col)
		}
	}

	return option, r, columns, nil
}
