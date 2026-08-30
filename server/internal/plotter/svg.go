package plotter

import (
	"fmt"
	"io"
	"math"
	"time"

	"astrohop/internal/astronomy/coordinates"

	svg "github.com/ajstarks/svgo"
)

type SvgPlotter struct {
}

const labelPad = 22

func (p *SvgPlotter) PlotAzimuth(out io.Writer, o Options, data *ChartData) {
	canvas := svg.New(out)
	canvas.Start(o.Size.Width, o.Size.Height)

	plotFrame(canvas, o)

	content := contentBounds(o)
	gridArea, sidebarArea := splitLayout(content, o.Style.SidebarGap)

	plotTitle(canvas, o, gridArea, data.Title)
	plotGrid(canvas, o, gridArea)
	plotObjects(canvas, o, gridArea, data)
	plotSidebar(canvas, o, sidebarArea, data)

	canvas.End()
}

func plotFrame(canvas *svg.SVG, o Options) {
	canvas.Rect(0, 0, o.Size.Width, o.Size.Height, o.Style.BackgroundStyle)

	off := o.Style.BorderOffset
	borderStyle := fmt.Sprintf("fill:none;stroke:%s;stroke-width:%d", o.Style.BorderColor, o.Style.BorderThickness)
	canvas.Rect(off, off, o.Size.Width-2*off, o.Size.Height-2*off, borderStyle)
}

func plotTitle(canvas *svg.SVG, o Options, gridArea rect, title string) {
	if title == "" {
		return
	}
	pad := o.Style.PanelPadding
	canvas.Text(gridArea.X0+pad, gridArea.Y0+pad+24, title, o.Style.TitleStyle)
}

func plotGrid(canvas *svg.SVG, o Options, area rect) {
	s := o.Style

	radius := area.height()/2 - labelPad
	centerX := area.centerX()
	centerY := area.centerY()
	labelR := radius + labelPad

	for alt := 0; alt <= 90; alt += s.AltitudeStep {
		r := altToRadius(radius, alt)
		style := s.GridStyle
		if alt == 0 {
			style = s.HorizonStyle
		}
		canvas.Circle(centerX, centerY, r, style)
	}

	for az := 0; az < 360; az += s.AzimuthStep {
		x2, y2 := polar(centerX, centerY, radius, float64(az))
		style := s.GridStyle
		if az%90 == 0 {
			style = s.AxisStyle
		}
		canvas.Line(centerX, centerY, x2, y2, style)
	}

	cardinal := map[int]string{0: "N", 90: "E", 180: "S", 270: "W"}
	for az := 0; az < 360; az += 90 {
		x, y := polar(centerX, centerY, labelR, float64(az))
		canvas.Text(x, y+6, cardinal[az], s.DirectionTextStyle)
	}

	for az := 0; az < 360; az += s.LabelStep {
		if az%90 == 0 {
			continue
		}
		x, y := polar(centerX, centerY, labelR, float64(az))
		canvas.Text(x, y+4, fmt.Sprintf("%d°", az), s.DegreeTextStyle)
	}

	for alt := s.AltitudeStep; alt < 90; alt += s.AltitudeStep {
		r := altToRadius(radius, alt)
		canvas.Text(centerX+8, centerY-r-4, fmt.Sprintf("%d°", alt), s.AltitudeTextStyle)
	}

	canvas.Circle(centerX, centerY, 2, "fill:#000000;stroke:none")
}

