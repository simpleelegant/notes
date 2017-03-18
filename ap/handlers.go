package api

import (
	"bytes"
	cMD5 "crypto/md5"
	"fmt"
	"net/http"

	"github.com/simpleelegant/notes/diagram/automata"
	"github.com/simpleelegant/notes/diagram/sequence"
)

// RenderDiagram render a diagram in svg format
func RenderDiagram(w http.ResponseWriter, r *http.Request) {
	s := bytes.TrimSpace([]byte(r.FormValue("source")))
	firstAt := bytes.IndexByte(s, '\n')
	if firstAt == -1 {
		firstAt = len(s)
	}
	first := bytes.TrimSpace(s[0:firstAt])
	var out bytes.Buffer

	switch string(first) {
	case sequence.DiagramType:
		dia := sequence.New(8, 150, 60, 50, 10, 30)
		if err := dia.Parse(s); err != nil {
			replyBadRequest(w, err)
			return
		}

		dia.Draw(&out, &sequence.Style{
			PStroke:            "#CCCCFF",
			PFill:              "#ECECFF",
			LifelineStroke:     "grey",
			ClosureStroke:      "#FFCCCC",
			ClosureFill:        "#FFECEC",
			MessageStroke:      "black",
			SequenceNumberFill: "#FF0000",
		})
	case automata.DiagramType:
		dia := automata.New(10, 100)
		if err := dia.Parse(s); err != nil {
			replyBadRequest(w, err)
			return
		}

		dia.Draw(&out, &automata.Style{LinkStroke: "#A5A8FF"})
	default:
		replyBadRequest(w, fmt.Errorf(`"%s" not supported`, first))
		return
	}

	reply(w, http.StatusOK, map[string]string{"svg": out.String()})
}

// MD5 calculate md5 digest
func MD5(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, map[string]string{
		"md5": fmt.Sprintf("%x", cMD5.Sum([]byte(r.FormValue("data")))),
	})
}
