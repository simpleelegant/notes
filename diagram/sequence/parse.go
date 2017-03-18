package sequence

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	// DiagramType ...
	DiagramType = `sequenceDiagram`

	closureEnd         = `end`
	autoSequenceNumber = `autoSequenceNumber`
	title              = `Title:`
	note               = `Note:`
	reMessage          = regexp.MustCompile(`^(.*[^-]+)(-->>|->>)(.+):(.*)$`)
	reClosureStart     = regexp.MustCompile(`^(loop|alt) (.+)$`)
)

func (d *Diagram) parse(b []byte) error {
	lineNum := 0
	blank := true
	buf := bytes.NewBuffer(b)

	for {
		lineNum++
		l, err := buf.ReadString('\n')
		switch err {
		case nil:
			// ignore
		case io.EOF:
			if l == "" {
				return nil // finished
			}
		default:
			return err
		}

		l = strings.TrimSpace(l)

		// ignore empty line
		if l == "" {
			continue
		}

		// first non-blank line must be DiagramType
		if blank {
			if l != DiagramType {
				return errors.New("diagram type is unkown")
			}
			blank = false
			continue
		}

		if l == autoSequenceNumber {
			d.autoSequenceNumber = true
		} else if l == closureEnd {
			// closure end
			if err := d.endClosure(); err != nil {
				return err
			}
		} else if strings.HasPrefix(l, title) {
			// title
			d.title = l[6:]
		} else if strings.HasPrefix(l, note) {
			// note
			d.addNote(l[5:])
		} else if m := reMessage.FindStringSubmatch(l); m != nil {
			// message
			d.addMessage(strings.TrimSpace(m[1]), strings.TrimSpace(m[3]), m[4], m[2] == "->>")
		} else if m := reClosureStart.FindStringSubmatch(l); m != nil {
			// closure start
			t := loopClosure
			if m[1] == "alt" {
				t = altClosure
			}
			d.startClosure(t, m[2])
		} else {
			return fmt.Errorf("line %d: unable to parse", lineNum)
		}
	}

	return nil
}