func plotObjects(canvas *svg.SVG, o Options, area rect, data *ChartData) {
	if len(data.Objects) == 0 {
		return
	}
	s := o.Style

	radius := area.height()/2 - labelPad
	centerX := area.centerX()
	centerY := area.centerY()

	type point struct{ x, y int }
	points := make([]point, len(data.Objects))

	for i, obj := range data.Objects {
		x, y := polar(centerX, centerY, altToRadius(radius, int(obj.Coords.Alt)), obj.Coords.Az)
		points[i] = point{x: x, y: y}
	}

	for i := range data.Objects {
		from, to := points[i], points[data.Tour.Order[i]]
		canvas.Line(from.x, from.y, to.x, to.y, s.PathStyle)

		mx, my := (from.x+to.x)/2, (from.y+to.y)/2
		lx, ly := offsetAwayFromCenter(centerX, centerY, from.x, from.y, to.x, to.y, mx, my, 12)
		canvas.Text(lx, ly, fmt.Sprintf("%.1f°", data.Tour.Distances[i][data.Tour.Order[i]]), s.DistanceTextStyle)
	}

	for i, obj := range data.Objects {
		p := points[i]
		xs, ys := starPoints(p.x, p.y, s.StarSize, s.StarSize/2)
		canvas.Polygon(xs, ys, s.StarStyle)
		canvas.Text(p.x, p.y-s.StarSize-4, obj.Label, s.ObjectLabelStyle)
	}
}

// offsetAwayFromCenter сдвигает точку (mx, my) на отрезке from-to на amount
// пикселей перпендикулярно отрезку, в сторону от центра (cx, cy) - чтобы
// подпись расстояния не перекрывала линию пути.
func offsetAwayFromCenter(cx, cy, fromX, fromY, toX, toY, mx, my, amount int) (int, int) {
	dx, dy := float64(toX-fromX), float64(toY-fromY)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return mx, my
	}
	nx, ny := -dy/length, dx/length
	if nx*float64(mx-cx)+ny*float64(my-cy) < 0 {
		nx, ny = -nx, -ny
	}
	return mx + int(nx*float64(amount)), my + int(ny*float64(amount))
}

func plotSidebar(canvas *svg.SVG, o Options, area rect, data *ChartData) {
	y := area.Y0

	y = plotTextPanel(canvas, o, area, y, "Place of Observation",
		[]string{formatLocation(data.Location)})

	y = plotTextPanel(canvas, o, area, y, "Date and Time",
		[]string{formatTime(data.Time)})

	y = plotTextPanel(canvas, o, area, y, "Conditions",
		formatConditions(data.Conditions))

	y = plotTipsPanel(canvas, o, area, y, "Angular Distance Tips",
		data.AngularTips, data.ArmNote)

	y = plotLegendPanel(canvas, o, area, y, "Legend", data.Legend)
}

func plotTextPanel(canvas *svg.SVG, o Options, area rect, y int, header string, lines []string) int {
	s := o.Style

	y += s.PanelPadding + s.LineHeight
	canvas.Text(area.X0, y, header, s.HeaderStyle)
	y += s.HeaderGap

	for _, line := range lines {
		y += s.LineHeight
		canvas.Text(area.X0, y, line, s.BodyStyle)
	}

	y += s.PanelPadding
	canvas.Line(area.X0, y, area.X1, y, s.DividerStyle)
	return y + s.PanelGap
}

func plotTipsPanel(canvas *svg.SVG, o Options, area rect, y int, header string, tips []string, note string) int {
	s := o.Style

	y += s.PanelPadding + s.LineHeight
	canvas.Text(area.X0, y, header, s.HeaderStyle)
	y += s.HeaderGap

	for _, tip := range tips {
		y += s.LineHeight
		canvas.Text(area.X0, y, tip, s.BodyStyle)
	}

	if note != "" {
		y += s.LineHeight
		canvas.Text(area.X0, y, note, s.NoteStyle)
	}

	y += s.PanelPadding
	canvas.Line(area.X0, y, area.X1, y, s.DividerStyle)
	return y + s.PanelGap
}

func plotLegendPanel(canvas *svg.SVG, o Options, area rect, y int, header string, items []LegendItem) int {
	s := o.Style

	y += s.PanelPadding + s.LineHeight
	canvas.Text(area.X0, y, header, s.HeaderStyle)
	y += s.HeaderGap

	for _, item := range items {
		y += s.LineHeight
		drawLegendMarker(canvas, o, area.X0, y, item.Marker)
		canvas.Text(area.X0+s.LegendSwatchSize+s.LegendGap, y, item.Label, s.BodyStyle)
	}

	y += s.PanelPadding
	canvas.Line(area.X0, y, area.X1, y, s.DividerStyle)
	return y + s.PanelGap
}

