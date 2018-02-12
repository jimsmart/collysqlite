package collysqlite

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// TODO Consider URL query string normalisation?

const (
	createDDL = `
		CREATE TABLE IF NOT EXISTS cache (
			url				TEXT NOT NULL UNIQUE,
			data			BLOB,
			created_at		DATETIME NOT NULL,
			PRIMARY KEY (url)
		);
		CREATE INDEX IF NOT EXISTS idx_cache_created_at ON cache(created_at);
	`
	dropDDL = `
		DROP INDEX IF EXISTS idx_created_at;
		DROP TABLE IF EXISTS resource;
	`
)

type cacheItem struct {
	URL       string    `db:"url"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}

// cache is the proposed interface for pluggable cache implementations in Colly.
type cache interface {
	Init() error
	Get(url string) ([]byte, error)
	Put(url string, data []byte) error
	Remove(url string) error
}

var _ cache = &Cache{}

type Cache struct {
	Path string
}

func NewCache(path string) *Cache {
	return &Cache{
		Path: path + ".sqlite",
	}
}

func (c *Cache) Init() error {
	i := strings.LastIndexByte(c.Path, '/')
	if i >= 0 {
		dir := c.Path[:i]
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	db, err := c.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(createDDL)
	return err
}

func (c *Cache) Destroy() error {
	db, err := c.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(dropDDL)
	if err != nil {
		return err
	}
	// TODO(js) We leave behind a folder, which, if empty, maybe should remove?
	// But what if we are several folders deep? Should we then recursively remove folders?
	// This is why I'm happy to leave it, document it, and just let the caller deal with it.
	return os.Remove(c.Path)
}

func (c *Cache) Get(url string) ([]byte, error) {
	db, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var b []byte
	err = db.Get(&b, "SELECT data FROM cache WHERE url = ?", url)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (c *Cache) Put(url string, data []byte) error {
	db, err := c.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	item := &cacheItem{
		URL:       url,
		Data:      data,
		CreatedAt: time.Now(),
	}
	_, err = db.NamedExec("INSERT INTO cache (url, data, created_at) VALUES (:url, :data, :created_at)", item)
	return err
}

func (c *Cache) Remove(url string) error {
	db, err := c.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM cache WHERE url = ?", url)
	return err
}

func (c *Cache) connect() (*sqlx.DB, error) {
	return sqlx.Connect("sqlite3", c.Path)
}
