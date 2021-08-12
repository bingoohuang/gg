package sqlrun

import (
	"database/sql"
	"reflect"
)

// NewStructPreparer creates a new StructPreparer.
func NewStructPreparer(v interface{}) *StructPreparer {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic(t.String() + " hasn't' not a struct or pointer to ptr type")
	}

	return &StructPreparer{
		StructType: t,
	}
}

// StructPreparer is the the structure to create struct mapping.
type StructPreparer struct {
	StructType reflect.Type
}

// Prepare prepares to scan query rows.
func (m *StructPreparer) Prepare(rows *sql.Rows, columns []string) Mapping {
	return &StructMapping{
		rows:           rows,
		mapFields:      m.newStructFields(columns),
		StructPreparer: m,
		rowsData:       reflect.MakeSlice(reflect.SliceOf(m.StructType), 0, 0),
	}
}

// StructMapping is the structure for mapping row to a structure.
type StructMapping struct {
	mapFields selectItemSlice
	*StructPreparer
	rows     *sql.Rows
	rowsData reflect.Value
}

// Scan scans the query result to fetch the rows one by one.
func (s *StructMapping) Scan(rowNum int) error {
	pointers, structPtr := s.mapFields.ResetDestinations(s.StructPreparer)

	err := s.rows.Scan(pointers...)
	if err != nil {
		return err
	}

	for i, field := range s.mapFields {
		if p, ok := pointers[i].(*NullAny); ok {
			field.SetField(p.getVal())
		} else {
			field.SetField(reflect.ValueOf(pointers[i]).Elem())
		}
	}

	s.rowsData = reflect.Append(s.rowsData, structPtr.Elem())

	return nil
}

// RowsData returns the mapped rows data.
func (s *StructMapping) RowsData() interface{} { return s.rowsData.Interface() }

func (mapFields selectItemSlice) ResetDestinations(mapper *StructPreparer) ([]interface{}, reflect.Value) {
	pointers := make([]interface{}, len(mapFields))
	structPtr := reflect.New(mapper.StructType)

	for i, fv := range mapFields {
		fv.SetRoot(structPtr.Elem())

		if ImplSQLScanner(fv.Type()) {
			pointers[i] = reflect.New(fv.Type()).Interface()
		} else {
			pointers[i] = &NullAny{Type: fv.Type()}
		}
	}

	return pointers, structPtr
}
