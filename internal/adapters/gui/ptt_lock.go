//go:build gui

// Tracks the tap window for one pointer or trigger capture using monotonic time.
// Canvas and native editors use the same transitions and retain their original target.
package gui

import "time"

const pttTapWindow = 350 * time.Millisecond

type tapDecision int

const (
	tapWait tapDecision = iota
	tapBegin
	tapSubmit
)

type pttLock struct {
	source            string
	pressed, deadline time.Time
	locked            bool
}

func (l *pttLock) press(source string, now time.Time) tapDecision {
	if l.locked {
		*l = pttLock{}
		return tapSubmit
	}
	if l.source != "" {
		if !l.deadline.IsZero() && !now.Before(l.deadline) {
			*l = pttLock{}
			return tapSubmit
		}
		if l.source == source && !l.deadline.IsZero() {
			l.deadline = time.Time{}
			l.locked = true
		}
		return tapWait
	}
	l.source = source
	l.pressed = now
	return tapBegin
}

func (l *pttLock) release(source string, now time.Time) tapDecision {
	if l.source != source || l.locked || !l.deadline.IsZero() {
		return tapWait
	}
	if now.Sub(l.pressed) <= pttTapWindow {
		l.deadline = now.Add(pttTapWindow)
		return tapWait
	}
	*l = pttLock{}
	return tapSubmit
}

func (l *pttLock) expired(now time.Time) bool {
	if l.deadline.IsZero() || now.Before(l.deadline) {
		return false
	}
	*l = pttLock{}
	return true
}

func (g *Game) tapPTT(source string, held bool) {
	l := &g.voice.tap
	now := time.Now()
	if !held {
		if l.release(source, now) == tapSubmit {
			g.endPTT()
		}
		return
	}
	if l.source == "" && g.voice.canvas != nil {
		return
	}
	switch l.press(source, now) {
	case tapBegin:
		g.beginPTT(source)
		if g.voice.canvas == nil {
			*l = pttLock{}
		}
	case tapSubmit:
		g.endPTT()
	}
}
