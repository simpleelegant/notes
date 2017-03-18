package models

import "github.com/boltdb/bolt"

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
