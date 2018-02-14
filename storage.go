package collysqlite

import (
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// storage is a local copy of colly.Storage.
type storage interface {
	// Init initializes the storage
	Init() error
	// Visited receives and stores a request ID that is visited by the Collector
	Visited(requestID uint64) error
	// IsVisited returns true if the request was visited before IsVisited
	// is called
	IsVisited(requestID uint64) (bool, error)
	// GetCookieJar returns with cookie jar that implements the
	// http.CookieJar interface
	GetCookieJar() http.CookieJar
}

var _ storage = &Storage{}

type Storage struct {
	Path string
	*VisitTracker
	*ExplodingCookieJar
	*Cache
}

// TODO(js) Is Store a better name for Storage?

func NewStorage(path string) *Storage {
	s := &Storage{
		Path:               path,
		VisitTracker:       NewVisitTracker(path + "-visits"),
		ExplodingCookieJar: &ExplodingCookieJar{Jar: NewCookieJar(path + "-cookies")},
		Cache:              NewCache(path + "-cache"),
	}
	return s
}

func (s *Storage) Init() error {
	err := s.VisitTracker.Init()
	if err != nil {
		return err
	}
	err = s.ExplodingCookieJar.Init()
	if err != nil {
		return err
	}
	return s.Cache.Init()
}

func (s *Storage) Destroy() error {
	err1 := s.VisitTracker.Destroy()
	err2 := s.ExplodingCookieJar.Destroy()
	err3 := s.Cache.Destroy()
	// TODO(js) Is there some type of existing multi-error? If not, implement one.
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	return nil
}

func (s *Storage) GetCookieJar() http.CookieJar {
	return s
}
