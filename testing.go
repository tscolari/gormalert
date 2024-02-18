package gormalert

import (
	"testing"

	"gorm.io/gorm"
)

// ExpectDBWithoutSequentialScan will annotate the given db connection
// and will - in case it detects a table scan - fail the running test.
func ExpectDBWithoutSequentialScan(t *testing.T, db *gorm.DB) {
	options := AlertOptions{
		Name:  "scanalert-testing",
		Async: false,
		QueryTypes: []QueryType{
			CreateQuery, DeleteQuery, RawQuery, SelectQuery, UpdateQuery,
		},
		ErrorLogger: t.Log,
	}

	assertion := func(sourceQuery, _ string) {
		t.Errorf("the query %q executed a sequential scan", sourceQuery)
	}

	RegisterScanAlert(db, options, assertion)
}
