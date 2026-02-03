package chart

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/config"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/model"
)

// Builder constructs charts from scenarized benchmark data.
type Builder struct {
	cfg      *config.Config
	scenario *model.Scenario
	l        *slog.Logger
}

// New creates a new chart [Builder], given a [config.Config] and a calculated [model.Scenario].
//
// The builder embeds a [slog.Logger] to croak about warnings and issues.
func New(cfg *config.Config, scenario *model.Scenario) *Builder {
	return &Builder{
		cfg:      cfg,
		scenario: scenario,
		l:        slog.Default().With(slog.String("module", "chart")),
	}
}

// BuildPage creates a page with all charts for all metrics and categories.
func (b *Builder) BuildPage() *Page {
	page := NewPage(b.scenario.Name)

	for _, category := range b.scenario.Categories {
		chart := b.buildChart(category)
		if chart == nil {
			b.l.Warn("empty chart skipped", slog.String("category_id", category.ID))

			continue
		}

		page.AddChart(chart)
	}

	b.l.Info("added charts", slog.Int("charts", len(page.charts)))

	return page
}

// buildChart creates a single chart for one metric (possibly two) and one category.
func (b *Builder) buildChart(category model.Category) *Chart {
	layoutConfig := b.cfg.Render
	showLegend := b.cfg.Render.Legend != config.LegendPositionNone
	title := category.Title
	xLabels := category.Labels()

	chart := NewChart(title, metric.String(),
		WithSubtitle(category.Environment),
		WithLegend(showLegend), // TODO: configurable legend position
	)
	chart.SetCategories(xLabels)

	for _, data := range category.Data { // iterate the series in a category
		for _, series := range data.Series { // each category, iterate over series
		}
	}
	/*
		x // Build X-axis labels: function × context
		xLabels := b.buildXLabels(category)
		if len(xLabels) == 0 {
			return nil
		}

		// Determine chart title
		title := b.chartTitle(category)

		// Create chart
		showLegend := len(cb.benchmarks.Versions) > 1
		chart := NewChart(title, metric.String(),
			WithSubtitle(cb.env),
			WithLegend(showLegend),
		)
		chart.SetCategories(xLabels)

		// Add series for each version
		for _, version := range cb.benchmarks.Versions {
			data := cb.buildSeriesData(metric, category, version, xLabels)
			chart.AddSeries(version, data)
		}
	*/

	return chart
}

/*
// buildXLabels creates X-axis labels for a category.
// Format: "Function - Context" or just "Function" if only one context.
func (b *Builder) buildXLabels(category Category) []string {
	var labels []string
	seen := make(map[string]struct{})

	// Determine which functions to include
	functions := category.Functions
	if len(functions) == 0 {
		functions = cb.benchmarks.Functions
	}

	// Determine which contexts to include
	contexts := category.Contexts
	if len(contexts) == 0 {
		contexts = cb.benchmarks.Contexts
	}

	for _, fn := range functions {
		for _, ctx := range contexts {
			// Check if we have data for this combination
			if cb.hasData(fn, ctx) {
				label := formatLabel(fn, ctx, len(contexts) > 1)
				if _, ok := seen[label]; !ok {
					labels = append(labels, label)
					seen[label] = struct{}{}
				}
			}
		}
	}

	return labels
}
*/

// hasData checks if there's any benchmark data for a function/context pair.
func (b *Builder) hasData(fn, ctx string) bool {
	for _, b := range cb.benchmarks.Benchmarks {
		if b.Function == fn && b.Context == ctx {
			return true
		}
	}
	return false
}

// buildSeriesData creates the data points for a series.
func (b *Builder) buildSeriesData(metric Metric, category Category, version string, xLabels []string) []opts.BarData {
	// Build set of valid contexts for this category
	contexts := category.Contexts
	if len(contexts) == 0 {
		contexts = cb.benchmarks.Contexts
	}
	validContexts := make(map[string]struct{})
	for _, ctx := range contexts {
		validContexts[ctx] = struct{}{}
	}

	// Build set of valid functions for this category
	functions := category.Functions
	if len(functions) == 0 {
		functions = cb.benchmarks.Functions
	}
	validFunctions := make(map[string]struct{})
	for _, fn := range functions {
		validFunctions[fn] = struct{}{}
	}

	// Build lookup map for quick access (only benchmarks in this category)
	lookup := make(map[string]*ParsedBenchmark)
	for i := range cb.benchmarks.Benchmarks {
		b := &cb.benchmarks.Benchmarks[i]
		if b.Version != version {
			continue
		}
		// Only include benchmarks whose context is in this category
		if _, ok := validContexts[b.Context]; !ok {
			continue
		}
		// Only include benchmarks whose function is in this category
		if _, ok := validFunctions[b.Function]; !ok {
			continue
		}
		key := formatLabel(b.Function, b.Context, len(contexts) > 1)
		lookup[key] = b
	}

	// Build data points in X-axis order
	data := make([]opts.BarData, 0, len(xLabels))
	for _, label := range xLabels {
		var value any
		if b, ok := lookup[label]; ok {
			value = cb.metricValue(b, metric)
		}
		data = append(data, opts.BarData{
			Name:  label,
			Value: value,
		})
	}

	return data
}

// metricValue extracts the metric value from a benchmark.
func (b *Builder) metricValue(b *ParsedBenchmark, metric Metric) any {
	switch metric {
	case MetricNsPerOp:
		return int(math.Round(b.NsPerOp))
	case MetricAllocsPerOp:
		return b.AllocsPerOp
	case MetricBytesPerOp:
		return b.BytesPerOp
	default:
		return nil
	}
}

// chartTitle generates a title for the chart.
//func (b *Builder) chartTitle(category model.Category) string {
//return category.Title
/*
	var metricName string
	switch metric {
	case MetricNsPerOp:
		metricName = "Timings"
	case MetricAllocsPerOp:
		metricName = "Allocations"
	case MetricBytesPerOp:
		metricName = "Memory"
	}

	if category.Name == "all" || len(cb.scenario.Categories) == 1 {
		return fmt.Sprintf("Benchmark %s", metricName)
	}
	return fmt.Sprintf("Benchmark %s (%s)", metricName, category.Name)
*/
//}

// formatLabel creates an X-axis label from function and context.
func formatLabel(fn, ctx string, includeContext bool) string {
	if !includeContext || ctx == "default" {
		return fn
	}
	return fmt.Sprintf("%s - %s", fn, ctx)
}
