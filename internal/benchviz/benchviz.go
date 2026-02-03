package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/chart"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/config"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/image"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/organizer"
	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/parser"
)

type cliFlags struct {
	Config      string
	OutputFile  string
	IsJSON      bool
	Environment string
	l           *slog.Logger
}

func main() {
	var cli cliFlags

	// register CLI flags
	flag.BoolVar(&cli.IsJSON, "json", false, "read input from JSON")
	flag.StringVar(&cli.Config, "config", "config.yaml", "config file")
	flag.StringVar(&cli.Config, "c", "config.yaml", "config file (shorthand)")
	flag.StringVar(&cli.OutputFile, "output", "-", "file output or - for standard output")
	flag.StringVar(&cli.OutputFile, "o", "-", "file output or - for standard output (shorthand)")
	flag.StringVar(&cli.Environment, "environment", "-", "environment string")
	flag.StringVar(&cli.Environment, "e", "-", "environment string (shorthand)")

	// command line parsing, exit if invalid
	flag.Parse()

	// inject a structured logger
	cli.l = slog.Default().With(slog.String("module", "main"))

	if err := execute(cli, flag.Args()); err != nil {
		cli.l.Error(err.Error())
		log.Fatalf("%v", err)
	}
}

func execute(c cliFlags, args []string) error {
	cfg, cleanup, err := prepareConfig(c, args)
	defer cleanup()

	// 1. parse benchmark reports and build a chart page
	htmlRenderer, err := buildPage(cfg, args)
	if err != nil {
		return err
	}

	// 2. render the page as HTML, possibly to stdout, possibly to temp file
	html, htmlCloser, err := getWriter(cfg.Outputs.HtmlFile, "HTML")
	if err != nil {
		return err
	}

	if err := htmlRenderer.Render(html); err != nil {
		htmlCloser()
		return fmt.Errorf("rendering page: %w", err)
	}

	htmlCloser()

	if cfg.Outputs.PngFile == "" {
		// html only: we're done
		return nil
	}

	// 3. convert the HTML page to a PNG image, possibly to stdout
	html, htmlCloser, err = getReader(cfg.Outputs.HtmlFile, "HTML")
	if err != nil {
		return err
	}

	png, pngCloser, err := getWriter(cfg.Outputs.PngFile, "PNG")
	if err != nil {
		htmlCloser()
		return err
	}

	defer pngCloser()

	r := image.New()

	if err = r.Render(png, html); err != nil {
		return fmt.Errorf("rendering image: %w", err)
	}

	return nil
}

func getReader(file, kind string) (rdr *os.File, cleanup func(), err error) {
	rdr, err = os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("opening %s file: %q: %w", kind, file, err)
	}

	cleanup = func() {
		_ = rdr.Close()
	}

	return rdr, cleanup, nil
}

func getWriter(file, kind string) (wrt *os.File, cleanup func(), err error) {
	wrt, err = os.Create(file)
	if err != nil {
		return nil, nil, fmt.Errorf("opening %s file for writing: %q: %w", kind, file, err)
	}

	cleanup = func() {
		_ = wrt.Close()
	}

	return wrt, cleanup, nil
}

func prepareConfig(c cliFlags, args []string) (cfg *config.Config, cleanup func(), err error) {
	cfg, err = config.Load(c.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	if len(args) == 0 {
		args = append(args, "-")
	}

	if err = setConfig(cfg, c); err != nil {
		return nil, nil, fmt.Errorf("preparing config: %w", err)
	}

	if cfg.Outputs.IsTemp {
		cleanup = func() {
			_ = os.Remove(cfg.Outputs.HtmlFile)
		}

		return cfg, cleanup, err
	}

	return cfg, func() {}, err
}

// apply CLI flags overrides to YAML config
func setConfig(cfg *config.Config, c cliFlags) error {
	cfg.IsJSON = c.IsJSON

	if c.Environment != "" {
		cfg.Environment = c.Environment
	}

	if c.OutputFile != "" && c.OutputFile != "-" {
		// an outfile is defined: infer the PNG file from the HTML file provided
		cfg.Outputs.HtmlFile = inferHTMLFile(c.OutputFile)
		if cfg.Outputs.PngFile != "" { // override previously configured value
			cfg.Outputs.PngFile = inferImageFile(cfg.Outputs.HtmlFile)
		}
	}

	switch {
	case cfg.Outputs.HtmlFile == "" && cfg.Outputs.PngFile == "":
		c.l.Info("output sent to standard output as HTML, no PNG image rendered")
		cfg.Outputs.HtmlFile = "-"
	case cfg.Outputs.HtmlFile == "" && cfg.Outputs.PngFile != "":
		c.l.Info("HTML generated as a temporary file to produce PNG")
		tmp, err := os.CreateTemp("", "benchviz.*.html")
		if err != nil {
			return err
		}
		cfg.Outputs.HtmlFile = tmp.Name()
		cfg.Outputs.IsTemp = true
		_ = tmp.Close()
	}

	return nil
}

func buildPage(cfg *config.Config, args []string) (*chart.Page, error) {
	// 1. parse input benchmarks passed as CLI args
	p := parser.New(cfg, parser.WithParseJSON(cfg.IsJSON))
	if err := p.ParseFiles(args...); err != nil {
		return nil, fmt.Errorf("parsing files: %w", err)
	}

	// 2. re-organize the data series according to the configuration
	o := organizer.New(cfg)
	scenario := o.Scenarize(p.Sets())

	// 3. build a page with this visualization scenario
	builder := chart.New(cfg, scenario)
	page := builder.BuildPage()

	return page, nil
}

func inferHTMLFile(base string) string {
	ext := path.Ext(base)
	image, _ := strings.CutSuffix(base, ext)

	return image + ".html"
}

func inferImageFile(base string) string {
	ext := path.Ext(base)
	image, _ := strings.CutSuffix(base, ext)

	return image + ".png"
}
