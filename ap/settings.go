package api

import (
	"io"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/simpleelegant/notes/conf"
	"github.com/simpleelegant/notes/models"
)

// Settings resource
type Settings struct{}

// Get get settings & system information
func (*Settings) Get(w http.ResponseWriter, r *http.Request) {
	i := conf.GatherInfo()
	info := map[string]interface{}{
		"server started at":            i.StartAt.String(),
		"server listening at":          i.ServerAddress,
		"server local ip":              i.ComputerLocalIP,
		"recent restored data at":      i.RecentRestoredDataAt,
		"recent restored data version": "unsupported",
	}
	if i.ErrorInMemory != "" {
		info["error in memory"] = i.ErrorInMemory
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
		replyInfo(w, r, err.Error())
		return
	}
	defer f.Close()

	// write to a temporary file
	tfn := os.TempDir() + "/restore_upload.tmp"
	{
		tf, err := os.Create(tfn)
		if err != nil {
			replyInfo(w, r, err.Error())
			return
		}
		defer os.Remove(tfn)
		defer tf.Close()

		if _, err := io.Copy(tf, f); err != nil {
			replyInfo(w, r, err.Error())
			return
		}
		tf.Close()
	}

	// open file by boltdb
	db, err := bolt.Open(tfn, 0600, nil)
	if err != nil {
		replyInfo(w, r, err.Error())
		return
	}

	// checking
	a := (*models.Article)(nil)
	if err := a.CheckCollection(db); err != nil {
		replyInfo(w, r, err.Error())
		return
	}

	// really restore
	if err := a.Restore(db); err != nil {
		replyInfo(w, r, err.Error())
		return
	}

	// record this operation
	conf.FreshRecentRestoredDataAt()

	replyInfo(w, r, "Successfully restored.")
}
