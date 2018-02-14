package collysqlite

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	createCookieJarDDL = `
		CREATE TABLE IF NOT EXISTS cookie_jar (
			host			TEXT NOT NULL UNIQUE,
			cookies			TEXT NOT NULL,
			modified_at		DATETIME,
			created_at		DATETIME NOT NULL,
			PRIMARY KEY (host)
		);
		CREATE INDEX IF NOT EXISTS idx_cookie_jar_created_at ON cookie_jar(created_at);
		CREATE INDEX IF NOT EXISTS idx_cookie_jar_modified_at ON cookie_jar(modified_at);
	`
	dropCookieJarDDL = `
		DROP INDEX IF EXISTS idx_cookie_jar_created_at;
		DROP INDEX IF EXISTS idx_cookie_jar_modified_at;
		DROP TABLE IF EXISTS cookie_jar;
	`
)

// CollyPersistentCookieJar is like http.CookieJar but with returned errors.
type CollyPersistentCookieJar interface {
	Init() error
	Destroy() error
	Cookies(u *url.URL) ([]*http.Cookie, error)
	SetCookies(u *url.URL, cookies []*http.Cookie) error
}

// TODO(js) It would be much better if every implementation of a
// CookieJar/CookieStore didn't need to implement its own serialisation.
// In reality, this would be a cleaner interface to the cookie store:-
//
// type CollyCookieStore interface {
// 	Init() error
// 	Destroy() error
// 	Cookies(host string) (string, error)
// 	SetCookies(host string, cs string) error
// }

var _ CollyPersistentCookieJar = &CookieJar{}

