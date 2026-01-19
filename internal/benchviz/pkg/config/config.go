package config

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"go.yaml.in/yaml/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed default_config.yaml
var efs embed.FS

type MetricName string

const (
	MetricNsPerOp     MetricName = "nsPerOp"
	MetricAllocsPerOp MetricName = "allocsPerOp"
	MetricBytesPerOp  MetricName = "bytesPerOp"
	MetricMBPerS      MetricName = "MBytesPerS"
)

func (m MetricName) String() string {
	return string(m)
}

func (m MetricName) IsValid() bool {
	switch m {
	case MetricNsPerOp, MetricAllocsPerOp, MetricBytesPerOp, MetricMBPerS:
		return true
	default:
		return false
	}
}

func AllMetricNames() []MetricName {
	return []MetricName{
		MetricNsPerOp,
		MetricAllocsPerOp,
		MetricBytesPerOp,
		MetricMBPerS,
	}
}

type Config struct {
	Name        string
	Environment string
	Render      Rendering
	Outputs     Output
	Metrics     []Metric
	Functions   []Function
	Contexts    []Context
	Versions    []Version
	Categories  []Category
	Files       []File // Files allows for enrichments based on the input file name

	functionIndex map[string]Function
	contextIndex  map[string]Context
	versionIndex  map[string]Version
	metricIndex   map[MetricName]Metric
	// TODO: provision default context, version for regexp mismatches
}

/*
func (c Config) Open(file string) (io.ReadCloser, error) {
}
*/

func (c Config) GetFunction(id string) (Function, bool) {
	v, ok := c.functionIndex[id]

	return v, ok
}

func (c Config) GetContext(id string) (Context, bool) {
	v, ok := c.contextIndex[id]

	return v, ok
}

func (c Config) GetVersion(id string) (Version, bool) {
	v, ok := c.versionIndex[id]

	return v, ok
}

func (c Config) GetMetric(id MetricName) (Metric, bool) {
	v, ok := c.metricIndex[id]

	return v, ok
}

func (c Config) FindFunction(name string) (id string, ok bool) {
	for _, def := range c.Functions {
		if id, ok := def.MatchString(name); ok {
			return id, true
		}
	}

	return "", false
}

func (c Config) FindVersion(name string) (id string, ok bool) {
	for _, def := range c.Versions {
		if id, ok := def.MatchString(name); ok {
			return id, true
		}
	}

	return "", false
}

func (c Config) FindVersionFromFile(file string) (id string, ok bool) {
	for _, def := range c.Files {
		if _, ok := def.MatchString(file); !ok {
			continue
		}

		for _, version := range def.Versions {
			if id, ok := version.MatchString(file); ok {
				return id, true
			}
		}
	}

	return "", false
}

func (c Config) FindContext(name string) (id string, ok bool) {
	for _, def := range c.Versions {
		if id, ok := def.MatchString(name); ok {
			return id, true
		}
	}

	return "", false
}

func (c Config) FindContextFromFile(file string) (id string, ok bool) {
	for _, def := range c.Files {
		if _, ok := def.MatchString(file); !ok {
			continue
		}

		for _, context := range def.Contexts {
			if id, ok := context.MatchString(file); ok {
				return id, true
			}
		}
	}

	return "", false
}

type Rendering struct {
	Title       string
	Theme       string
	Layout      Layout
	Chart       string
	Legend      LegendPosition
	Scale       Scale
	DualScale   bool
	Orientation Orientation
}

type Orientation string

const (
	OrientationVertical   Orientation = "vertical"
	OrientationHorizontal Orientation = "horizontal"
)

type File struct {
	ID        string
	MatchFile string
	Contexts  []Context
	Versions  []Version

	match *regexp.Regexp
}

func (f File) MatchString(file string) (id string, ok bool) {
	if f.match == nil {
		return "", false
	}

	if ok := f.match.MatchString(file); !ok {
		return "", false
	}

	return f.ID, true
}

type Layout struct {
	Horizontal int
	Vertical   int
}

type Scale string

const (
	ScaleAuto Scale = "auto"
	ScaleLog  Scale = "log"
)

type LegendPosition string

const (
	LegendPositionNone   LegendPosition = "none"
	LegendPositionBotton LegendPosition = "bottom"
	LegendPositionTop    LegendPosition = "top"
	LegendPositionLeft   LegendPosition = "left"
	LegendPositionRight  LegendPosition = "right"
)

type Output struct {
	HtmlFile string
	PngFile  string
	IsTemp   bool
}

type Metric struct {
	ID    MetricName
	Title string
	Axis  string
}

