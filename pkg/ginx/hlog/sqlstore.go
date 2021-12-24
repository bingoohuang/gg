package hlog

import (
	"github.com/bingoohuang/gg/pkg/sqx"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SQLStore stores the log into database.
type SQLStore struct {
	DB        sqx.SqxDB
	LogTables []string

	TableCols map[string]*tableSchema
}

// NewSQLStore creates a new SQLStore.
func NewSQLStore(db sqx.SqxDB, defaultLogTables ...string) *SQLStore {
	s := &SQLStore{DB: db}
	s.LogTables = defaultLogTables
	s.TableCols = make(map[string]*tableSchema)

	return s
}

func (s *SQLStore) loadTableSchema(tableName string) (*tableSchema, error) {
	if v, ok := s.TableCols[tableName]; ok {
		return v, nil
	}

	sqy := sqx.NewSQL(`
		 select column_name, column_comment, data_type, extra, is_nullable nullable,
			character_maximum_length max_length
		 from information_schema.columns
		 where table_schema = database()
		 and table_name = ?`, tableName)
	var tableCols []TableCol
	if err := sqy.Query(s.DB, &tableCols); err != nil {
		return nil, err
	}

	v := &tableSchema{
		Name: tableName,
		Cols: tableCols,
	}
	v.createInsertSQL()
	s.TableCols[tableName] = v
	return v, nil
}

// TableCol defines the schema of a table.
type TableCol struct {
	Name      string `name:"column_name"`
	Comment   string `name:"column_comment"`
	DataType  string `name:"data_type"`
	Extra     string `name:"extra"`
	Nullable  string `name:"nullable"`
	MaxLength int    `name:"max_length"`

	ValueGetter col `name:"-"`
}

func (s *TableCol) IsNullable() bool {
	return strings.HasPrefix(strings.ToLower(s.Nullable), "y")
}

// Store stores the log in database like MySQL, InfluxDB, and etc.
func (s *SQLStore) Store(c *gin.Context, l *Log) {
	tables := l.Option.Tables
	if len(tables) == 0 {
		tables = s.LogTables
	}

	for _, t := range tables {
		schema, err := s.loadTableSchema(t)
		if err != nil {
			logrus.Errorf("failed to loadTableSchema for table %s, error: %v", t, err)
			continue
		}

		schema.log(s.DB, l)
	}
}

type tableSchema struct {
	Name         string
	Cols         []TableCol
	InsertSQL    string
	ValueGetters []col
}

func (t tableSchema) log(db sqx.SqxDB, l *Log) {
	if len(t.ValueGetters) == 0 {
		return
	}

	params := make([]interface{}, len(t.ValueGetters))
	for i, vg := range t.ValueGetters {
		params[i] = vg.get(l)
	}

	if result, err := sqx.NewSQL(t.InsertSQL, params...).Update(db); err != nil {
		logrus.Warnf("do insert error: %v", err)
	} else {
		logrus.Debugf("log result %+v", result)
	}
}

func (t *tableSchema) createInsertSQL() {
	colsNum := len(t.Cols)
	if colsNum == 0 {
		logrus.Warnf("table %s not found", t.Name)

		return
	}

	getters := make([]col, 0, colsNum)
	columns := make([]string, 0, colsNum)
	marks := make([]string, 0, colsNum)

	for _, c := range t.Cols {
		c.parseComment()

		if c.ValueGetter == nil {
			continue
		}

		columns = append(columns, c.Name)
		marks = append(marks, "?")
		getters = append(getters, c.ValueGetter)
	}

	t.InsertSQL = "insert into " + t.Name + "(" +
		strings.Join(columns, ",") +
		") values(" +
		strings.Join(marks, ",") + ")"
	t.ValueGetters = getters
}
