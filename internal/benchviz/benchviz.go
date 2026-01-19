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
	cfg, err := config.Load(c.Config)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(args) == 0 {
		args = append(args, "-")
	}

	// 0. apply CLI flags overrides to YAML config
	if err = setConfigOverrides(cfg, c); err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	if cfg.Output.IsTemp {
		defer func() {
			_ = os.Remove(cfg.Output.HtmlFile)
		}()
	}

	// 1. parse input benchmarks passed as CLI args
	p := parser.New(cfg, WithParseJSON(cli.IsJSON))
	if err = p.ParseFiles(c.args...); err != nil {
		return fmt.Errorf("parsing files: %w", err)
	}

	// 2. re-organize the data series according to the configuration
	o := organizer.New(cfg)
	scenario := o.Scenarize(parsers.Sets())

	// 3. build a page with this visualization scenario
	c := chart.New(cfg, scenario)
	page := c.BuildPage()

	// 4. render the page as HTML, possibly to stdout, possibly to temp file
	if err := page.Render(); err != nil {
		return fmt.Errorf("rendering page: %w", err)
	}

	if cfg.Output.PngFile == "" {
		// html only: we're done
		return nil
	}

	// 5. render the page as a PNG image, possibly to stdout
	r := image.New(cfg.Output.HtmlFile, cfg.Output.PngFile)

	if err = r.Render(); err != nil {
		return fmt.Errorf("rendering image: %w", err)
	}

	return nil
}

func setConfigOverrides(cfg *config.Config, c cliFlags) error {
	if c.Environment != "" {
		cfg.Environment = c.Environment
	}

	if c.OutputFile != "" {
		cfg.Output.HtmlFile = c.OutputFile
		if cfg.Output.PngFile != "" { // override previously configured value
			cfg.Output.PngFile = inferImageFile(cfg.Output.HtmlFile)
		}
	}

	switch {
	case cfg.Output.HtmlFile == "" && cfg.Output.PngFile == "":
		c.l.Info("output sent to standard output as HTML, no PNG image rendered")
		cfg.Output.HtmlFile = "-"
	case cfg.Output.HtmlFile == "" && cfg.Output.PngFile != "":
		c.l.Info("HTML generated as a temporary file to produce PNG")
		tmp, err := os.CreateTemp("", "benchviz.*.html")
		if err != nil {
			return err
		}
		cfg.Output.HtmlFile = tmp.Name()
		cfg.Output.IsTemp = true
		_ = tmp.Close()
	}

	return nil
}

func inferImageFile(base string) string {
	ext := path.Ext(base)
	image, _ := strings.CutSuffix(base, ext)

	return image + ".png"
}
