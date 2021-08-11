package sqx

// refer https://yougg.github.io/2017/08/24/用go语言写一个简单的mysql客户端/
import (
	"context"
	"database/sql"
	"github.com/bingoohuang/gg/pkg/ss"
	"time"
)

// Result defines the result structure of sql execution.
type Result struct {
	Error        error
	CostTime     time.Duration
	Headers      []string
	Rows         [][]string
	RowsAffected int64
	LastInsertID int64
	IsQuerySQL   bool
	FirstKey     string
}

func (r Result) Return(start time.Time, err error) Result {
	r.CostTime = time.Since(start)
	r.Error = err
	return r
}

type SQLExecContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type SqxDB interface {
	SQLExec
	SQLExecContext
}

// SQLExec wraps Exec method.
type SQLExec interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type ExecOption struct {
	MaxRows     int
	NullReplace string
	BlobReplace string
}

func (o ExecOption) reachMaxRows(row int) bool {
	return o.MaxRows > 0 && row >= o.MaxRows
}

// Exec executes a SQL.
func Exec(db SQLExec, query string, option ExecOption) Result {
	firstKey, isQuerySQL := IsQuerySQL(query)
	if isQuerySQL {
		return processQuery(db, query, firstKey, option)
	}

	return execNonQuery(db, query, firstKey)
}

func processQuery(db SQLExec, query string, firstKey string, option ExecOption) (r Result) {
	start := time.Now()
	r.FirstKey = firstKey
	r.IsQuerySQL = true

	rows, err := db.Query(query)
	if err != nil || rows != nil && rows.Err() != nil {
		if err == nil {
			err = rows.Err()
		}

		return r.Return(start, err)
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return r.Return(start, err)
	}

	r.Headers = columns

	columnSize := len(columns)
	columnTypes, _ := rows.ColumnTypes()
	data := make([][]string, 0)

	var columnLobs []bool
	if option.BlobReplace != "" {
		columnLobs = make([]bool, columnSize)
		for i := 0; i < len(columnTypes); i++ {
			columnLobs[i] = ss.ContainsFold(columnTypes[i].DatabaseTypeName(), "LOB")
		}
	}

	for row := 0; rows.Next() && !option.reachMaxRows(row); row++ {
		holders := make([]sql.NullString, columnSize)
		pointers := make([]interface{}, columnSize)

		for i := 0; i < columnSize; i++ {
			pointers[i] = &holders[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return r.Return(start, err)
		}

		values := make([]string, columnSize)

		for i, v := range holders {
			values[i] = ss.If(v.Valid, v.String, option.NullReplace)
			if option.BlobReplace != "" && v.Valid && columnLobs[i] {
				values[i] = "(" + columnTypes[i].DatabaseTypeName() + ")"
			}
		}

		data = append(data, values)
	}

	r.Rows = data

	return r.Return(start, nil)
}

func execNonQuery(db SQLExec, query string, firstKey string) Result {
	start := time.Now()
	r, err := db.Exec(query)

	var affected int64
	if r != nil {
		affected, _ = r.RowsAffected()
	}

	var lastInsertID int64
	if r != nil && firstKey == "INSERT" {
		lastInsertID, _ = r.LastInsertId()
	}

	return Result{
		Error:        err,
		CostTime:     time.Since(start),
		RowsAffected: affected,
		IsQuerySQL:   false,
		LastInsertID: lastInsertID,
		FirstKey:     firstKey,
	}
}
