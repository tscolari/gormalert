# GORM SCAN ALERT

This is a library intended to automatically detect sequential scans on our database,
and give you a way to be warned about them.

This currently only supports Postgres as the dialect.
(The limitation is pretty much in the scan detection, e.g. Does this explain contains "Seq Scan"?)

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
	ErrorLogger func(string)
}
```

You can instrument any of the query types:

CreateQuery
UpdateQuery
SelectQuery
DeleteQuery
RawQuery

You can also instrument them all, but you will need to a separate call to `RegisterScanAlert` with each one of them.
Note that the name in the AlertOptions must be unique.

The scans are attached to callbacks on gorm for the respective methods (e.g. `CreateQuery` will trigger only for `db.Create(...)`). The `RawQuery` will
trigger for calls of `Exec(...)` or `Raw(...)`, no mater which kind of query those methods are performing.

The `Async` option will run the callback within a Goroutine, which should not block the main flow.
