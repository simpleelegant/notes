package models

import (
	"io"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

// Init must be called before any other models' operations
func Init(dbFile string) error {
	// open database
	var err error
	db, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return err
	}

	return (*Article)(nil).initCollection()
}

func WriteTo(w io.Writer) error {
	return db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)

		return err
	})
}
