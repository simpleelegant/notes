package resources

import (
	"io"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

// OpenDatabase must be called before any other models' operations
func OpenDatabase(dbFile string) error {
	var err error
	db, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return err
	}

	return initArticleCollection()
}

// Export exports all data to w
func Export(w io.Writer) error {
	return db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}
