// Package mobile wrap golang.org/x/mobile to provide conveniences
package mobile

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
)

// AppWrap type of app.App wrap
type AppWrap interface {
	SetApp(app.App)

	OnCreate(lifecycle.Event)
	OnDestroy(lifecycle.Event)

	BecomeVisible(lifecycle.Event)
	BecomeInvisible(lifecycle.Event)

	GainFocus(lifecycle.Event)
	LoseFocus(lifecycle.Event)

	OnSize(size.Event)

	OnPaint(paint.Event)

	OnTouch(touch.Event)

	OnOtherEvent(interface{})
}

// Run run w, don't invoke more than one time
func Run(w AppWrap) {
	app.Main(func(a app.App) {
		w.SetApp(a)

		for f := range a.Events() {
			switch e := a.Filter(f).(type) {
			case lifecycle.Event:
				if e.From < e.To {
					if e.Crosses(lifecycle.StageAlive) == lifecycle.CrossOn {
						w.OnCreate(e)
					}
					if e.Crosses(lifecycle.StageVisible) == lifecycle.CrossOn {
						w.BecomeVisible(e)
					}
					if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOn {
						w.GainFocus(e)
					}
				} else {
					if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOff {
						w.LoseFocus(e)
					}
					if e.Crosses(lifecycle.StageVisible) == lifecycle.CrossOff {
						w.BecomeInvisible(e)
					}
					if e.Crosses(lifecycle.StageAlive) == lifecycle.CrossOff {
						w.OnDestroy(e)
					}
				}
			case size.Event:
				w.OnSize(e)
			case paint.Event:
				w.OnPaint(e)
				a.Publish()
			case touch.Event:
				w.OnTouch(e)
			default:
				w.OnOtherEvent(a.Filter(f))
			}
		}
	})
}
