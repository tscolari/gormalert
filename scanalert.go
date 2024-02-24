package gormalert

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

type QueryType string

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
	ErrorLogger func(args ...any)
}

// DefaultAlertOptions returns an AlertOptions object that is tailored
// for testing usage.
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
		ErrorLogger: log.Print,
	}
}

type actionFunc func(sourceQuery string, scanResult string)

// RegisterScanAlert registers a plugin in the gorm.DB argument that
// will detect sequential scans using the given options.
// The action function will be called every time a sequential scan is detected.
func RegisterScanAlert(db *gorm.DB, options AlertOptions, action actionFunc) error {
	alerter := NewScanAlerterPlugin(options, action)
	return db.Use(alerter)
}

// NewScanAlerterPlugin returns a configured gorm plugin that will
// call `action` every time a sequential scan is detected using the given options.
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

// Name returns the name of the plugin for gorm.
func (s *scanAlerter) Name() string {
	return s.options.Name
}

// Initialize will, based on the options, create callbacks
// to each query type. Those callbacks will trigger every time
// those particular query types run.
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

		if err := processor.Register(s.Name()+"_"+string(queryType), scanFunc); err != nil {
			return err
		}
	}

	return nil
}

// AsyncScan will run `Scan` as a goroutine, without blocking the caller.
func (s *scanAlerter) AsyncScan(db *gorm.DB) {
	go s.Scan(db)
}

// Scan will extract the query that was just executed and prepend it with `EXPLAIN`.
// It will use the result of the explain query to check if a sequential scan was or
// not executed.
// When a scan is detected, it will call the configured action with the
// query string and the explain result.
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
		if err := rows.Scan(&result); err != nil {
			s.options.ErrorLogger("failed to scan explain results:", err)
		}
		explainResult = append(explainResult, result)
	}

	results := strings.Join(explainResult, "\n")
	if strings.Contains(results, tableScanString[db.Name()]) {
		s.action(statement.SQL.String(), results)
	}
}
