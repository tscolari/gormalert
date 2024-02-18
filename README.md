# GORM SCAN ALERT

This is a library intended to automatically detect sequential scans on our database,
and give you a way to be warned about them.

This currently only supports Postgres and Mysql as the dialect.
(This is not battle tested on Mysql, apart from the tests on this repo)

## Usage

Once you have your *gorm.DB object, you can add the gormalert scanner as a plugin to it.
To make it easier, there's a helper method for it:

```go

...
gormalert.RegisterScanAlert(db, gormalert.DefaultAlertOptions(), func(query, explain string) {
    // Tell someone about the sequential scan!
    fmt.Printf("The query %q just did a sequential scan!\n", query)
})
```

The `db` object will be instrumented with the alert after that.
I don't recommend running this in production, but instead on your tests and staging environments.

### Example

This is a helper that create DB objects for tests, and this code adds a clause to automatically fail
any test that performs a table scan using that object.

```go
func testDB(t *testing.T) (*gorm.DB, func()) {
	db, closer := dbtest.DB(t, ...)

	gormalert.RegisterScanAlert(db, gormalert.DefaultAlertOptions(), func(source string, result string) {
		t.Errorf("the query %q executed a sequential scan: %s", source, result)
	})

	return db, closer
}
```

Alternatively the `ExpectDBWithoutSequentialScan` helper can be used to achieve similar result:

```go
func TestSomething(t testing.T) {
	// setup
	db := ...
	ExpectDBWithoutSequentialScan(t, db)

	// tests
	...
}
```

### Options

```go
type AlertOptions struct {
	Name        string
	Async       bool
	QueryType   []QueryType
	ErrorLogger func(string)
}
```

You can instrument any of the query types:

* CreateQuery
* UpdateQuery
* SelectQuery
* DeleteQuery
* RawQuery

The scans are attached to callbacks on gorm for the respective methods (e.g. `CreateQuery` will trigger only for `db.Create(...)`). The `RawQuery` will
trigger for calls of `Exec(...)` or `Raw(...)`, no mater which kind of query those methods are performing.

The `Async` option will run the callback within a Goroutine, which should not block the main flow.
