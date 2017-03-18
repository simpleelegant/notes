package automata

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	// DiagramType ...
	DiagramType = "automataDiagram"
	title       = `Title:`
	note        = `Note:`
)

var (
	reLink           = regexp.MustCompile(`^([^-]*)-(.*)->(.*)$`)
	reFinalState     = regexp.MustCompile(`^\((.+)\)$`)
	errSyntaxInvalid = errors.New("syntax invalid")
)

// Parse parse b to fill diagram
func (d *Diagram) Parse(b []byte) error {
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

		if strings.HasPrefix(l, title) {
			// title
			d.title = l[6:]
			continue
		}
		if strings.HasPrefix(l, note) {
			// note
			d.addNote(l[5:])
			continue
		}
		if err := d.parseLinkLine(l); err != nil {
			return fmt.Errorf("line %d: %s", lineNum, err)
		}
	}

	return nil
}

func (d *Diagram) parseLinkLine(s string) error {
	t := strings.Split(s, "->")
	l := len(t)
	to := t[l-1]
	if l < 2 || to == "" {
		return errSyntaxInvalid
	}

	if m := reFinalState.FindStringSubmatch(to); len(m) > 0 {
		to = m[1]
		d.addState(to, true)
	} else {
		d.addState(to, false)
	}

	links := [][]string{}
	for i := l - 2; i > -1; i-- {
		if to == "" {
			return errSyntaxInvalid
		}

		n := reLink.FindStringSubmatch(fmt.Sprintf("%s->%s", t[i], to))
		if len(n) == 0 {
			return errSyntaxInvalid
		}

		to = n[1]
		if to != "" {
			if m := reFinalState.FindStringSubmatch(to); len(m) > 0 {
				to = m[1]
				d.addState(to, true)
			} else {
				d.addState(to, false)
			}
		}

		links = append(links, []string{to, n[2], n[3]})
	}

	// add links to diagram
	for i := len(links) - 1; i > -1; i-- {
		d.addLink(links[i][0], links[i][2], links[i][1])
	}

	return nil
}