type Object struct {
	ID       string
	Title    string
	Match    string
	NotMatch string
	match    *regexp.Regexp
	notMatch *regexp.Regexp
}

func (o Object) Matchers() (match, notMatch *regexp.Regexp) {
	return o.match, o.notMatch
}

func (o Object) MatchString(name string) (id string, ok bool) {
	var matchOk, notMatchOk bool
	id = o.ID
	matcher, notMatcher := o.Matchers()

	if matcher == nil && notMatcher == nil {
		return "", false
	}

	if matcher != nil {
		matchOk = matcher.MatchString(name)
	}

	if notMatcher != nil {
		notMatchOk = notMatcher.MatchString(name)
	}

	if matchOk && !notMatchOk {
		return id, true
	}

	if matcher == nil && !notMatchOk {
		return id, true
	}

	return "", false
}

type Function struct {
	Object `mapstructure:",squash"`
}

type Context struct {
	Object `mapstructure:",squash"`
}

type Version struct {
	Object `mapstructure:",squash"`
}

type Category struct {
	ID       string
	Title    string
	Includes Includes
}

type Includes struct {
	Functions []string
	Versions  []string
	Contexts  []string
	Metrics   []MetricName
}

func Load(file string) (*Config, error) {
	return load(os.DirFS("."), file)
}

func loadDefaults() (*Config, error) {
	return load(efs, "default_config.yaml")
}