type cookieJarRecord struct {
	Host       string     `db:"host"`
	Cookies    string     `db:"cookies"`
	ModifiedAt *time.Time `db:"modified_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

type CookieJar struct {
	Path string
	mu   sync.RWMutex
}

func NewCookieJar(path string) *CookieJar {
	j := &CookieJar{
		Path: path + ".sqlite",
	}
	return j
}

func (j *CookieJar) Init() error {
	err := ensurePathExists(j.Path)
	if err != nil {
		return err
	}
	db, err := j.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(createCookieJarDDL)
	return err
}

func (j *CookieJar) Destroy() error {
	db, err := j.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	j.mu.Lock()
	defer j.mu.Unlock()
	_, err = db.Exec(dropCookieJarDDL)
	if err != nil {
		return err
	}
	return removeIfNoTables(db, j.Path)
}

func (j *CookieJar) Cookies(u *url.URL) ([]*http.Cookie, error) {
	db, err := j.connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	cookiesStr := ""
	j.mu.RLock()
	err = db.Get(&cookiesStr, "SELECT cookies FROM cookie_jar WHERE host = ?", u.Host)
	j.mu.RUnlock()
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse raw cookies string to []*http.Cookie.
	cookies := unstringify(cookiesStr)

	// Filter.
	now := time.Now()
	cnew := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		// Drop expired cookies.
		if c.RawExpires != "" && c.Expires.Before(now) {
			continue
		}
		// Drop secure cookies if not over https.
		if c.Secure && u.Scheme != "https" {
			continue
		}
		cnew = append(cnew, c)
	}
	return cnew, nil
}

func (j *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) error {
	db, err := j.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	// We need to use a write lock to prevent a race in the db:
	// if two callers set cookies in a very small window of time,
	// it is possible to drop the new cookies from one caller
	// ('last update wins' == best avoided).
	j.mu.Lock()
	defer j.mu.Unlock()

	var r cookieJarRecord
	err = db.Get(&r, "SELECT * FROM cookie_jar WHERE host = ?", u.Host)
	if err == sql.ErrNoRows {
		// Insert new record.
		r.Host = u.Host
		r.Cookies = stringify(cookies)
		r.CreatedAt = time.Now()
		_, err = db.NamedExec("INSERT INTO cookie_jar (host, cookies, created_at) VALUES (:host, :cookies, :created_at)", r)
		return err
	}
	if err != nil {
		return err
	}
	// Update existing record.

	// Merge existing cookies, new cookies have precendence.
	cnew := make([]*http.Cookie, len(cookies))
	copy(cnew, cookies)
	existing := unstringify(r.Cookies)
	for _, c := range existing {
		if !contains(cnew, c.Name) {
			cnew = append(cnew, c)
		}
	}

	now := time.Now()
	r.ModifiedAt = &now
	r.Cookies = stringify(cnew)
	_, err = db.NamedExec("UPDATE cookie_jar SET cookies = :cookies, modified_at = :modified_at WHERE host = :host", r)
	return err
}

func contains(cookies []*http.Cookie, name string) bool {
	for _, c := range cookies {
		if c.Name == name {
			return true
		}
	}
	return false
}

func stringify(cookies []*http.Cookie) string {
	// Stringify cookies.
	cs := make([]string, len(cookies))
	for i, c := range cookies {
		cs[i] = c.String()
	}
	return strings.Join(cs, "\n")
}

func unstringify(s string) []*http.Cookie {
	h := http.Header{}
	for _, c := range strings.Split(s, "\n") {
		h.Add("Set-Cookie", c)
	}
	r := http.Response{Header: h}
	return r.Cookies()
}

func (j *CookieJar) connect() (*sqlx.DB, error) {
	return sqlx.Connect("sqlite3", j.Path)
}

// TODO Eventually remove ExplodingCookieJar, when Colly handles PersistentCookieJars or CookieStore.

var _ http.CookieJar = &ExplodingCookieJar{}

// ExplodingCookieJar is a temporary wrapper around CollyPersistentCookieJar
// to make it look like http.CookieJar. The problem with this is that the
// persistent jar could encounter db errors during its operation, and it has
// no way to return them to the caller.
//
// Because of this, if the wrapped instance of CollyPersistentCookieJar
// encounters any kind of error, ExplodingCookieJar will exit with a
// message to log.Fatal.
//
// If Colly is to work correctly with persistent cookie jars, it will need to
// internally adopt an interface like CollyPersistentCookieJar so it can handle
// the errors - then the ExplodingCookieJar can be removed.
//
// It is easy to adapt an http.CookieJar (which does not return errors) into
// a CollyPersistentCookieJar (which does return errors) - see CollyCookieJarAdapter
// for the implementation - but vice versa ignores error states and will eventually
// explode / end up in an inconsistent system state.
type ExplodingCookieJar struct {
	Jar *CookieJar
}

func (j *ExplodingCookieJar) Init() error {
	return j.Jar.Init()
}

func (j *ExplodingCookieJar) Destroy() error {
	return j.Jar.Destroy()
}

func (j *ExplodingCookieJar) Cookies(u *url.URL) []*http.Cookie {
	// TODO We have no way of returning a db error?
	c, err := j.Jar.Cookies(u)
	if err != nil {
		log.Fatalf("error %s", err)
	}
	return c
}

func (j *ExplodingCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	// TODO We have no way of returning a db error?
	err := j.Jar.SetCookies(u, cookies)
	if err != nil {
		log.Fatalf("error %s", err)
	}
}

var _ CollyPersistentCookieJar = &CollyCookieJarAdapter{}

// CollyCookieJarAdapter wraps an http.CookieJar to look like a CollyPersistentCookieJar by returning nil errors.
type CollyCookieJarAdapter struct {
	Jar http.CookieJar
}

func (j *CollyCookieJarAdapter) Init() error {
	// TODO Check if wrapped Jar is an Initer.
	return nil
}

func (j *CollyCookieJarAdapter) Destroy() error {
	// TODO Check if wrapped Jar is a Destroyer.
	return nil
}

func (j *CollyCookieJarAdapter) Cookies(u *url.URL) ([]*http.Cookie, error) {
	return j.Jar.Cookies(u), nil
}

func (j *CollyCookieJarAdapter) SetCookies(u *url.URL, cookies []*http.Cookie) error {
	j.Jar.SetCookies(u, cookies)
	return nil
}
