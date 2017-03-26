package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/simpleelegant/notes/conf"
	"github.com/simpleelegant/notes/models"
)

// Debug store internal error
var Debug interface{}

// Settings resource
type Settings struct{}

// Get get settings & system information
func (*Settings) Get(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"started at":   conf.StartedAt.Format("2006-01-02 15:04:05 -0700 MST"),
		"serving at":   conf.GetHTTPAddress(),
		"data version": "(unsupported now)",
	}

	ips, err := conf.GetComputerLocalIP()
	if err != nil {
		info["local IP"] = err.Error()
	} else {
		info["local IP"] = strings.Join(ips, ", ")
	}

	t, err := conf.GetLastRestoringTimestamp()
	if err != nil {
		info["last data restoring"] = err.Error()
	} else {
		info["last data restoring"] = t
	}

	if Debug != nil {
		info["debug"] = Debug
	}

	reply(w, http.StatusOK, map[string]interface{}{
		"info": info,
	})
}

// Restore restore data
func (*Settings) Restore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		replyInfo(w, r, err)
		return
	}
	defer f.Close()

	// write to a temporary file
	tfn := os.TempDir() + "/restore_upload.tmp"
	{
		tf, err := os.Create(tfn)
		if err != nil {
			replyInfo(w, r, err)
			return
		}
		defer os.Remove(tfn)
		defer tf.Close()

		if _, err := io.Copy(tf, f); err != nil {
			replyInfo(w, r, err)
			return
		}
		tf.Close()
	}

	// open file by boltdb
	db, err := bolt.Open(tfn, 0600, nil)
	if err != nil {
		replyInfo(w, r, err)
		return
	}

	// checking
	a := (*models.Article)(nil)
	if err := a.CheckCollection(db); err != nil {
		replyInfo(w, r, err)
		return
	}

	// really restore
	if err := a.Restore(db); err != nil {
		replyInfo(w, r, err)
		return
	}

	// record this operation
	if err := conf.SetLastRestoringTimestamp(); err != nil {
		replyInfo(w, r, err)
		return
	}

	replyInfo(w, r, "Successfully restored.")
}

// Export export data
func (*Settings) Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	fn := fmt.Sprintf("notes.%s.db", time.Now().Format("2006-01-02.15-04-05.-0700.MST"))

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fn))
	w.WriteHeader(http.StatusOK)
	if err := models.WriteTo(w); err != nil {
		fmt.Println(err)
	}
}