func load(fsys fs.FS, file string) (*Config, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var raw any
	err = yaml.Unmarshal(content, &raw)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = mapstructure.Decode(raw, &cfg)
	if err != nil {
		return nil, err
	}

	// build indices and validate unique IDs
	cfg.functionIndex = make(map[string]Function, len(cfg.Functions))
	cfg.contextIndex = make(map[string]Context, len(cfg.Contexts))
	cfg.versionIndex = make(map[string]Version, len(cfg.Versions))
	cfg.metricIndex = make(map[MetricName]Metric, len(cfg.Metrics))

	for i, v := range cfg.Functions {
		if v.ID == "" {
			return nil, fmt.Errorf("invalid functions: empty ID found: functions[%d]", i)
		}
		if _, ok := cfg.functionIndex[v.ID]; ok {
			return nil, fmt.Errorf("invalid functions: duplicate ID key found: %s", v.ID)
		}
		if v.Title == "" {
			v.Title = titleize(v.ID)
		}
		cfg.functionIndex[v.ID] = v
	}

	for i, v := range cfg.Contexts {
		if v.ID == "" {
			return nil, fmt.Errorf("invalid contexts: empty ID found: contexts[%d]", i)
		}
		if _, ok := cfg.contextIndex[v.ID]; ok {
			return nil, fmt.Errorf("invalid contexts: duplicate ID key found: %s", v.ID)
		}
		if v.Title == "" {
			v.Title = titleize(v.ID)
		}
		cfg.contextIndex[v.ID] = v
	}

	for i, v := range cfg.Versions {
		if v.ID == "" {
			return nil, fmt.Errorf("invalid versions: empty ID found: versions[%d]", i)
		}
		if _, ok := cfg.versionIndex[v.ID]; ok {
			return nil, fmt.Errorf("invalid versions: duplicate ID key found: %s", v.ID)
		}
		if v.Title == "" {
			v.Title = titleize(v.ID)
		}
		cfg.versionIndex[v.ID] = v
	}

	for i, v := range cfg.Metrics {
		if v.ID == "" {
			return nil, fmt.Errorf("invalid metrics: empty ID found: metrics[%d]", i)
		}
		if !v.ID.IsValid() {
			return nil, fmt.Errorf("invalid metrics: invalid metric ID: metrics[%d]=%v (should be one of %v)", i, v.ID, AllMetricNames())
		}
		if v.Title == "" {
			v.Title = titleize(v.ID)
		}
		if _, ok := cfg.metricIndex[v.ID]; ok {
			return nil, fmt.Errorf("invalid metrics: duplicate ID key found: %s", v.ID)
		}
		// TODO: validate metric name
		cfg.metricIndex[v.ID] = v
	}

	// validate categories
	for i, v := range cfg.Categories {
		if v.ID == "" {
			return nil, fmt.Errorf("invalid categories: empty ID found: categories[%d]", i)
		}
		if v.Title == "" {
			v.Title = titleize(v.ID)
		}
		includes := v.Includes
		for j, ref := range includes.Functions {
			_, ok := cfg.functionIndex[ref]
			if !ok {
				return nil, fmt.Errorf("invalid category: function ID not found categories.%s.includes.functions[%d]=%s", v.ID, j, ref)
			}
		}

		if len(includes.Functions) == 0 {
			for _, injected := range cfg.Functions {
				v.Includes.Functions = append(v.Includes.Functions, injected.ID)
			}
		}

		for j, ref := range includes.Contexts {
			_, ok := cfg.contextIndex[ref]
			if !ok {
				return nil, fmt.Errorf("invalid category: context ID not found categories.%s.includes.contexts[%d]=%s", v.ID, j, ref)
			}
		}

		if len(includes.Contexts) == 0 {
			for _, injected := range cfg.Contexts {
				v.Includes.Contexts = append(v.Includes.Contexts, injected.ID)
			}
		}

		for j, ref := range includes.Versions {
			_, ok := cfg.versionIndex[ref]
			if !ok {
				return nil, fmt.Errorf("invalid category: version ID not found categories.%s.includes.versions[%d]=%s", v.ID, j, ref)
			}
		}

		if len(includes.Versions) == 0 {
			for _, injected := range cfg.Versions {
				v.Includes.Versions = append(v.Includes.Versions, injected.ID)
			}
		}

		for j, ref := range includes.Metrics {
			_, ok := cfg.metricIndex[ref]
			if !ok {
				return nil, fmt.Errorf("invalid category: metric ID not found categories.%s.includes.metrics[%d]=%s", v.ID, j, ref)
			}
			if j > 1 {
				return nil, fmt.Errorf("invalid category: up to 2 metrics can be included in a category. category.%s.metrics", v.ID)
			}
		}

		if len(includes.Metrics) == 0 {
			return nil, fmt.Errorf("invalid category: at least 1 metric must be included in a category. category.%s.metrics", v.ID)
		}
	}

	// parse all regexps
	for i, container := range cfg.Functions {
		match, notMatch, err := compileRex(container.Object)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp[function %d - %s]: %w", i, container.ID, err)
		}
		container.match = match
		container.notMatch = notMatch
		cfg.Functions[i] = container
	}

	for i, container := range cfg.Contexts {
		match, notMatch, err := compileRex(container.Object)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp[context %d - %s]: %w", i, container.ID, err)
		}
		container.match = match
		container.notMatch = notMatch
		cfg.Contexts[i] = container
	}

	for i, container := range cfg.Versions {
		match, notMatch, err := compileRex(container.Object)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp[version %d - %s]: %w", i, container.ID, err)
		}
		container.match = match
		container.notMatch = notMatch
		cfg.Versions[i] = container
	}

	for i, container := range cfg.Files {
		if container.ID == "" {
			return nil, fmt.Errorf("missing ID for file in files[%d]", i)
		}

		if container.MatchFile != "" {
			match, err := regexp.Compile(container.MatchFile)
			if err != nil {
				return nil, err
			}

			container.match = match
			for j, def := range container.Contexts {
				_, ok := cfg.contextIndex[def.ID]
				if !ok {
					return nil, fmt.Errorf("invalid file: context ID not found files[%d].context[%d]=%s", i, j, def.ID)
				}

				match, notMatch, err := compileRex(def.Object)
				if err != nil {
					return nil, fmt.Errorf("invalid regexp[files[%d].contexts[%d] - %s]: %w", i, j, def.ID, err)
				}
				def.match = match
				def.notMatch = notMatch
				container.Contexts[j] = def
			}

			for j, def := range container.Versions {
				_, ok := cfg.versionIndex[def.ID]
				if !ok {
					return nil, fmt.Errorf("invalid file: version ID not found files[%d].versions[%d]=%s", i, j, def.ID)
				}

				match, notMatch, err := compileRex(def.Object)
				if err != nil {
					return nil, fmt.Errorf("invalid regexp[files[%d].versions[%d] - %s]: %w", i, j, def.ID, err)
				}
				def.match = match
				def.notMatch = notMatch
				container.Versions[j] = def
			}

			cfg.Files[i] = container
		}
	}

	return &cfg, nil
}

func compileRex(o Object) (match, notMatch *regexp.Regexp, err error) {
	if o.Match != "" {
		match, err = regexp.Compile(o.Match)
		if err != nil {
			return nil, nil, err
		}
	}
	if o.NotMatch != "" {
		notMatch, err = regexp.Compile(o.NotMatch)
		if err != nil {
			return nil, nil, err
		}
	}

	return match, notMatch, nil
}

type str interface {
	~string
}

func titleize[T str](in T) string {
	caser := cases.Title(language.English, cases.NoLower) // the case is stateful: cannot declare it globally

	return caser.String(strings.Map(func(r rune) rune {
		switch r {
		case '_', '-':
			return ' '
		default:
			return r
		}
	}, string(in),
	))
}
