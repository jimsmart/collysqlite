package collysqlite

import (
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	createVisitDDL = `
		CREATE TABLE IF NOT EXISTS visit (
			id				INTEGER NOT NULL UNIQUE,
			created_at		DATETIME NOT NULL,
			PRIMARY KEY (id)
		);
		CREATE INDEX IF NOT EXISTS idx_visit_created_at ON visit(created_at);
	`
	dropVisitDDL = `
		DROP INDEX IF EXISTS idx_visit_created_at;
		DROP TABLE IF EXISTS visit;
	`
)

type visitRecord struct {
	ID        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type visitTracker interface {
	// Init initializes the visitTracker.
	Init() error
	Destroy() error
	// Visited receives and stores a request ID that is visited by the Collector.
	Visited(requestID uint64) error
	// IsVisited returns true if the request was visited before IsVisited
	// is called.
	IsVisited(requestID uint64) (bool, error)
}

// TODO(js) Do we also need a remove/delete/unvisit? At least to help when testing?

// TODO I think this interface would have a cleaner cut if it was based around 'url string' instead of 'requestID uint64'.

var _ visitTracker = &VisitTracker{}

type VisitTracker struct {
	Path string
}

func NewVisitTracker(path string) *VisitTracker {
	t := &VisitTracker{
		Path: path + ".sqlite",
	}
	return t
}

func (t *VisitTracker) Init() error {
	err := ensurePathExists(t.Path)
	if err != nil {
		return err
	}
	db, err := t.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(createVisitDDL)
	return err
}

func (t *VisitTracker) Destroy() error {
	db, err := t.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(dropVisitDDL)
	if err != nil {
		return err
	}
	return removeIfNoTables(db, t.Path)
}

func (t *VisitTracker) Visited(requestID uint64) error {
	// TODO(js) The API would be better without an opaque ID going across the boundary,
	// it should just be a url string, and the tracker implementation should decide how best to store it.
	db, err := t.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	r := &visitRecord{
		ID:        requestID,
		CreatedAt: time.Now(),
	}
	_, err = db.NamedExec("INSERT INTO visit (id, created_at) VALUES (:id, :created_at)", r)
	return err
}

func (t *VisitTracker) IsVisited(requestID uint64) (bool, error) {
	// TODO(js) The API would be better without an opaque ID going across the boundary,
	// it should just be a url string, and the tracker implementation should decide how best to store it.
	db, err := t.connect()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var count int
	err = db.Get(&count, "SELECT COUNT(id) FROM visit WHERE id = ?", requestID)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

func (t *VisitTracker) connect() (*sqlx.DB, error) {
	return sqlx.Connect("sqlite3", t.Path)
}
