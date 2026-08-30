package plotter

import (
	"astrohop/pkg/algo"
	"time"

	"astrohop/internal/astronomy/coordinates"
)

// ===== Лист и вёрстка =====

type Size struct {
	Width  int
	Height int
}

var SizeA4 = Size{
	Width:  1123,
	Height: 794,
}

type Style struct {
	Margins         int
	BorderOffset    int
	BorderThickness int
	BorderColor     string

	SidebarGap   int
	PanelGap     int
	PanelPadding int
	LineHeight   int
	HeaderGap    int

	AzimuthStep  int
	AltitudeStep int
	LabelStep    int

	AxisStyle          string
	GridStyle          string
	HorizonStyle       string
	DegreeTextStyle    string
	AltitudeTextStyle  string
	DirectionTextStyle string
	BackgroundStyle    string

	TitleStyle   string
	HeaderStyle  string
	BodyStyle    string
	NoteStyle    string
	DividerStyle string

	StarSize          int
	StarStyle         string
	PathStyle         string
	DistanceTextStyle string
	ObjectLabelStyle  string

	LegendSwatchSize int
	LegendGap        int
}

type Options struct {
	Size  Size
	Style Style
}

var OptionsDefault = Options{
	Size: SizeA4,
	Style: Style{
		Margins:         40,
		BorderOffset:    20,
		BorderThickness: 4,
		BorderColor:     "#000000",

		SidebarGap:   30,
		PanelGap:     14,
		PanelPadding: 8,
		LineHeight:   22,
		HeaderGap:    6,

		AzimuthStep:  15,
		AltitudeStep: 15,
		LabelStep:    30,

		AxisStyle:          "stroke:#333333;stroke-width:1.5;stroke-opacity:0.8;fill:none",
		GridStyle:          "stroke:#aaaaaa;stroke-width:1;stroke-dasharray:2,3;stroke-opacity:0.8;fill:none",
		HorizonStyle:       "stroke:#000000;stroke-width:2;stroke-opacity:0.8;fill:none",
		DegreeTextStyle:    "font-family:Arial,sans-serif;font-size:12px;fill:#555555;text-anchor:middle",
		AltitudeTextStyle:  "font-family:Arial,sans-serif;font-size:11px;fill:#777777;text-anchor:start",
		DirectionTextStyle: "font-family:Arial,sans-serif;font-size:18px;font-weight:bold;fill:#000000;text-anchor:middle",
		BackgroundStyle:    "fill:#ffffff;stroke:none",

		TitleStyle:   "font-family:Georgia,serif;font-size:30px;font-weight:bold;fill:#000000;text-anchor:start",
		HeaderStyle:  "font-family:Georgia,serif;font-size:17px;font-weight:bold;fill:#000000;text-anchor:start",
		BodyStyle:    "font-family:Arial,sans-serif;font-size:15px;fill:#000000;text-anchor:start",
		NoteStyle:    "font-family:Arial,sans-serif;font-size:12px;font-style:italic;fill:#666666;text-anchor:start",
		DividerStyle: "stroke:#dddddd;stroke-width:1",

		StarSize:          6,
		StarStyle:         "fill:#000000;stroke:none",
		PathStyle:         "stroke:#333333;stroke-width:1;stroke-dasharray:4,3;fill:none",
		DistanceTextStyle: "font-family:Arial,sans-serif;font-size:11px;fill:#555555;text-anchor:middle",
		ObjectLabelStyle:  "font-family:Arial,sans-serif;font-size:12px;fill:#000000;text-anchor:middle",

		LegendSwatchSize: 16,
		LegendGap:        10,
	},
}

type ChartData struct {
	Title       string
	Location    coordinates.GeoLocation
	Time        time.Time
	Conditions  ChartConditions
	Objects     []Object
	Tour        *algo.Tour
	AngularTips []string
	ArmNote     string
	Legend      []LegendItem
}

type ChartConditions struct {
	Bortle            int
	Equipment         string
	LimitingMagnitude float32
}

type Object struct {
	Label  string
	Coords coordinates.Horizontal
}

type LegendMarker string

const (
	MarkerStar          LegendMarker = "star"
	MarkerPath          LegendMarker = "path"
	MarkerMilkyWay      LegendMarker = "milkyway"
	MarkerConstellation LegendMarker = "constellation"
)

type LegendItem struct {
	Label  string
	Marker LegendMarker
}

func DefaultAngularTips() []string {
	return []string{
		"Finger width ~ 1°",
		"Three middle fingers ~ 5°",
		"Fist width ~ 10°",
		"Index to pinky ~ 15°",
		"Open hand width ~ 20°",
	}
}

func DefaultLegend() []LegendItem {
	return []LegendItem{
		{Label: "Star", Marker: MarkerStar},
		{Label: "Path", Marker: MarkerPath},
		{Label: "Milky way", Marker: MarkerMilkyWay},
		{Label: "Constellation", Marker: MarkerConstellation},
	}
}
