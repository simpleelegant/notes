package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func reply(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(b)
}

func replyBadRequest(w http.ResponseWriter, err error) {
	reply(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
}

func formValue(r *http.Request, key string) (string, bool) {
	if vs := r.Form[key]; len(vs) > 0 {
		return vs[0], true
	}

	return "", false
}

var tmpl = `<!DOCTYPE html>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>info - notes</title>
    </head>

    <body>
	<p>
		<a href="%s">Back</a>
	</p>
	<p style="text-align: center;">%s</p>
    </body>
</html>
`

// reply a simple information page
func replyInfo(w http.ResponseWriter, r *http.Request, info interface{}) {
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(tmpl, ref, info)))
}
