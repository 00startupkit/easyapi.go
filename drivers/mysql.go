package drivers

import (
	"fmt"

	"github.com/00startupkit/easyapi.go/core"
	_ "github.com/go-sql-driver/mysql"
)

func CreateMysqlDataProvider () (*core.DataProvider, error) {
	return nil, fmt.Errorf("unimpl")
}
