package sequence

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/ajstarks/svgo"
	"github.com/simpleelegant/notes/diagram/svg_util"
)

// Style type of diagram's style
type Style struct {
	PFill, PStroke             string // participants' style, colors of fill / stroke
	LifelineStroke             string // lifeline stroke color
	ClosureFill, ClosureStroke string // closure's fill & stroke color
	MessageStroke              string // message stroke color
	SequenceNumberFill         string // color of auto sequence number filling
}

// Diagram type of sequence diagram
type Diagram struct {
	title              string
	autoSequenceNumber bool
	participants       []*participant
	closures           []*closure
	messages           []*message
	notes              []string

	padding                  int // diagram's and title's padding
	titleHeight              int
	noteHeight               int
	pWidth, pHeight, pMargin int // participants' width / height / margin
	cMargin                  int // closure's margin & padding
	mMargin                  int // messages' margin

	usedX, usedY   int      // indicates how many (x,y) in used
	pY             int      // y of participants' start point
	lY, lY2        int      // y of lifelines' start point and end point, respectively
	currentClosure *closure // the most inner opened closure
}

// New return an Diagram instance
func New(padding, pWidth, pHeight, pMargin, cMargin, mMargin int) *Diagram {
	d := &Diagram{
		padding:     padding,
		titleHeight: 24,
		noteHeight:  16,
		pWidth:      pWidth,
		pHeight:     pHeight,
		pMargin:     pMargin,
		cMargin:     cMargin,
		mMargin:     mMargin,

		usedX: padding,
		usedY: padding,
	}

	d.usedY += d.titleHeight // update usedY for title
	d.pY = d.usedY + d.padding
	d.lY = d.pY + d.pHeight
	d.usedY = d.lY // update usedY for participants

	return d
}

func (d *Diagram) addParticipant(label string) *participant {
	// if participant is already existed
	for _, p := range d.participants {
		if p.label == label {
			return p
		}
	}

	p := &participant{label: label, y: d.pY}

	if len(d.participants) != 0 {
		p.x = d.usedX + d.pMargin
	} else {
		p.x = d.usedX
	}

	p.llx = p.x + d.pWidth/2
	d.usedX = p.x + d.pWidth // update usedX for participants
	d.participants = append(d.participants, p)

	return p
}

func (d *Diagram) startClosure(t closureType, condition string) {
	c := &closure{
		typ:       t,
		condition: condition,

		y:      d.usedY + d.cMargin,
		parent: d.currentClosure,
	}
	d.closures = append(d.closures, c)
	d.currentClosure = c

	// update usedX for closure top
	// assuming the height of closure label & condition is 20
	d.usedY = c.y + 20
}

func (d *Diagram) endClosure() error {
	c := d.currentClosure
	if c == nil {
		return errors.New("close before start closure")
	}

	// if no message or descendant in closure, remove it.
	// IMPORTANT, reverse effects of startClosure()
	{
		var hasDescendants bool
		for _, a := range d.closures {
			if a.parent == c {
				hasDescendants = true
				break
			}
		}
		if c.lastMessage == nil && !hasDescendants {
			d.usedY -= 20 + d.cMargin
			d.currentClosure = c.parent
			d.closures = d.closures[0 : len(d.closures)-1]

			return nil
		}
	}

	// calculate c.x
	{
		x := math.MaxInt32
		if c.lp != nil {
			x = c.lp.llx
		}

		// check its first generation descendants
		for _, a := range d.closures {
			if a.parent == c && a.x < x {
				x = a.x
			}
		}
		c.x = x - d.cMargin
	}

	// calculate c.w
	{
		var x2 int
		if c.rp != nil {
			x2 = c.rp.llx
		}

		if c.rSelfCall {
			x2 += 60 // assuming self-called message's width is 60
		}

		// check its first generation descendants
		for _, a := range d.closures {
			if a.parent == c && a.x+a.w > x2 {
				x2 = a.x + a.w
			}
		}
		c.w = x2 + d.cMargin - c.x
	}

	// calculate c.h
	c.h = d.usedY + d.cMargin - c.y
	if c.lastMessage != nil && c.lastMessage.selfCall {
		c.h += 20 // assuming self-called message's height is 20
	}

	c.closed = true
	d.usedY = c.y + c.h         // update usedX for closure bottom
	d.currentClosure = c.parent // change currentClosure

	return nil
}

func (d *Diagram) addMessage(from, to, label string, call bool) {
	f := d.addParticipant(from)
	t := d.addParticipant(to)

	m := &message{
		from:     f,
		to:       t,
		label:    label,
		call:     call,
		x1:       f.llx,
		y1:       d.usedY + d.mMargin,
		x2:       t.llx,
		selfCall: from == to,
	}

	d.messages = append(d.messages, m)

	// update usedY for message
	d.usedY = m.y1
	if m.selfCall {
		d.usedY += 20
	}

	if d.currentClosure != nil {
		d.currentClosure.addMessage(m)
	}
}

func (d *Diagram) addNote(note string) {
	d.notes = append(d.notes, note)
}

