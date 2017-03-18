package sequence

type message struct {
	from  *participant
	to    *participant
	label string
	call  bool // message type, true means "call", false means "return"

	x1, y1, x2 int  // start & end points, y2 is ommited
	selfCall   bool // true when source is equal to target
}
