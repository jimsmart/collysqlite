package collysqlite

import (
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

//

func ensurePathExists(path string) error {
	i := strings.LastIndexByte(path, '/')
	if i >= 0 {
		dir := path[:i]
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeIfNoTables(db *sqlx.DB, path string) error {
	count := 0
	err := db.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table'")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// TODO(js) We leave behind a folder, which, if empty, maybe should remove?
	// But what if we are several folders deep? Should we then recursively remove folders?
	// This is why I'm happy to leave it, document it, and just let the caller deal with it.
	return os.Remove(path)
}
