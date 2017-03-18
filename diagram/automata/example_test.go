package automata_test

import (
	"fmt"
	"os"

	"github.com/simpleelegant/notes/diagram/automata"
)

func Example() {
	source := `
automataDiagram
    Title: Demo
	Note:this is a note.

	-ε->1-i->2-f->(3)
	1-ε->4-a-z->5

    Note: this is a note.
    Note: this is another note.
	`

	dia := automata.New(10, 100)

	// parse
	if err := dia.Parse([]byte(source)); err != nil {
		fmt.Println(err)
		return
	}

	dia.Draw(os.Stdout, &automata.Style{LinkStroke: "#A5A8FF"})
}
