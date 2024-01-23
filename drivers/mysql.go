package drivers

import (
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

// TODO: Instead of passing the table name and having to create a new connection
// for every request, implement some type of database pooling.
func CreateMysqlDataProvider (
	user, password, database_name, table_name string,
	columns []Column,
) (*core.DataProvider, error) {

	return &core.DataProvider{
		All: func(offset int, count int) ([]map[string]interface{}, error) {
			// db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", user, password, database_name))
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
			return nil, fmt.Errorf("findone unimpl")
		},
	}, nil

	// return nil, fmt.Errorf("mysql data provider unimpl")
}
