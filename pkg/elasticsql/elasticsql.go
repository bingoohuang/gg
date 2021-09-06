package elasticsql

import (
	"errors"
	"fmt"
	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"github.com/bingoohuang/gg/pkg/ss"
	"strings"
)

// Convert will transform sql to elasticsearch dsl string
func Convert(sql string) (dsl string, err error) {
	switch firstWord := strings.ToLower(ss.FirstWord(sql)); firstWord {
	case "update", "delete", "insert":
		return "", errors.New("unsupported")
	case "limit", "order", "where":
		sql = "select * from t " + sql
	default:
		sql = "select * from t where " + sql
	}

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}

	//sql valid, start to handle
	switch t := stmt.(type) {
	case *sqlparser.Select:
		return handleSelect(t)
	default:
		return "", errors.New("unsupported")
	}
}

func handleSelect(sel *sqlparser.Select) (dsl string, err error) {
	// Handle where
	// top level node pass in an empty interface
	// to tell the children this is root
	// is there any better way?
	var rootParent sqlparser.Expr
	var queryMapStr string

	// use may not pass where clauses
	if sel.Where != nil {
		queryMapStr, err = handleSelectWhere(&sel.Where.Expr, true, &rootParent)
		if err != nil {
			return "", err
		}
	}
	if queryMapStr == "" {
		queryMapStr = `{"bool" : {"must": [{"match_all" : {}}]}}`
	}

	queryFrom, querySize := "", ""

	// if the request is to aggregation
	// then set aggFlag to true, and querySize to 0
	// to not return any query result

	sel.GroupBy = nil
	sel.SelectExprs = nil

	// Handle limit
	if sel.Limit != nil {
		if sel.Limit.Offset != nil {
			queryFrom = sqlparser.String(sel.Limit.Offset)
		}
		querySize = sqlparser.String(sel.Limit.Rowcount)
	}

	// Handle order by
	// when executing aggregations, order by is useless
	var orderByArr []string
	for _, orderByExpr := range sel.OrderBy {
		s := strings.Replace(sqlparser.String(orderByExpr.Expr), "`", "", -1)
		orderByStr := fmt.Sprintf(`{"%v": "%v"}`, s, orderByExpr.Direction)
		orderByArr = append(orderByArr, orderByStr)
	}

	resultMap := map[string]interface{}{"query": queryMapStr}

	if querySize != "" {
		resultMap["size"] = ss.ParseInt(querySize)
	}
	if queryFrom != "" {
		resultMap["from"] = ss.ParseInt(queryFrom)
	}

	if len(orderByArr) > 0 {
		resultMap["sort"] = fmt.Sprintf("[%v]", strings.Join(orderByArr, ","))
	}

	// keep the traversal in order, avoid unpredicted json
	var resultArr []string
	for _, mapKey := range []string{"query", "from", "size", "sort"} {
		if val, ok := resultMap[mapKey]; ok {
			resultArr = append(resultArr, fmt.Sprintf(`"%v" : %v`, mapKey, val))
		}
	}

	dsl = "{" + strings.Join(resultArr, ",") + "}"
	return dsl, nil
}

func buildNestedFuncStrValue(nestedFunc *sqlparser.FuncExpr) (string, error) {
	return "", errors.New("elasticsql: unsupported function" + nestedFunc.Name.String())
}

func handleSelectWhereAndExpr(expr *sqlparser.Expr, parent *sqlparser.Expr) (string, error) {
	andExpr := (*expr).(*sqlparser.AndExpr)
	leftExpr := andExpr.Left
	rightExpr := andExpr.Right
	leftStr, err := handleSelectWhere(&leftExpr, false, expr)
	if err != nil {
		return "", err
	}
	rightStr, err := handleSelectWhere(&rightExpr, false, expr)
	if err != nil {
		return "", err
	}

	// not toplevel
	// if the parent node is also and, then the result can be merged

	var resultStr string
	if leftStr == "" || rightStr == "" {
		resultStr = leftStr + rightStr
	} else {
		resultStr = leftStr + `,` + rightStr
	}

	if _, ok := (*parent).(*sqlparser.AndExpr); ok {
		return resultStr, nil
	}
	return fmt.Sprintf(`{"bool" : {"must" : [%v]}}`, resultStr), nil
}

