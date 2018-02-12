# collysqlite

collysqlite is a [Go](https://golang.org) package providing an experimental work-in-progress prototype of an SQLite cache for [Colly](https://github.com/gocolly/colly).

## Installation
```bash
$ go get github.com/jimsmart/collysqlite
```

```go
import "github.com/jimsmart/collysqlite"
```

### Dependencies

- [go-sqlite3](https://github.com/mattn/go-sqlite3) — Go database driver.
- [sqlx](https://github.com/jmoiron/sqlx) — `database/sql` extensions.
- Standard library.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) if you wish to run the tests.

## Example

```go
import "github.com/jimsmart/collysqlite"

cache := collysqlite.NewCache("./cache")

// TODO SetCache is a proposed method on colly.Collector that is currently unimplemented.
// c := colly.NewCollector()
// c.SetCache(cache)
// ...

```

## Documentation

GoDocs [https://godoc.org/github.com/jimsmart/collysqlite](https://godoc.org/github.com/jimsmart/collysqlite)

## Testing

To run the tests execute `go test` inside the project folder.

For a full coverage report, try:

```bash
$ go test -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## License

Package collysqlite is copyright 2018 by Jim Smart and released under the [MIT License](LICENSE.md)
