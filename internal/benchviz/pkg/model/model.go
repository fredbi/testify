package model

import "github.com/go-openapi/testify/internal/benchviz/v2/pkg/config"

// Scenario defines a complete configuration for benchmark visualization on a single page.
//
// A [Scenario] exposes several categories, each to be rendered in a separate chart on the page.
type Scenario struct {
	Name       string
	Categories []Category
}

// Category defines all the series for one or two metrics, regrouped on a single chart.
//
// Multiple versions correspond to several chart series represented side by side.
//
// Dual metric visualization implies a double scale.
type Category struct {
	ID          string
	Title       string
	Environment string
	Data        []CategoryData
}

const sensibleAlocations = 10 // preallocated labels
func (c Category) Labels() []string {
	labels := make([]string, 0, sensibleAlocations)

	for _, data := range c.Data {
		for _, series := range data.Series {
			for _, point := range series.Points {
				labels = append(labels, point.Name)
			}
		}
	}

	return labels
}

// CategoryData holds the data series for one metric and one version.
//
// Each series represented by a [CategoryData] is represented as one single data series on the chart.
//
// Each point of the data series corresponds to a context for the measurement.
type CategoryData struct {
	Version config.Version
	Metric  config.Metric
	Series  []MetricSeries
}

// SeriesKey uniquely identify a benchmark series.
//
// The keys to identify a series are: function, version, context and metric.
type SeriesKey struct {
	Function string
	Version  string
	Context  string
	Metric   config.MetricName
}

// MetricSeries correspond to a single series composed of points.
//
// The Title is used to display in a legend and corresponds to the version.
type MetricSeries struct {
	Title  string
	Points []MetricPoint
}

// MetricPoint is a single data point. Each data point has a label and a float64 value.
//
// The label is composed like "{function} - {context} - {version}" and may be used by tooltips
// when hovering over a data point.
type MetricPoint struct {
	Name  string
	Value float64
}
