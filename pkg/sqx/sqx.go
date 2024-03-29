package sqx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/gg/pkg/strcase"
)

// ErrConditionKind tells that the condition kind should be struct or its pointer
var ErrConditionKind = errors.New("condition kind should be struct or its pointer")

// SQL is a structure for query and vars.
type SQL struct {
	Name    string
	Q       string
	AppendQ string
	Vars    []interface{}
	Ctx     context.Context
	NoLog   bool

	Timeout        time.Duration
	Limit          int
	ConvertOptions []sqlparser.ConvertOption

	adapted bool
}

func (s *SQL) AppendIf(ok bool, sub string, args ...interface{}) *SQL {
	if !ok {
		return s
	}

	return s.Append(sub, args...)
}

// Append appends sub statement to the query.
func (s *SQL) Append(sub string, args ...interface{}) *SQL {
	if sub == "" {
		return s
	}

	if strings.HasPrefix(sub, " ") {
		s.Q += sub
	} else {
		s.Q += " " + sub
	}

	s.Vars = append(s.Vars, args...)

	return s
}

// NewSQL create s SQL object.
func NewSQL(query string, vars ...interface{}) *SQL {
	return &SQL{Q: query, Vars: vars}
}

// WithVars replace vars.
func WithVars(vars ...interface{}) []interface{} { return vars }

// WithConvertOptions set SQL conversion options.
func (s *SQL) WithConvertOptions(convertOptions []sqlparser.ConvertOption) *SQL {
	s.ConvertOptions = convertOptions
	return s
}

// WithTimeout set sql execution timeout
func (s *SQL) WithTimeout(timeout time.Duration) *SQL {
	s.Timeout = timeout
	return s
}

// WithVars replace vars.
func (s *SQL) WithVars(vars ...interface{}) *SQL {
	s.Vars = vars
	return s
}

func (s *SQL) AndIf(ok bool, cond string, args ...interface{}) *SQL {
	if !ok {
		return s
	}

	return s.And(cond, args...)
}

func (s *SQL) And(cond string, args ...interface{}) *SQL {
	switch len(args) {
	case 0:
		if !ss.ContainsFold(s.Q, "where") {
			s.Q += " where " + cond
		} else {
			s.Q += " and " + cond
		}
		return s
	case 1:
		arg := reflect.ValueOf(args[0])
		if arg.IsZero() {
			return s
		}

		isSlice := arg.Kind() == reflect.Slice
		if isSlice && arg.Len() == 0 {
			return s
		}
		if isSlice && arg.Len() > 1 && strings.Count(cond, "?") == 1 {
			cond = strings.Replace(cond, "?", ss.Repeat("?", ",", arg.Len()), 1)
		}
		if !ss.ContainsFold(s.Q, "where") {
			s.Q += " where " + cond
		} else {
			s.Q += " and " + cond
		}

		if isSlice {
			for i := 0; i < arg.Len(); i++ {
				s.Vars = append(s.Vars, arg.Index(i).Interface())
			}
		} else {
			s.Vars = append(s.Vars, args[0])
		}
		return s
	default:
		panic("not supported")
	}
}

func (s *SQL) adaptUpdate(db SqxDB) error {
	if dbTypeAware, ok := db.(DBTypeAware); ok {
		dbType := dbTypeAware.GetDBType()
		options := s.ConvertOptions
		cr, err := dbType.Convert(s.Q, options...)
		if err != nil {
			return err
		}

		s.Q, s.Vars = cr.PickArgs(s.Vars)
	}

	if !s.NoLog {
		logQuery(s.Name, s.Q, s.Vars)
	}

	return nil
}

func (s *SQL) adaptQuery(db SqxDB) error {
	if dbTypeAware, ok := db.(DBTypeAware); ok {
		dbType := dbTypeAware.GetDBType()
		options := s.ConvertOptions
		if s.Limit > 0 {
			options = append([]sqlparser.ConvertOption{sqlparser.WithLimit(s.Limit)}, options...)
		}
		cr, err := dbType.Convert(s.Q, options...)
		if err != nil {
			return err
		}

		s.Q, s.Vars = cr.PickArgs(s.Vars)
		if s.AppendQ != "" {
			s.Q += " " + s.AppendQ
		}

		s.adapted = true
	}

	if !s.NoLog {
		logQuery(s.Name, s.Q, s.Vars)
	}

	return nil
}