func drawLegendMarker(canvas *svg.SVG, o Options, x, baselineY int, marker LegendMarker) {
	size := o.Style.LegendSwatchSize
	cx := x + size/2
	cy := baselineY - size/2

	switch marker {
	case MarkerStar:
		xs, ys := starPoints(cx, cy, size/2, size/5)
		canvas.Polygon(xs, ys, "fill:#000000;stroke:none")

	case MarkerPath:
		canvas.Line(x, cy, x+size, cy, "stroke:#000000;stroke-width:2;stroke-dasharray:3,2")

	case MarkerMilkyWay:
		canvas.Rect(x, cy-size/4, size, size/2, "fill:#d9d9d9;stroke:none")

	case MarkerConstellation:
		canvas.Line(x, cy+size/4, x+size/2, cy-size/4, "stroke:#000000;stroke-width:1.5")
		canvas.Line(x+size/2, cy-size/4, x+size, cy+size/6, "stroke:#000000;stroke-width:1.5")
		canvas.Circle(x, cy+size/4, 2, "fill:#000000")
		canvas.Circle(x+size/2, cy-size/4, 2, "fill:#000000")
		canvas.Circle(x+size, cy+size/6, 2, "fill:#000000")
	}
}

type rect struct {
	X0, Y0, X1, Y1 int
}

func (r rect) width() int   { return r.X1 - r.X0 }
func (r rect) height() int  { return r.Y1 - r.Y0 }
func (r rect) centerX() int { return (r.X0 + r.X1) / 2 }
func (r rect) centerY() int { return (r.Y0 + r.Y1) / 2 }

func contentBounds(o Options) rect {
	m := o.Style.Margins
	return rect{X0: m, Y0: m, X1: o.Size.Width - m, Y1: o.Size.Height - m}
}

func splitLayout(content rect, gap int) (grid rect, sidebar rect) {
	size := content.height()
	grid = rect{X0: content.X0, Y0: content.Y0, X1: content.X0 + size, Y1: content.Y1}
	sidebar = rect{X0: grid.X1 + gap, Y0: content.Y0, X1: content.X1, Y1: content.Y1}
	return
}

func polar(cx, cy, r int, azimuthDeg float64) (int, int) {
	rad := azimuthDeg * math.Pi / 180
	x := float64(cx) + float64(r)*math.Sin(rad)
	y := float64(cy) - float64(r)*math.Cos(rad)
	return int(math.Round(x)), int(math.Round(y))
}

func altToRadius(maxRadius int, altDeg int) int {
	return maxRadius * (90 - altDeg) / 90
}

func starPoints(cx, cy, outerR, innerR int) ([]int, []int) {
	const spikes = 5
	xs := make([]int, spikes*2)
	ys := make([]int, spikes*2)
	for i := 0; i < spikes*2; i++ {
		r := outerR
		if i%2 == 1 {
			r = innerR
		}
		angle := math.Pi/2 + float64(i)*math.Pi/spikes
		xs[i] = cx + int(float64(r)*math.Cos(angle))
		ys[i] = cy - int(float64(r)*math.Sin(angle))
	}
	return xs, ys
}

func formatLocation(loc coordinates.GeoLocation) string {
	latHemi, lat := "N", loc.Lat
	if lat < 0 {
		latHemi, lat = "S", -lat
	}
	lonHemi, lon := "E", loc.Long
	if lon < 0 {
		lonHemi, lon = "W", -lon
	}
	return fmt.Sprintf("%.2f° %s, %.2f° %s", lat, latHemi, lon, lonHemi)
}

func formatTime(t time.Time) string {
	_, offsetSec := t.Zone()
	offsetHours := offsetSec / 3600
	return fmt.Sprintf("%s (UTC +%d)", t.Format("2 January 2006 15:04"), offsetHours)
}

func formatConditions(c ChartConditions) []string {
	return []string{
		fmt.Sprintf("Bortle %d  %s", c.Bortle, c.Equipment),
		fmt.Sprintf("Max Magnitude %+.1f", c.LimitingMagnitude),
	}
}