func handleSelectWhereOrExpr(expr *sqlparser.Expr, parent *sqlparser.Expr) (string, error) {
	orExpr := (*expr).(*sqlparser.OrExpr)
	leftExpr := orExpr.Left
	rightExpr := orExpr.Right

	leftStr, err := handleSelectWhere(&leftExpr, false, expr)
	if err != nil {
		return "", err
	}

	rightStr, err := handleSelectWhere(&rightExpr, false, expr)
	if err != nil {
		return "", err
	}

	var resultStr string
	if leftStr == "" || rightStr == "" {
		resultStr = leftStr + rightStr
	} else {
		resultStr = leftStr + `,` + rightStr
	}

	// not toplevel
	// if the parent node is also or node, then merge the query param
	if _, ok := (*parent).(*sqlparser.OrExpr); ok {
		return resultStr, nil
	}

	return fmt.Sprintf(`{"bool" : {"should" : [%v]}}`, resultStr), nil
}

func buildComparisonExprRightStr(expr sqlparser.Expr) (string, bool, error) {
	var rightStr string
	var err error
	switch expr.(type) {
	case *sqlparser.SQLVal:
		rightStr = sqlparser.String(expr)
		rightStr = strings.Trim(rightStr, `'`)
	case *sqlparser.GroupConcatExpr:
		return "", false, errors.New("elasticsql: group_concat not supported")
	case *sqlparser.FuncExpr:
		// parse nested
		funcExpr := expr.(*sqlparser.FuncExpr)
		rightStr, err = buildNestedFuncStrValue(funcExpr)
		if err != nil {
			return "", false, err
		}
	case *sqlparser.ColName:
		if sqlparser.String(expr) == "missing" {
			return "", true, nil
		}

		return "", true, errors.New("elasticsql: column name on the right side of compare operator is not supported")
	case sqlparser.ValTuple:
		rightStr = sqlparser.String(expr)
	default:
		// cannot reach here
	}
	return rightStr, false, err
}

func handleSelectWhereComparisonExpr(expr *sqlparser.Expr, topLevel bool, parent *sqlparser.Expr) (string, error) {
	comparisonExpr := (*expr).(*sqlparser.ComparisonExpr)
	colName, ok := comparisonExpr.Left.(*sqlparser.ColName)

	if !ok {
		return "", errors.New("elasticsql: invalid comparison expression, the left must be a column name")
	}

	colNameStr := sqlparser.String(colName)
	colNameStr = strings.Replace(colNameStr, "`", "", -1)
	rightStr, missingCheck, err := buildComparisonExprRightStr(comparisonExpr.Right)
	if err != nil {
		return "", err
	}

	resultStr := ""

	switch comparisonExpr.Operator {
	case ">=":
		resultStr = fmt.Sprintf(`{"range" : {"%v" : {"from" : "%v"}}}`, colNameStr, rightStr)
	case "<=":
		resultStr = fmt.Sprintf(`{"range" : {"%v" : {"to" : "%v"}}}`, colNameStr, rightStr)
	case "=":
		// field is missing
		if missingCheck {
			resultStr = fmt.Sprintf(`{"missing":{"field":"%v"}}`, colNameStr)
		} else {
			resultStr = fmt.Sprintf(`{"match" : {"%v" : {"query" : "%v"}}}`, colNameStr, rightStr)
		}
	case ">":
		resultStr = fmt.Sprintf(`{"range" : {"%v" : {"gt" : "%v"}}}`, colNameStr, rightStr)
	case "<":
		resultStr = fmt.Sprintf(`{"range" : {"%v" : {"lt" : "%v"}}}`, colNameStr, rightStr)
	case "!=":
		if missingCheck {
			resultStr = fmt.Sprintf(`{"bool" : {"must_not" : [{"missing":{"field":"%v"}}]}}`, colNameStr)
		} else {
			resultStr = fmt.Sprintf(`{"bool" : {"must_not" : [{"match" : {"%v" : {"query" : "%v"}}}]}}`, colNameStr, rightStr)
		}
	case "in":
		// the default valTuple is ('1', '2', '3') like
		// so need to drop the () and replace ' to "
		rightStr = strings.Replace(rightStr, `'`, `"`, -1)
		rightStr = strings.Trim(rightStr, "(")
		rightStr = strings.Trim(rightStr, ")")
		resultStr = fmt.Sprintf(`{"terms" : {"%v" : [%v]}}`, colNameStr, rightStr)
	case "like":
		rightStr = strings.Replace(rightStr, `%`, ``, -1)
		resultStr = fmt.Sprintf(`{"match" : {"%v" : {"query" : "%v"}}}`, colNameStr, rightStr)
	case "not like":
		rightStr = strings.Replace(rightStr, `%`, ``, -1)
		resultStr = fmt.Sprintf(`{"bool" : {"must_not" : {"match" : {"%v" : {"query" : "%v"}}}}}`, colNameStr, rightStr)
	case "not in":
		// the default valTuple is ('1', '2', '3') like
		// so need to drop the () and replace ' to "
		rightStr = strings.Replace(rightStr, `'`, `"`, -1)
		rightStr = strings.Trim(rightStr, "(")
		rightStr = strings.Trim(rightStr, ")")
		resultStr = fmt.Sprintf(`{"bool" : {"must_not" : {"terms" : {"%v" : [%v]}}}}`, colNameStr, rightStr)
	}

	// the root node need to have bool and must
	if topLevel {
		resultStr = fmt.Sprintf(`{"bool" : {"must" : [%v]}}`, resultStr)
	}

	return resultStr, nil
}

