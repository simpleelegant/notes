package automata

import (
	"fmt"
	"io"
	"math"
	"math/rand"

	"github.com/ajstarks/svgo"
	"github.com/simpleelegant/notes/diagram/svg_util"
)

// Style type of diagram's style
type Style struct {
	LinkStroke string // link's stroke color
}

type state struct {
	label string
	final bool
	x, y  int // convention: (-1,-1) means haven't position
}

type link struct {
	from, to *state
	label    string
}

// Diagram type of automata diagram
type Diagram struct {
	title  string
	states []*state
	links  []*link
	notes  []string

	padding     int
	titleHeight int
	noteHeight  int
	sr, sm      int // state radius / margin
}

// New return an Diagram instance
func New(padding, stateMargin int) *Diagram {
	return &Diagram{
		padding:     padding,
		titleHeight: 24,
		noteHeight:  16,
		sr:          16,
		sm:          stateMargin,
	}
}

func (d *Diagram) addNote(note string) {
	d.notes = append(d.notes, note)
}

func (d *Diagram) addState(label string, final bool) {
	for _, s := range d.states {
		if s.label == label {
			if !s.final {
				s.final = final
			}

			return
		}
	}

	d.states = append(d.states, &state{label: label, final: final, x: -1, y: -1})
}

func (d *Diagram) getState(label string) *state {
	for _, s := range d.states {
		if s.label == label {
			return s
		}
	}

	return nil
}

func (d *Diagram) addLink(from, to, label string) {
	if from == "" {
		t := d.getState(to)
		if t.x == -1 {
			d.locate(t, 0, 0)
			d.links = append(d.links, &link{from: nil, to: t, label: label})
		}

		return
	}

	// they are promised that are not nil
	f := d.getState(from)
	t := d.getState(to)

	if f.x == -1 {
		if t.x == -1 {
			d.locate(f, 0, 0)
			d.locate(t, f.x, f.y)
		} else {
			d.locate(f, t.x, t.y)
		}
	} else if t.x == -1 {
		d.locate(t, f.x, f.y)
	}

	d.links = append(d.links, &link{from: f, to: t, label: label})
}

// how do I get location:
/*
  ++e+c++
  +8+++5+
  f++3++9
  ++401++
  i++2++a
  +7+++6+
  ++d+b++
*/
func (d *Diagram) locate(s *state, nearPointX, nearPointY int) {
	var tryCount, tryX, tryY int

OuterLoop:
	for {
		tryCount++

		switch tryCount {
		case 1:
			tryX, tryY = nearPointX+d.sm, nearPointY
		case 2:
			tryX, tryY = nearPointX, nearPointY+d.sm
		case 3:
			tryX, tryY = nearPointX, nearPointY-d.sm
		case 4:
			tryX, tryY = nearPointX-d.sm, nearPointY
		case 5:
			m := d.sm * 2
			tryX, tryY = nearPointX+m, nearPointY-m
		case 6:
			m := d.sm * 2
			tryX, tryY = nearPointX+m, nearPointY+m
		case 7:
			m := d.sm * 2
			tryX, tryY = nearPointX-m, nearPointY+m
		case 8:
			m := d.sm * 2
			tryX, tryY = nearPointX-m, nearPointY-m
		case 9:
			tryX, tryY = nearPointX+d.sm*3, nearPointY-d.sm
		case 10:
			tryX, tryY = nearPointX+d.sm*3, nearPointY+d.sm
		case 11:
			tryX, tryY = nearPointX+d.sm, nearPointY+d.sm*3
		case 12:
			tryX, tryY = nearPointX+d.sm, nearPointY-d.sm*3
		case 13:
			tryX, tryY = nearPointX-d.sm, nearPointY+d.sm*3
		case 14:
			tryX, tryY = nearPointX-d.sm, nearPointY-d.sm*3
		case 15:
			tryX, tryY = nearPointX-d.sm*3, nearPointY-d.sm
		case 16:
			tryX, tryY = nearPointX-d.sm*3, nearPointY+d.sm
		default:
			// random point on the cicle with r=d.sm*4
			r := d.sm * 4
			tryX = rand.Intn(r*2) - r
			tryY = int(math.Floor(math.Sqrt(float64(r*r - tryX*tryX))))

			// random sign
			if rand.Intn(2) == 0 {
				tryX = -tryX
			}
			if rand.Intn(2) == 0 {
				tryY = -tryY
			}
		}

		for _, t := range d.states {
			if t.x == tryX && t.y == tryY {
				continue OuterLoop
			}
		}

		break
	}

	// location is determined
	s.x, s.y = tryX, tryY
}

func (d *Diagram) calculateStatesArea() (leftX, rightX, topY, bottomY int) {
	for _, s := range d.states {
		if s.x < leftX {
			leftX = s.x
		} else if s.x > rightX {
			rightX = s.x
		}

		if s.y < topY {
			topY = s.y
		} else if s.y > bottomY {
			bottomY = s.y
		}
	}

	s := d.sr

	return leftX - s - d.sm/2, rightX + s, topY - s - d.sm/2, bottomY + s
}