// Parse parse s to fill diagram
func (d *Diagram) Parse(b []byte) error {
	if err := d.parse(b); err != nil {
		return err
	}

	// validation
	if len(d.closures) > 0 && !d.closures[0].closed {
		return errors.New("closures must be closed")
	}

	d.lY2 = d.usedY + d.mMargin

	// reserve space for bottom participants
	d.usedY = d.lY2 + d.pHeight

	// reserve space for notes
	if len(d.notes) > 0 {
		// assuming top margin of notes is 40, line spacing is 4
		d.usedY += 40 + len(d.notes)*(d.noteHeight+4)
	}

	// set diagram final size
	d.usedX += d.padding
	d.usedY += d.padding

	return nil
}

// Draw draw diagram into w
func (d *Diagram) Draw(w io.Writer, s *Style) {
	c := svg.New(w)
	c.Start(d.usedX, d.usedY)

	// draw title
	c.Text(d.usedX/2, d.pY/2+d.padding, d.title,
		fmt.Sprintf("text-anchor:middle;fill:black;font-size:%dpx", d.titleHeight))

	// draw lifelines
	c.Gstyle(fmt.Sprintf("stroke-width:0.5px;stroke:%s", s.LifelineStroke))
	for _, p := range d.participants {
		c.Line(p.llx, d.lY, p.llx, d.lY2)
	}
	c.Gend()

	// draw participants
	d.drawParticipants(c, s)

	// draw closures
	d.drawClosures(c, s)

	// draw messages
	d.drawMessages(c, s)

	// draw notes
	if len(d.notes) > 0 {
		c.Textlines(d.padding, d.lY2+d.pHeight+40, d.notes,
			d.noteHeight, d.noteHeight+4, "black", "start")
	}

	c.End()
}

func (d *Diagram) drawParticipants(c *svg.SVG, s *Style) {
	id := "participants"

	c.Gid(id)
	{
		// draw participant boxes
		c.Gstyle(fmt.Sprintf("fill:%s;stroke:%s", s.PFill, s.PStroke))
		for _, p := range d.participants {
			c.Roundrect(p.x, p.y, d.pWidth, d.pHeight, 4, 4)
		}
		c.Gend()

		// draw participant labels
		c.Gstyle("text-anchor:middle;fill:black")
		for _, p := range d.participants {
			c.Text(p.x+d.pWidth/2, p.y+d.pHeight/2+6, p.label)
		}
		c.Gend()
	}
	c.Gend()

	// draw participants at bottom
	// note: c.Use() use relative (x,y)
	c.Use(0, d.lY2-d.pY, "#"+id)
}

func (d *Diagram) drawClosures(c *svg.SVG, s *Style) {
	c.Gstyle(fmt.Sprintf("fill:none;stroke:%s", s.ClosureStroke))
	for _, p := range d.closures {
		c.Roundrect(p.x, p.y, p.w, p.h, 4, 4)
	}
	c.Gend()

	c.Gstyle(fmt.Sprintf("fill:%s;stroke:%s", s.ClosureFill, s.ClosureStroke))
	for _, p := range d.closures {
		c.Roundrect(p.x, p.y, 50, 20, 0, 0)
	}
	c.Gend()

	c.Gstyle("text-anchor:middle;fill:black;font-size:0.8em")
	var title string
	for _, p := range d.closures {
		switch p.typ {
		case loopClosure:
			title = "loop"
		case altClosure:
			title = "alt"
		default:
			// no title
		}
		c.Text(p.x+25, p.y+14, title)
	}
	c.Gend()

	c.Gstyle("text-anchor:middle;fill:black;font-size:0.9em")
	for _, p := range d.closures {
		c.Text(p.x+p.w/2, p.y+14, fmt.Sprintf("[ %s ]", p.condition))
	}
	c.Gend()
}

func (d *Diagram) drawMessages(c *svg.SVG, s *Style) {
	// define arrowhead
	c.Def()
	svg_util.Marker(c, "arrowhead", 5, 2, 6, 4, "auto", "fill:"+s.MessageStroke)
	c.Path("M 0,0 V 4 L6,2 Z")
	c.MarkerEnd()
	c.DefEnd()

	// draw message lines
	c.Gstyle(fmt.Sprintf(
		"marker-end:url(#arrowhead);fill:none;stroke-width:1.5px;stroke:%s", s.MessageStroke))
	{
		var style string
		for _, m := range d.messages {
			if m.call {
				style = ""
			} else {
				style = "stroke-dasharray:3,3"
			}
			if m.selfCall {
				c.Path(fmt.Sprintf("M %d,%d C %d,%d %d,%d %d,%d",
					m.x1, m.y1, m.x1+60, m.y1-10, m.x1+60, m.y1+30, m.x1, m.y1+20), style)
			} else {
				c.Line(m.x1, m.y1, m.x2, m.y1, style)
			}
		}
	}
	c.Gend()

	// draw message labels
	c.Gstyle("text-anchor:middle;font-size:0.9em;fill:black")
	{
		var (
			x   int
			pre string
		)

		for i, m := range d.messages {
			if m.selfCall {
				x = m.x1 + 20
			} else {
				x = m.x1/2 + m.x2/2
			}

			if d.autoSequenceNumber {
				pre = fmt.Sprintf(`<tspan style="fill:%s">%d.</tspan>`, s.SequenceNumberFill, i+1)
			} else {
				pre = ""
			}

			svg_util.Text(c, x, m.y1-4, pre, m.label)
		}
	}
	c.Gend()
}
