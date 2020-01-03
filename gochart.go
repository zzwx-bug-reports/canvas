package canvas

import (
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type Output int

const (
	OutputSVG Output = iota
	OutputPDF
	OutputEPS
	OutputPNG
	OutputJPG
	OutputGIF
)

type GoChart struct {
	c            *Canvas
	ctx          *Context
	height       float64
	output       Output
	dpi          float64
	font         *FontFamily
	fontSize     float64
	fontColor    drawing.Color
	textRotation float64
}

func NewGoChart(output Output) func(int, int) (chart.Renderer, error) {
	return func(w, h int) (chart.Renderer, error) {
		font := NewFontFamily("font")
		font.LoadLocalFont("Arimo", FontRegular)

		c := New(float64(w), float64(h))
		return &GoChart{
			c:      c,
			ctx:    NewContext(c),
			height: float64(h),
			output: output,
			dpi:    72.0,
			font:   font,
		}, nil
	}
}

func (r *GoChart) ResetStyle() {
	r.ctx.ResetStyle()
	r.textRotation = 0.0
}

func (r *GoChart) GetDPI() float64 {
	return r.dpi
}

func (r *GoChart) SetDPI(dpi float64) {
	r.dpi = dpi
}

func (r *GoChart) SetClassName(name string) {
	// TODO
}

func (r *GoChart) SetStrokeColor(col drawing.Color) {
	r.ctx.SetStrokeColor(col)
}

func (r *GoChart) SetFillColor(col drawing.Color) {
	r.ctx.SetFillColor(col)
}

func (r *GoChart) SetStrokeWidth(width float64) {
	r.ctx.SetStrokeWidth(width)
}

func (r *GoChart) SetStrokeDashArray(dashArray []float64) {
	r.ctx.SetDashes(0.0, dashArray...)
}

func (r *GoChart) MoveTo(x, y int) {
	r.ctx.MoveTo(float64(x), r.height-float64(y))
}

func (r *GoChart) LineTo(x, y int) {
	r.ctx.LineTo(float64(x), r.height-float64(y))
}

func (r *GoChart) QuadCurveTo(cx, cy, x, y int) {
	r.ctx.QuadTo(float64(cx), r.height-float64(cy), float64(x), r.height-float64(y))
}

func (r *GoChart) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) {
	startAngle *= 180.0 / math.Pi
	delta *= 180.0 / math.Pi

	start := ellipsePos(rx, -ry, 0.0, float64(cx), r.height-float64(cy), startAngle)
	if r.c.Empty() {
		r.ctx.MoveTo(start.X, r.height-start.Y)
	} else {
		r.ctx.LineTo(start.X, r.height-start.Y)
	}
	r.ctx.Arc(rx, ry, 0.0, startAngle, startAngle+delta)
}

func (r *GoChart) Close() {
	r.ctx.ClosePath()
	r.ctx.MoveTo(0.0, 0.0)
}

func (r *GoChart) Stroke() {
	r.ctx.Stroke()
}

func (r *GoChart) Fill() {
	r.ctx.Fill()
}

func (r *GoChart) FillStroke() {
	r.ctx.FillStroke()
}

func (r *GoChart) Circle(radius float64, x, y int) {
	r.ctx.DrawPath(float64(x), r.height-float64(y), Circle(radius))
}

func (r *GoChart) SetFont(font *truetype.Font) {
	// TODO
}

func (r *GoChart) SetFontColor(col drawing.Color) {
	r.fontColor = col
}

func (r *GoChart) SetFontSize(size float64) {
	r.fontSize = size
}

func (r *GoChart) Text(body string, x, y int) {
	face := r.font.Face(r.fontSize*ptPerMm*r.dpi/72.0, r.fontColor, FontRegular, FontNormal)
	r.ctx.Push()
	r.ctx.SetFillColor(r.fontColor)
	r.ctx.ComposeView(Identity.Rotate(-r.textRotation * 180.0 / math.Pi))
	r.ctx.DrawText(float64(x), r.height-float64(y), NewTextLine(face, body, Left))
	r.ctx.Pop()
}

func (r *GoChart) MeasureText(body string) chart.Box {
	p, _ := r.font.Face(r.fontSize*ptPerMm*r.dpi/72.0, r.fontColor, FontRegular, FontNormal).ToPath(body)
	bounds := p.Bounds()
	bounds = bounds.Transform(Identity.Rotate(-r.textRotation * 180.0 / math.Pi))
	return chart.Box{Left: int(bounds.X + 0.5), Top: int(bounds.Y + 0.5), Right: int((bounds.W + bounds.X) + 0.5), Bottom: int((bounds.H + bounds.Y) + 0.5)}
}

func (r *GoChart) SetTextRotation(radian float64) {
	r.textRotation = radian
}

func (r *GoChart) ClearTextRotation() {
	r.textRotation = 0.0
}

func (r *GoChart) Save(w io.Writer) error {
	width, height := r.c.Size()
	switch r.output {
	case OutputSVG:
		svg := SVG(w, width, height)
		r.c.Render(svg)
		return svg.Close()
	case OutputPDF:
		pdf := PDF(w, width, height)
		r.c.Render(pdf)
		return pdf.Close()
	case OutputEPS:
		eps := EPS(w, width, height)
		r.c.Render(eps)
		return nil
	case OutputPNG:
		img := r.c.WriteImage(1.0)
		if err := png.Encode(w, img); err != nil {
			return err
		}
		return nil
	case OutputJPG:
		img := r.c.WriteImage(1.0)
		if err := jpeg.Encode(w, img, nil); err != nil {
			return err
		}
		return nil
	case OutputGIF:
		img := r.c.WriteImage(1.0)
		if err := gif.Encode(w, img, nil); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown output format")
}