package svg_util

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ajstarks/svgo"
)

// Marker variant of SVG.Marker() with "orient" attribute support
func Marker(svg *svg.SVG, id string, x, y, width, height int, orient string, s ...string) {
	fmt.Fprintf(svg.Writer,
		`<marker id="%s" orient="%s" refX="%d" refY="%d" markerWidth="%d" markerHeight="%d" %s`,
		id, orient, x, y, width, height, endstyle(s, ">\n"))
}

// Text variant of SVG.Text() with a prefix (should no escape for it) support
func Text(svg *svg.SVG, x int, y int, prefix, t string, s ...string) {
	fmt.Fprintf(svg.Writer, `<text %s %s`, loc(x, y), endstyle(s, ">"))
	fmt.Fprint(svg.Writer, prefix)
	xml.Escape(svg.Writer, []byte(t))
	fmt.Fprintln(svg.Writer, `</text>`)

}

// endstyle modifies an SVG object, with either a series of name="value" pairs,
// or a single string containing a style
//
// this function is copied from "github.com/ajstarks/svgo"
func endstyle(s []string, endtag string) string {
	if len(s) > 0 {
		nv := ""
		for i := 0; i < len(s); i++ {
			if strings.Index(s[i], "=") > 0 {
				nv += (s[i]) + " "
			} else {
				nv += style(s[i])
			}
		}
		return nv + endtag
	}
	return endtag

}

// style returns a style name,attribute string
//
// this function is copied from "github.com/ajstarks/svgo"
func style(s string) string {
	if len(s) > 0 {
		return fmt.Sprintf(`style="%s"`, s)
	}
	return s
}

// loc returns the x and y coordinate attributes
//
// this function is copied from "github.com/ajstarks/svgo"
func loc(x int, y int) string { return fmt.Sprintf(`x="%d" y="%d"`, x, y) }