// Draw draw diagram into w
func (d *Diagram) Draw(w io.Writer, s *Style) {
	leftX, rightX, topY, bottomY := d.calculateStatesArea()
	usedX := d.padding + (rightX - leftX) + d.padding
	usedY := d.padding + d.titleHeight + d.padding + (bottomY - topY) + d.padding
	notesStart := usedY + 40 // assuming top margin of notes is 40

	// reserve space for notes
	if len(d.notes) > 0 {
		// assuming line spacing is 4
		usedY = notesStart + len(d.notes)*(d.noteHeight+4)
	}

	c := svg.New(w)
	c.Start(usedX, usedY)

	// draw title
	c.Text(usedX/2, d.titleHeight+d.padding*2, d.title,
		fmt.Sprintf("text-anchor:middle;fill:black;font-size:%dpx", d.titleHeight))

	// draw state labels
	d.drawStates(c, s, -leftX+d.padding, -topY+d.padding+d.titleHeight+d.padding)

	// draw notes
	if len(d.notes) > 0 {
		c.Textlines(d.padding, notesStart, d.notes,
			d.noteHeight, d.noteHeight+4, "black", "start")
	}

	c.End()
}

func (d *Diagram) drawStates(c *svg.SVG, s *Style, x, y int) {
	id := "states"

	c.Def()
	c.Gid(id)
	{
		// draw labels
		c.Gstyle("text-anchor:middle;fill:black")
		for _, t := range d.states {
			c.Text(t.x, t.y+6, t.label)
		}
		c.Gend()

		// draw circles
		c.Gstyle("fill:none;stroke:black")
		for _, t := range d.states {
			c.Circle(t.x, t.y, d.sr)
			if t.final {
				c.Circle(t.x, t.y, d.sr+3)
			}
		}
		c.Gend()

		// draw links
		d.drawLinks(c, s)
	}
	c.Gend()
	c.DefEnd()

	c.Use(x, y, "#"+id)
}

func (d *Diagram) drawLinks(c *svg.SVG, s *Style) {
	// define arrowhead
	c.Def()
	svg_util.Marker(c, "arrowhead", 5, 2, 6, 4, "auto", "fill:"+s.LinkStroke)
	c.Path("M 0,0 V 4 L6,2 Z")
	c.MarkerEnd()
	c.DefEnd()

	// draw links
	c.Gstyle(fmt.Sprintf(
		"marker-end:url(#arrowhead);fill:none;stroke-width:1.5px;stroke:%s", s.LinkStroke))
	for _, k := range d.links {
		if k.from == nil {
			x1 := d.sm / 2
			x4 := int(math.Sqrt(float64(d.sr * d.sr / 2)))
			third := (x4 - x1) / 3
			x2, x3 := x1+third, x4-third
			c.Path(fmt.Sprintf("M %d,%d C %d,%d %d,%d %d,%d",
				k.to.x-x1, k.to.y-x1,
				k.to.x-x2, k.to.y-x2-x1/3,
				k.to.x-x3, k.to.y-x3-x1/3,
				k.to.x-x4, k.to.y-x4))
		} else if d.areNeighbour(k.from, k.to) {
			dx, dy := k.to.x-k.from.x, k.to.y-k.from.y
			ratio := float64(d.sr) / math.Sqrt(float64(dx*dx+dy*dy))
			sx, sy := int(ratio*float64(dx)), int(ratio*float64(dy))
			c.Line(k.from.x+sx, k.from.y+sy, k.to.x-sx, k.to.y-sy)
		} else {
			// draw curve to avoid cross over states
			// FIXME
			dx, dy := k.to.x-k.from.x, k.to.y-k.from.y
			ratio := float64(d.sr) / math.Sqrt(float64(dx*dx+dy*dy))
			sx, sy := int(ratio*float64(dx)), int(ratio*float64(dy))
			c.Line(k.from.x+sx, k.from.y+sy, k.to.x-sx, k.to.y-sy)
		}
	}
	c.Gend()

	// draw link labels
	c.Gstyle("text-anchor:middle;font-size:0.9em;fill:black")
	{
		var x, y int
		for _, k := range d.links {
			if k.from == nil {
				x = d.sm / 4
				c.Text(k.to.x-x-4, k.to.y-x, k.label)

				continue
			}

			x, y = (k.from.x+k.to.x)/2, (k.from.y+k.to.y)/2

			if k.from.x == k.to.x {
				x += 6
			} else {
				y -= 6
			}

			c.Text(x, y, k.label)
		}
	}
	c.Gend()
}

func (d *Diagram) areNeighbour(s, t *state) bool {
	x := t.x - s.x
	if x < 0 {
		x = -x
	}
	if x == d.sm {
		return true
	}

	x = t.y - s.y
	if x < 0 {
		x = -x
	}
	if x == d.sm {
		return true
	}

	return false
}
