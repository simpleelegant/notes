package sequence_test

import (
	"fmt"
	"os"

	"github.com/simpleelegant/notes/diagram/sequence"
)

func Example() {
	source := `
sequenceDiagram
    autoSequenceNumber

    Title: Demo

    loop x > 1
        A->>B: hello, world
        B-->>A: version data

        alt y == z
            B->>C: call
            C-->>B: return
        end
    end

    A->>A: self-call

    Note: this is a note.
    Note: this is another note.
	`

	dia := sequence.New(8, 150, 60, 50, 10, 30)

	// parse
	if err := dia.Parse([]byte(source)); err != nil {
		fmt.Println(err)
		return
	}

	dia.Draw(os.Stdout, &sequence.Style{
		PStroke:        "#CCCCFF",
		PFill:          "#ECECFF",
		LifelineStroke: "grey",
		ClosureStroke:  "#FFCCCC",
		ClosureFill:    "#FFECEC",
		MessageStroke:  "black",
	})
}