func handleSelectWhere(expr *sqlparser.Expr, topLevel bool, parent *sqlparser.Expr) (string, error) {
	if expr == nil {
		return "", errors.New("elasticsql: error expression cannot be nil here")
	}

	switch e := (*expr).(type) {
	case *sqlparser.AndExpr:
		return handleSelectWhereAndExpr(expr, parent)

	case *sqlparser.OrExpr:
		return handleSelectWhereOrExpr(expr, parent)
	case *sqlparser.ComparisonExpr:
		return handleSelectWhereComparisonExpr(expr, topLevel, parent)

	case *sqlparser.IsExpr:
		return "", errors.New("elasticsql: is expression currently not supported")
	case *sqlparser.RangeCond:
		// between a and b
		// the meaning is equal to range query
		rangeCond := (*expr).(*sqlparser.RangeCond)
		colName, ok := rangeCond.Left.(*sqlparser.ColName)

		if !ok {
			return "", errors.New("elasticsql: range column name missing")
		}

		colNameStr := sqlparser.String(colName)
		colNameStr = strings.Replace(colNameStr, "`", "", -1)
		fromStr := strings.Trim(sqlparser.String(rangeCond.From), `'`)
		toStr := strings.Trim(sqlparser.String(rangeCond.To), `'`)

		resultStr := fmt.Sprintf(`{"range" : {"%v" : {"from" : "%v", "to" : "%v"}}}`, colNameStr, fromStr, toStr)
		if topLevel {
			resultStr = fmt.Sprintf(`{"bool" : {"must" : [%v]}}`, resultStr)
		}

		return resultStr, nil

	case *sqlparser.ParenExpr:
		parentBoolExpr := (*expr).(*sqlparser.ParenExpr)
		boolExpr := parentBoolExpr.Expr

		// if paren is the top level, bool must is needed
		var isThisTopLevel = false
		if topLevel {
			isThisTopLevel = true
		}
		return handleSelectWhere(&boolExpr, isThisTopLevel, parent)
	case *sqlparser.NotExpr:
		return "", errors.New("elasticsql: not expression currently not supported")
	case *sqlparser.FuncExpr:
		switch e.Name.Lowered() {
		case "multi_match":
			params := e.Exprs
			if len(params) > 3 || len(params) < 2 {
				return "", errors.New("elasticsql: the multi_match must have 2 or 3 params, (query, fields and type) or (query, fields)")
			}

			var typ, query, fields string
			for i := 0; i < len(params); i++ {
				elem := strings.Replace(sqlparser.String(params[i]), "`", "", -1) // a = b
				kv := strings.Split(elem, "=")
				if len(kv) != 2 {
					return "", errors.New("elasticsql: the param should be query = xxx, field = yyy, type = zzz")
				}
				k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
				switch k {
				case "type":
					typ = strings.Replace(v, "'", "", -1)
				case "query":
					query = strings.Replace(v, "`", "", -1)
					query = strings.Replace(query, "'", "", -1)
				case "fields":
					fieldList := strings.Split(strings.TrimRight(strings.TrimLeft(v, "("), ")"), ",")
					for idx, field := range fieldList {
						fieldList[idx] = fmt.Sprintf(`"%v"`, strings.TrimSpace(field))
					}
					fields = strings.Join(fieldList, ",")
				default:
					return "", errors.New("elaticsql: unknow param for multi_match")
				}
			}
			if typ == "" {
				return fmt.Sprintf(`{"multi_match" : {"query" : "%v", "fields" : [%v]}}`, query, fields), nil
			}
			return fmt.Sprintf(`{"multi_match" : {"query" : "%v", "type" : "%v", "fields" : [%v]}}`, query, typ, fields), nil
		default:
			return "", errors.New("elaticsql: function in where not supported" + e.Name.Lowered())
		}
	}

	return "", errors.New("elaticsql: logically cannot reached here")
}
