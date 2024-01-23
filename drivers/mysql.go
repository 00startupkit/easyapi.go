package drivers

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/00startupkit/easyapi.go/core"
	// _ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ColType int
const (
	ColType_UNDEF ColType = 0
	ColType_INT ColType = 1
	ColType_STRING ColType = 2
)

type Column struct {
	Type ColType
	Name string
}

func column_field_names (columns []Column) []string {
	column_names := []string {}
	for _, c := range columns {
		column_names = append(column_names, c.Name)
	}
	return column_names
}

func constraint_comparison_to_sql (comparison core.Comparison) string {
	switch comparison {
		case core.Comparison_EQ:
			return "=" 
		case core.Comparison_LT:
			return "<" 
		case core.Comparison_LE:
			return "<=" 
		case core.Comparison_GT:
			return ">" 
		case core.Comparison_GE:
			return ">=" 
	}
	return "???"
}

func wrap_based_on_column_type (value string, column Column) string {
	switch column.Type {
	case ColType_INT:
		return value;
	case ColType_STRING:
		return fmt.Sprintf("\"%s\"", value)
	}
	return "???"
}

func constraint_to_sql_clause (constraint core.Constraint, column Column) string {
	return fmt.Sprintf(
		"%s %s %s",
		constraint.Property,
		constraint_comparison_to_sql(constraint.Comparison),
		wrap_based_on_column_type(constraint.Value, column))
}

func constraints_to_sql_clauses (constraints []core.Constraint, columns []Column) []string {
	clauses := []string{}
	for _, c := range constraints {
		var column Column
		found := false
		for _, col := range columns {
			if col.Name == c.Property {
				column = col
				found = true
				break
			}
		}
		if !found { continue } // skip constraint if the propoerty is not found.

		clauses = append(clauses, constraint_to_sql_clause(c, column))
	}
	return clauses
}

func fix_payload_types (payload map[string]interface{}, columns []Column) (map[string]interface{}, error) {

	fixed_payload := map[string]interface{}{}
	for k, v := range payload {

		var column Column
		found := false
		for _, c := range columns {
			if  c.Name == k {
				column = c
				found = true
				break
			}
		}

		if !found { return nil, fmt.Errorf(fmt.Sprintf("could not find column definition for field \"%s\" in the payload", k)) }

		bytearr, is_bytearr := v.([]byte)
		if !is_bytearr { return nil, fmt.Errorf("parsed payload is not a byte array") }

		switch column.Type {
		case ColType_INT:
			// Parse as int
			fixed_payload[k] = binary.BigEndian.Uint64(bytearr)
		case ColType_STRING:
			// Parse as string
			fixed_payload[k] = string(bytearr)
		default:
			return nil, fmt.Errorf(fmt.Sprintf("fix_payload_types unimplemented for type: %d", column.Type))
		}
	}
	return fixed_payload, nil
}

// TODO: Instead of passing the table name and having to create a new connection
// for every request, implement some type of database pooling.
func CreateMysqlDataProvider (
	user, password, database_name, table_name string,
	columns []Column,
) (*core.DataProvider, error) {

	return &core.DataProvider{
		All: func(offset int, count int) ([]map[string]interface{}, error) {
			db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@/%s", user, password, database_name))
			if err != nil { return nil, err }
			defer db.Close()

			query := fmt.Sprintf(
				`SELECT %s FROM %s LIMIT %d OFFSET %d`,
				strings.Join(column_field_names(columns), ","),
				table_name,
				count,
				offset,
			)

			rows, err := db.Queryx(query)
			if err != nil { return nil, err }
			defer rows.Close()

			entries := []map[string]interface{}{}
			for rows.Next() {
				entry := make(map[string]interface{})
				err = rows.MapScan(entry)
				if err != nil { return nil, err }

				entries = append(entries, entry)
			}
			return entries, nil
		},
		FindOne: func(constraints []core.Constraint) (*map[string]interface{}, error) {
			db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@/%s", user, password, database_name))
			if err != nil { return nil, err }
			defer db.Close()

			query := fmt.Sprintf(
				`SELECT %s FROM %s %s %s LIMIT 1`,
				strings.Join(column_field_names(columns), ","),
				table_name,
				func () string { if len(constraints) == 0 { return "" } else { return "WHERE" } }(),
				strings.Join(constraints_to_sql_clauses(constraints, columns), " AND "),
			)

			fmt.Printf("Query: %s\n", query)

			rows, err := db.Queryx(query)
			defer rows.Close()

			for rows.Next() {
				entry := make(map[string]interface{})
				err = rows.MapScan(entry)
				if err != nil { return nil, err }

				entry, err = fix_payload_types(entry, columns)
				if err != nil { return nil, err }

				return &entry, nil
			}
			return nil, fmt.Errorf("no entries found")
		},
	}, nil

	// return nil, fmt.Errorf("mysql data provider unimpl")
}
