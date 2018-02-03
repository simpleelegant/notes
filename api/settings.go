package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/simpleelegant/notes/resources"
)

// Restore restore data
func Restore(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		f, _, err := r.FormFile("file")
		if err != nil {
			return err
		}
		defer f.Close()

		// write to a temporary file
		tfn := os.TempDir() + "/restore_upload.tmp"
		{
			tf, err := os.Create(tfn)
			if err != nil {
				return err
			}
			defer os.Remove(tfn)
			defer tf.Close()

			if _, err := io.Copy(tf, f); err != nil {
				return err
			}
			tf.Close()
		}

		// open file by boltdb
		db, err := bolt.Open(tfn, 0600, nil)
		if err != nil {
			return err
		}

		// checking
		if err := resources.CheckArticleCollection(db); err != nil {
			return err
		}

		// really restore
		return resources.RestoreArticlesFrom(db)
	}()
	if err != nil {
		replyInfo(w, err)
		return
	}

	replyInfo(w, "Restored success.")
}

// Export export data
func Export(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="notes.%s.db"`,
			time.Now().Format("2006-01-02.15_04_05.000Z")))
	w.WriteHeader(http.StatusOK)
	if err := resources.Export(w); err != nil {
		log.Println(err)
	}
}
