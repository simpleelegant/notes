package api

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"github.com/simpleelegant/notes/diagram"
)

// RenderDiagram render a diagram in svg format
func RenderDiagram(r *http.Request) (int, interface{}) {
	out, err := diagram.Parse([]byte(formValue(r, "source")))
	if err != nil {
		return http.StatusBadRequest, err
	}
	return http.StatusOK, map[string]string{"svg": string(out)}
}

// MD5 calculate md5 digest
func MD5(r *http.Request) (int, interface{}) {
	return http.StatusOK, map[string]string{
		"md5": fmt.Sprintf("%x", md5.Sum([]byte(formValue(r, "data")))),
	}
}

func formValue(r *http.Request, key string) string {
	return strings.TrimSpace(r.FormValue(key))
}

func replyInfo(w http.ResponseWriter, info interface{}) {
	const tmpl = `<!DOCTYPE html>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>notes</title>
    </head>

    <body>
		<div style="max-width: 500px;margin: auto;">
			<p><a href="/">Back</a></p>
			<p style="text-align: center;font-size: 1.2em;">%s</p>
		</div>
    </body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, tmpl, info)
}
