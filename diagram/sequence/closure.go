package sequence

type closureType int

const (
	loopClosure closureType = iota
	altClosure
)

type closure struct {
	typ       closureType
	condition string

	x, y, w, h int // box's (x,y), width, height
	parent     *closure
	closed     bool

	lp, rp      *participant // the most left participant, and the most right one
	rSelfCall   bool         // true means a self-called message at the most right participant
	lastMessage *message
}

func (c *closure) addMessage(m *message) {
	c.lastMessage = m

	lp, rp := m.from, m.to
	if lp.x > rp.x {
		lp, rp = rp, lp
	}

	if c.lp == nil {
		// assuming this is first message to c
		c.lp, c.rp, c.rSelfCall = lp, rp, m.selfCall
		return
	}

	if lp.x < c.lp.x {
		c.lp = lp
	} else if rp.x > c.rp.x {
		c.rp = rp
		c.rSelfCall = m.selfCall
	} else if rp == c.rp && !c.rSelfCall {
		c.rSelfCall = m.selfCall
	}
}
