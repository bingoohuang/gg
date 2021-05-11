package sqx

import (
	"errors"

	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
)

// ErrNotSelect shows an error that the query is not a select statement.
var ErrNotSelect = errors.New("not a select query statement")

// CreateCount creates a count query sql.
func (s SQL) CreateCount() (*SQL, error) {
	parsed, err := sqlparser.Parse(s.Query)
	if err != nil {
		return nil, err
	}

	sel, ok := parsed.(*sqlparser.Select)
	if !ok {
		return nil, ErrNotSelect
	}

	limitVarsCount := 0
	if sel.Limit != nil {
		limitVarsCount++
		if sel.Limit.Offset != nil {
			limitVarsCount++
		}
	}

	sel.SelectExprs = countStar
	sel.OrderBy = nil
	sel.Having = nil
	sel.Limit = nil

	c := &SQL{
		Query: sqlparser.String(sel),
		Vars:  s.Vars,
		Ctx:   s.Ctx,
	}

	if limitVarsCount > 0 && len(s.Vars) >= limitVarsCount {
		c.Vars = s.Vars[:len(s.Vars)-limitVarsCount]
	}

	return c, nil
}

var countStar = func() sqlparser.SelectExprs {
	p, _ := sqlparser.Parse(`select count(*)`)
	return p.(*sqlparser.Select).SelectExprs
}()
