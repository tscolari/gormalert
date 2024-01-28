package gormalert

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

type QueryType string
type hookType string

const (
	// CreateQuery will scan for calls to `Create()`.
	CreateQuery QueryType = "create"
	// DeleteQuery will scan for calls to `Delete()`
	DeleteQuery QueryType = "delete"
	// RawQuery will scan for calls to `Raw()` or `Exec()`
	RawQuery QueryType = "raw"
	// SelectQuery will scan for calls to `Select()`
	SelectQuery QueryType = "query"
	// UpdateQuery will scan for calls to `Update()`.
	UpdateQuery QueryType = "update"
)

var explainFormat = map[string]string{
	"postgres": "EXPLAIN ",
	"mysql":    "EXPLAIN format=tree ",
}

var tableScanString = map[string]string{
	"postgres": "Seq Scan",
	"mysql":    "Table scan",
}

// AlertOptions contains a group of options to be used by the scanalert plugin.
type AlertOptions struct {
	// Name is the name of the plugin. In case you are registering multiple
	// scanalerts, this must be unique for each one of them.
	Name string

	// Async will perform the detection within a go routing, instead of blocking.
	Async bool

	// QueryType contains the selection of which kind of query this should be scanning for.
	// This filter only applies for explicit methods from gorm's DB object
	// (e.g. Update, First, Find, Create, etc...)
	QueryTypes []QueryType

	// ErrorLogger provides a way to flush out internal errors from the plugin.
	// If not selected errors will be ignored.
	ErrorLogger func(string)
}

func DefaultAlertOptions() AlertOptions {
	return AlertOptions{
		Name:  "scanalert",
		Async: false,
		QueryTypes: []QueryType{
			CreateQuery,
			DeleteQuery,
			RawQuery,
			SelectQuery,
			UpdateQuery,
		},
		ErrorLogger: func(msg string) {
			fmt.Fprintln(os.Stderr, msg)
		},
	}
}

type actionFunc func(sourceQuery string, scanResult string)

func RegisterScanAlert(db *gorm.DB, options AlertOptions, action actionFunc) {
	alerter := NewScanAlerterPlugin(options, action)
	db.Use(alerter)
}

func NewScanAlerterPlugin(options AlertOptions, action actionFunc) *scanAlerter {
	return &scanAlerter{
		options: options,
		action:  action,
	}
}

type scanAlerter struct {
	options AlertOptions
	action  actionFunc
}

func (s *scanAlerter) Name() string {
	return s.options.Name
}

func (s *scanAlerter) Initialize(db *gorm.DB) error {
	processor := db.Callback().Create()

	for _, queryType := range s.options.QueryTypes {
		switch queryType {
		case DeleteQuery:
			processor = db.Callback().Delete()
		case RawQuery:
			processor = db.Callback().Raw()
		case SelectQuery:
			processor = db.Callback().Query()
		case UpdateQuery:
			processor = db.Callback().Update()
		}

		scanFunc := s.Scan
		if s.options.Async {
			scanFunc = s.AsyncScan
		}

		processor.Register(s.Name()+"_"+string(queryType), scanFunc)
	}

	return nil
}

func (s *scanAlerter) AsyncScan(db *gorm.DB) {
	go s.Scan(db)
}

func (s *scanAlerter) Scan(db *gorm.DB) {
	statement := db.Statement
	query := db.Explain(statement.SQL.String(), statement.Vars...)

	sqldb, err := db.DB()
	if err != nil {
		if s.options.ErrorLogger != nil {
			s.options.ErrorLogger(fmt.Sprintf("failed to access DB object: %v", err))
		}
		return
	}

	rows, err := sqldb.Query(explainFormat[db.Name()] + query)
	if err != nil {
		if s.options.ErrorLogger != nil {
			s.options.ErrorLogger(fmt.Sprintf("failed run the EXPLAIN query: %v", err))
		}
		return
	}

	explainResult := []string{}

	for rows.Next() {
		var result string
		rows.Scan(&result)
		explainResult = append(explainResult, result)
	}

	results := strings.Join(explainResult, "\n")
	if strings.Contains(results, tableScanString[db.Name()]) {
		s.action(statement.SQL.String(), results)
	}
}
