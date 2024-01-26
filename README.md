# GORM SCAN ALERT

This is a library intended to automatically detect sequential scans on our database,
and give you a way to be warned about them.

This currently only supports Postgres as the dialect.
(The limitation is pretty much in the scan detection, e.g. does this explain contains "Seq Scan"?)

## Usage

Once you have your *gorm.DB object, you can add the gormalert scanner as a plugin to it.
To make it easier, there's a helper method for it:

```go

...
scanalertv2.RegisterScanAlert(db, scanalertv2.DefaultAlertOptions(), func(query, explain string) {
    // Tell someone about the sequential scan!
    fmt.Printf("The query %q just did a sequential scan!\n", query)
})
```

The `db` object will be instrumented with the alert after that.
I don't recommend running this in production, but instead on your tests and staging environments.


### Options

```go
type AlertOptions struct {
	Name        string
	Async       bool
	QueryType   queryType
	IncludeRaw  bool
	ErrorLogger func(string)
}
```

You can instrument any of the query types:

CreateQuery
UpdateQuery
SelectQuery
DeleteQuery

You can also instrument them all, but you will need to a separate call to `RegisterScanAlert` with each one of them.
Note that the name in the AlertOptions must be unique.

The `Async` option will run the callback within a Goroutine, which should not block the main flow.

Note that by default these will only apply to explicit gorm helpers (e.g. db.Create, db.Update, db.First, db.Find, ...)
To enable it on Raw() queries, the option `IncludeRaw` must be set. But note that, if set, it will run for ALL raw queries, not only the one
mentioned in `QueryType`.
