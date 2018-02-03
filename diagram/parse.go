package diagram

import (
	"bytes"
	"fmt"

	"github.com/simpleelegant/notes/diagram/automata"
	"github.com/simpleelegant/notes/diagram/sequence"
)

// Parse ...
func Parse(source []byte) ([]byte, error) {
	firstAt := bytes.IndexByte(source, '\n')
	if firstAt == -1 {
		firstAt = len(source)
	}
	first := bytes.TrimSpace(source[0:firstAt])
	var out bytes.Buffer

	switch string(first) {
	case sequence.DiagramType:
		dia := sequence.New(8, 150, 60, 50, 10, 30)
		if err := dia.Parse(source); err != nil {
			return nil, err
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
		if err := dia.Parse(source); err != nil {
			return nil, err
		}
		dia.Draw(&out, &automata.Style{LinkStroke: "#A5A8FF"})
	default:
		return nil, fmt.Errorf(`"%s" not supported`, first)
	}

	return out.Bytes(), nil
}