// CreateSQL creates a composite SQL on base and condition cond.
func CreateSQL(base string, cond interface{}) (*SQL, error) {
	result := &SQL{}
	if cond == nil {
		result.Q = base
		return result, nil
	}

	vc, err := inferenceCondValue(cond)
	if err != nil {
		return nil, err
	}

	condSql, vars, err := iterateFields(vc)
	if err != nil {
		return nil, err
	}

	if condSql == "" {
		result.Q = base
		return result, nil
	}

	result.Vars = vars

	parsed, err := sqlparser.Parse(base)
	if err != nil {
		return nil, err
	}

	iw, ok := parsed.(sqlparser.IWhere)
	if !ok {
		return result, nil
	}

	x := `select 1 from t where ` + createNewWhere(iw, condSql)
	condParsed, err := sqlparser.Parse(x)
	if err != nil {
		return nil, err
	}

	iw.SetWhere(condParsed.(*sqlparser.Select).Where)
	result.Q = sqlparser.String(parsed)

	return result, nil
}

func createNewWhere(iw sqlparser.IWhere, condSql string) string {
	where := iw.GetWhere()
	if where == nil {
		return condSql
	}

	whereString := sqlparser.String(where)
	if _, ok := where.Expr.(*sqlparser.OrExpr); ok {
		return `(` + whereString[7:] + `) and ` + condSql
	}

	return `` + whereString[7:] + ` and ` + condSql
}

func inferenceCondValue(cond interface{}) (reflect.Value, error) {
	vc := reflect.ValueOf(cond)
	if vc.Kind() == reflect.Ptr {
		vc = vc.Elem()
	}

	if vc.Kind() != reflect.Struct {
		return reflect.Value{}, ErrConditionKind
	}

	return vc, nil
}

const andPrefix = " and "

func iterateFields(vc reflect.Value) (string, []interface{}, error) {
	condSql := ""
	vars := make([]interface{}, 0)
	t := vc.Type()

	for i := 0; i < vc.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // not exported
			continue
		}

		cond := f.Tag.Get("cond")
		if cond == "-" { // ignore as a condition field
			continue
		}

		v := vc.Field(i)
		if f.Anonymous {
			embeddedSQL, embeddedVars, err := iterateFields(v)
			if err != nil {
				return "", nil, err
			}

			condSql += andPrefix + embeddedSQL
			vars = append(vars, embeddedVars...)
			continue
		}

		cond, fieldVars, err := processTag(f.Tag, f.Name, v)
		if err != nil {
			return "", nil, err
		}

		if cond != "" {
			condSql += andPrefix + cond
			vars = append(vars, fieldVars...)
		}
	}

	if condSql != "" {
		condSql = condSql[len(andPrefix):]
	}

	return condSql, vars, nil
}

func processTag(tag reflect.StructTag, fieldName string, v reflect.Value) (cond string, vars []interface{}, err error) {
	cond = tag.Get("cond")
	zero := tag.Get("zero")
	if yes, err1 := isZero(v, zero); err1 != nil {
		return "", nil, err1
	} else if yes { // ignore zero field
		return "", nil, nil
	}

	if cond == "" {
		cond = strcase.ToSnake(fieldName) + "=?"
	}

	vi := v.Interface()
	if modifier := tag.Get("modifier"); modifier != "" {
		vi = strings.ReplaceAll(modifier, "v", fmt.Sprintf("%v", vi))
	}

	for i := 0; i < strings.Count(cond, "?"); i++ {
		vars = append(vars, vi)
	}
	return
}

func isZero(v reflect.Value, zero string) (bool, error) {
	if zero == "" {
		return v.IsZero(), nil
	}

	switch v.Kind() {
	case reflect.String:
		return zero == v.Interface(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		zeroV, err := strconv.ParseInt(zero, 10, 64)
		if err != nil {
			return false, err
		}
		return zeroV == v.Convert(TypeInt64).Interface(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		zeroV, err := strconv.ParseUint(zero, 10, 64)
		if err != nil {
			return false, err
		}
		return zeroV == v.Convert(TypeUint64).Interface(), nil
	case reflect.Float32, reflect.Float64:
		zeroV, err := strconv.ParseFloat(zero, 64)
		if err != nil {
			return false, err
		}

		return zeroV == v.Convert(TypeFloat64).Interface(), nil
	case reflect.Bool:
		zeroV, err := strconv.ParseBool(zero)
		if err != nil {
			return false, err
		}
		return zeroV == v.Interface(), nil
	}

	return false, nil
}

var (
	TypeInt64   = reflect.TypeOf(int64(0))
	TypeUint64  = reflect.TypeOf(uint64(0))
	TypeFloat64 = reflect.TypeOf(float64(0))
)
