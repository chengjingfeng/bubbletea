package tea

import "errors"

type MouseMsg MouseEvent

// MouseEvent represents a mouse event, which could be a click, a scroll wheel
// movement, a cursor movement, or a combination.
type MouseEvent struct {
	X    int
	Y    int
	Type MouseEventType
	Alt  bool
	Ctrl bool
}

// String returns a string representation of a mouse event.
func (m MouseEvent) String() (s string) {
	if m.Ctrl {
		s += "ctrl+"
	}
	if m.Alt {
		s += "alt+"
	}
	s += mouseEventTypes[m.Type]
	return s
}

// MouseEventType indicates the type of mouse event occurring.
type MouseEventType int

const (
	MouseUnknown MouseEventType = iota
	MouseLeft
	MouseRight
	MouseMiddle
	MouseRelease
	MouseWheelUp
	MouseWheelDown
	MouseMotion
)

var mouseEventTypes = map[MouseEventType]string{
	MouseUnknown:   "unknown",
	MouseLeft:      "left",
	MouseRight:     "right",
	MouseMiddle:    "middle",
	MouseRelease:   "release",
	MouseWheelUp:   "wheel up",
	MouseWheelDown: "wheel down",
	MouseMotion:    "motion",
}

// Parse an X10-encoded mouse event; the simplest kind. The last release of
// X10 was December 1986, by the way.
//
// X10 mouse events look like:
//
//     ESC [M Cb Cx Cy
//
// See: http://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
func parseX10MouseEvent(buf []byte) (m MouseEvent, err error) {
	if len(buf) != 6 || string(buf[:3]) != "\x1b[M" {
		return m, errors.New("not an X10 mouse event")
	}

	const byteOffset = 32

	e := buf[3] - byteOffset

	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000

		bitsMask = 0b0000_0011

		bitsLeft    = 0b0000_0000
		bitsMiddle  = 0b0000_0001
		bitsRight   = 0b0000_0010
		bitsRelease = 0b0000_0011

		bitsWheelUp   = 0b0000_0000
		bitsWheelDown = 0b0000_0001
	)

	if e&bitWheel != 0 {
		// Check the low two bits.
		switch e & bitsMask {
		case bitsWheelUp:
			m.Type = MouseWheelUp
		case bitsWheelDown:
			m.Type = MouseWheelDown
		}
	} else {
		// Check the low two bits.
		// We do not separate clicking and dragging.
		switch e & bitsMask {
		case bitsLeft:
			m.Type = MouseLeft
		case bitsMiddle:
			m.Type = MouseMiddle
		case bitsRight:
			m.Type = MouseRight
		case bitsRelease:
			if e&bitMotion != 0 {
				m.Type = MouseMotion
			} else {
				m.Type = MouseRelease
			}
		}
	}

	if e&bitAlt != 0 {
		m.Alt = true
	}
	if e&bitCtrl != 0 {
		m.Ctrl = true
	}

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = int(buf[4]) - byteOffset - 1
	m.Y = int(buf[5]) - byteOffset - 1

	return m, nil
}
