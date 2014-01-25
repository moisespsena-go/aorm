package callback

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"time"
)

func Create(scope *gorm.Scope) {
}

func init() {
	gorm.DefaultCallback.Create().Before().Register(Create)
}

func query(db *DB) {
}

func save(db *DB) {
}

func create(db *DB) {
}

func update(db *DB) {
}

func Delete(scope *Scope) {
	scope.CallMethod("BeforeDelete")

	if !scope.HasError() {
		if !scope.Search.unscope && scope.HasColumn("DeletedAt") {
			scope.Raw(fmt.Sprintf("UPDATE %v SET deleted_at=%v %v", scope.Table(), scope.AddToVars(time.Now()), scope.CombinedSql()))
		} else {
			scope.Raw(fmt.Sprintf("DELETE FROM %v %v", scope.Table(), scope.CombinedSql()))
		}
		scope.Exec()
		scope.CallMethod("AfterDelete")
	}
}

func init() {
	DefaultCallback.Create().Before("Delete").After("Lalala").Register("delete", Delete)
	DefaultCallback.Update().Before("Delete").After("Lalala").Remove("replace", Delete)
	DefaultCallback.Delete().Before("Delete").After("Lalala").Replace("replace", Delete)
	DefaultCallback.Query().Before("Delete").After("Lalala").Replace("replace", Delete)
}

// Scope
// HasError(), HasColumn(), CallMethod(), Raw(), Exec()
// TableName(), CombinedQuerySQL()