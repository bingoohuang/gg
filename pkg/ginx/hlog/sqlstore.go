package hlog

import (
	"database/sql"
	"github.com/bingoohuang/gg/pkg/ginx/sqlrun"
	"strings"

	"github.com/sirupsen/logrus"
)

// SQLStore stores the log into database.
type SQLStore struct {
	DB         *sql.DB
	DriverName string
	LogTables  []string

	TableCols map[string]*tableSchema
}

// NewSQLStore creates a new SQLStore.
func NewSQLStore(db *sql.DB, defaultLogTables ...string) *SQLStore {
	s := &SQLStore{DB: db}
	s.DriverName = sqlrun.LookupDriverName(db.Driver())
	s.LogTables = defaultLogTables
	s.TableCols = make(map[string]*tableSchema)

	return s
}

func (s *SQLStore) loadTableSchema(tableName string) (*tableSchema, error) {
	if v, ok := s.TableCols[tableName]; ok {
		return v, nil
	}

	run := sqlrun.NewSQLRun(s.DB, sqlrun.NewStructPreparer(TableCol{}))

	result := run.DoQuery(`
		 select column_name, column_comment, data_type, extra, is_nullable nullable,
			character_maximum_length max_length
		 from information_schema.columns
		 where table_schema = database()
		 and table_name = ?`, tableName)
	if result.Error != nil {
		return nil, result.Error
	}

	tableCols := result.Rows.([]TableCol)

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
func (s *SQLStore) Store(l *Log) {
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

func (t tableSchema) log(db sqlrun.MiniDB, l *Log) {
	if len(t.ValueGetters) == 0 {
		return
	}

	params := make([]interface{}, len(t.ValueGetters))
	for i, vg := range t.ValueGetters {
		params[i] = vg.get(l)
	}

	result := sqlrun.NewSQLExec(db).DoUpdate(t.InsertSQL, params...)

	if result.Error != nil {
		logrus.Warnf("do insert error: %v", result.Error)
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
