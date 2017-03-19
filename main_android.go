// the logic in this file follows a rule:
//   let HTTP API can show graphic issues if any,
//   and let screen can other issues.

package main

import (
	"fmt"
	"image"
	"image/draw"
	"net/http"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/simpleelegant/notes/api"
	"github.com/simpleelegant/notes/conf"
	"github.com/simpleelegant/notes/mobile"
	"github.com/simpleelegant/notes/models"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	mobile_font "golang.org/x/mobile/exp/font"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

const screenText = `
notes - a portable wiki system

author: Wang Yujian <simpleelegant@163.com>


%s
`

type logic struct {
	a          app.App
	glctx      gl.Context
	size       size.Event
	fontFace   font.Face
	screenText string

	// below fields must be set manually
	fontSize int
}

func (c *logic) SetApp(a app.App) {
	c.a = a
}

func (c *logic) OnCreate(e lifecycle.Event) {
	go c.startHTTPServer()

	// get system font
	f, err := freetype.ParseFont(mobile_font.Default())
	if err != nil {
		api.Debug = err
		return
	}
	c.fontFace = truetype.NewFace(f, &truetype.Options{Size: float64(c.fontSize)})
}

func (c *logic) OnDestroy(e lifecycle.Event) {}

func (c *logic) BecomeVisible(e lifecycle.Event) {
	if c.glctx == nil {
		var ok bool
		c.glctx, ok = e.DrawContext.(gl.Context)
		if !ok {
			api.Debug = "fails on e.DrawContext.(gl.Context)"
			return
		}
	}
}

func (c *logic) BecomeInvisible(e lifecycle.Event) {}

func (c *logic) GainFocus(e lifecycle.Event) {}
func (c *logic) LoseFocus(e lifecycle.Event) {}

func (c *logic) OnSize(e size.Event) {
	c.size = e
}

func (c *logic) OnPaint(e paint.Event) {
	c.drawText(c.screenText)
}

func (c *logic) OnTouch(e touch.Event) {}

func (c *logic) OnOtherEvent(e interface{}) {}

func (c *logic) startHTTPServer() {
	// init models
	if err := models.Init(conf.GetDataFolder()); err != nil {
		c.setScreenText(err.Error())
		return
	}

	registerRoutes(&assetsHandler{modTime: time.Now()})

	addr := conf.GetHTTPAddress()

	tmpl := `%s
serving at http://%s/
Open this url by a web browser to use it.`
	c.setScreenText(fmt.Sprintf(tmpl,
		conf.StartedAt.Format("2006-01-02 15:04:05 -0700 MST"),
		addr))

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		c.setScreenText(err.Error())
	}
}

func (c *logic) drawText(s string) {
	// new RGBA image
	images := glutil.NewImages(c.glctx)
	img := images.NewImage(c.size.WidthPx, c.size.HeightPx)

	// draw background
	draw.Draw(img.RGBA, img.RGBA.Bounds(), image.White, image.ZP, draw.Src)

	// draw the text to the image
	if c.fontFace == nil {
		return
	}
	d := &font.Drawer{Dst: img.RGBA, Src: image.Black, Face: c.fontFace}
	dy := 10 + c.fontSize
	y := dy
	for _, l := range strings.Split(s, "\n") {
		d.Dot = fixed.P(10, y)
		d.DrawString(l)
		y += dy
	}

	// draw the image to glctx
	c.glctx.ClearColor(1, 1, 1, 1)
	c.glctx.Clear(gl.COLOR_BUFFER_BIT)
	img.Upload()
	img.Draw(
		c.size,
		geom.Point{0, 0},
		geom.Point{c.size.WidthPt, 0},
		geom.Point{0, c.size.HeightPt},
		img.RGBA.Bounds(),
	)
}

func (c *logic) setScreenText(s string) {
	c.screenText = fmt.Sprintf(screenText, s)
	c.a.Send(paint.Event{})
}

type assetsHandler struct {
	modTime time.Time
}

func (h *assetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a := strings.TrimPrefix(r.URL.Path, "/assets/")
	if a == "" {
		a = "index.html"
	}

	f, err := asset.Open(a)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()

	http.ServeContent(w, r, a, h.modTime, f)
}

func main() {
	a := &logic{fontSize: 40}

	// config
	//conf.Host = "0.0.0.0"
	conf.Host = "127.0.0.1"
	conf.Port = 9030
	// there assumes its Android package-name is "yujian.notes"
	err := conf.SetDataFolder("/data/data/yujian.notes/files")
	if err != nil {
		a.setScreenText(err.Error())
	}

	mobile.Run(a)
}
