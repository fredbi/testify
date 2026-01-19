package chart

import (
	"io"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Theme constants from go-echarts.
const (
	ThemeRoma = "roma"
)

// Series represents a named data series in a chart.
type Series struct {
	Name string
	Data []opts.BarData
}

// Chart represents a benchmark comparison bar chart.
type Chart struct {
	Title      string
	Subtitle   string
	YAxisLabel string
	Theme      string
	ShowLegend bool
	Categories []string
	series     []Series
}

// ChartOption configures a Chart.
type ChartOption func(*Chart)

// WithSubtitle sets the chart subtitle (typically environment info).
func WithSubtitle(subtitle string) ChartOption {
	return func(c *Chart) {
		c.Subtitle = subtitle
	}
}

// WithTheme sets the ECharts theme.
func WithTheme(theme string) ChartOption {
	return func(c *Chart) {
		c.Theme = theme
	}
}

// WithLegend enables or disables the legend.
func WithLegend(show bool) ChartOption {
	return func(c *Chart) {
		c.ShowLegend = show
	}
}

// NewChart creates a new chart with the given title and y-axis label.
func NewChart(title, yAxisLabel string, options ...ChartOption) *Chart {
	c := &Chart{
		Title:      title,
		YAxisLabel: yAxisLabel,
		Theme:      ThemeRoma,
		ShowLegend: true,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// SetCategories sets the x-axis categories.
func (c *Chart) SetCategories(categories []string) {
	c.Categories = categories
}

// AddSeries adds a named data series to the chart.
func (c *Chart) AddSeries(name string, data []opts.BarData) {
	c.series = append(c.series, Series{Name: name, Data: data})
}

// Build creates the ECharts bar chart from the accumulated configuration.
func (c *Chart) Build() *charts.Bar {
	bar := charts.NewBar()

	// Title options
	titleOpts := opts.Title{
		Title: c.Title,
	}
	if c.Subtitle != "" {
		titleOpts.Subtitle = c.Subtitle
		titleOpts.SubtitleStyle = &opts.TextStyle{
			FontStyle: "italic",
			FontSize:  12,
		}
	}

	// Legend options
	legendOpts := opts.Legend{
		Show: opts.Bool(c.ShowLegend),
	}
	if c.ShowLegend {
		legendOpts.X = "right"
		legendOpts.Y = "bottom"
	}

	// X-axis options
	xAxisOpts := opts.XAxis{
		Name:         "Workload",
		Type:         "category",
		Position:     "bottom",
		NameLocation: "end",
		AxisTick: &opts.AxisTick{
			AlignWithLabel: opts.Bool(true),
		},
		AxisLabel: &opts.AxisLabel{
			Rotate:       30,
			Interval:     "0",
			ShowMinLabel: opts.Bool(true),
			ShowMaxLabel: opts.Bool(true),
			HideOverlap:  opts.Bool(false),
		},
	}

	// Y-axis options
	yAxisOpts := opts.YAxis{
		Name:  c.YAxisLabel,
		Type:  "value",
		Scale: opts.Bool(true),
		AxisLabel: &opts.AxisLabel{
			Formatter: opts.FuncOpts("function (value,index) { return value.toFixed(0).toString();}"),
		},
	}

	// Grid options
	gridOpts := opts.Grid{
		Bottom: "100",
		Top:    "100",
	}

	// Toolbox options
	toolboxOpts := opts.Toolbox{
		Left: "right",
		Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
				Title: "Save as image",
			},
		},
	}

	// Apply global options
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: c.Theme}),
		charts.WithToolboxOpts(toolboxOpts),
		charts.WithTitleOpts(titleOpts),
		charts.WithLegendOpts(legendOpts),
		charts.WithGridOpts(gridOpts),
		charts.WithXAxisOpts(xAxisOpts),
		charts.WithYAxisOpts(yAxisOpts),
	)

	// Set categories
	bar.SetXAxis(c.Categories)

	// Add all series
	for _, s := range c.series {
		bar.AddSeries(s.Name, s.Data)
	}

	return bar
}

// Page represents a page containing multiple charts.
type Page struct {
	Title  string
	charts []*Chart
}

// NewPage creates a new page with the given title.
func NewPage(title string) *Page {
	return &Page{
		Title: title,
	}
}

// AddChart adds a chart to the page.
func (p *Page) AddChart(c *Chart) {
	p.charts = append(p.charts, c)
}

// Render writes the page HTML to the given writer.
func (p *Page) Render(w io.Writer) error {
	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.SetPageTitle(p.Title)

	for _, c := range p.charts {
		page.AddCharts(c.Build())
	}

	return page.Render(w)
}
