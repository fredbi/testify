package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/go-openapi/testify/internal/benchviz/v2/pkg/config"
	"golang.org/x/tools/benchmark/parse"
)

type Set struct {
	parse.Set

	File        string
	Environment string
}

type BenchmarkParser struct {
	options

	config *config.Config
	sets   []Set
	l      *slog.Logger
}

// New [BenchmarkParser] ready to parse benchmark files.
func New(cfg *config.Config, opts ...Option) *BenchmarkParser {
	return &BenchmarkParser{
		options: optionsWithDefaults(opts),
		config:  cfg,
		l:       slog.Default().With(slog.String("module", "parser")),
	}
}

func (p *BenchmarkParser) ParseFiles(files ...string) error {
	for _, file := range files {
		var (
			reader io.ReadCloser
			err    error
		)

		if file == "-" {
			reader = os.Stdin
		} else {
			reader, err = os.Open(file)
			if err != nil {
				return fmt.Errorf("input file %q: %w", file, err)
			}
		}

		set, err := p.ParseInput(reader)
		if err != nil {
			if file != "-" {
				_ = reader.Close()
			}

			return err
		}

		set.File = file
		p.sets = append(p.sets, set)

		if file != "-" {
			_ = reader.Close()
		}
	}

	p.l.Info("benchmark input parsed", slog.Int("parsed_files", len(files)))

	return nil
}

func (p *BenchmarkParser) ParseInput(r io.Reader) (Set, error) {
	if p.isJSON {
		return p.parseJSON(r)
	}

	return p.parseText(r)
}

func (p *BenchmarkParser) Sets() []Set {
	return p.sets
}

func (p *BenchmarkParser) parseText(r io.Reader) (Set, error) {
	// Read all input to extract environment info
	content, err := io.ReadAll(r) // TODO: replace with io.TeeReader
	if err != nil {
		return Set{}, fmt.Errorf("reading input: %w", err)
	}

	// Extract environment info
	environment := extractEnvironment(string(content))

	// Parse benchmarks
	set, err := parse.ParseSet(strings.NewReader(string(content)))
	if err != nil {
		return Set{}, err
	}

	s := Set{
		Set:         set,
		Environment: environment,
	}

	return s, nil
}

// parseJSON parses JSON output from `go test -json -bench`.
// It extracts the Output fields from "output" events and feeds them
// to the standard benchmark parser.
func (p *BenchmarkParser) parseJSON(r io.Reader) (Set, error) {
	// Read JSON events line by line and extract Output fields
	var textOutput strings.Builder
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event testEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// Skip lines that aren't valid JSON (shouldn't happen with -json flag)
			continue
		}

		// Only collect output from "output" action events
		if event.Action == "output" && event.Output != "" {
			textOutput.WriteString(event.Output)
		}
	}

	if err := scanner.Err(); err != nil {
		return Set{}, fmt.Errorf("scanning input: %w", err)
	}

	// Extract environment info
	outputText := textOutput.String()
	environment := extractEnvironment(outputText)

	// Now parse the collected text output using the standard parser
	set, err := parse.ParseSet(strings.NewReader(outputText))
	if err != nil {
		return Set{}, fmt.Errorf("parsing benchmark output: %w", err)
	}

	s := Set{
		Set:         set,
		Environment: environment,
	}

	return s, nil
}

// extractEnvironment extracts environment information from benchmark output.
// It looks for goos, goarch, and cpu lines and combines them.
func extractEnvironment(text string) string {
	var parts []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "goos: ") {
			parts = append(parts, strings.TrimPrefix(line, "goos: "))
		} else if strings.HasPrefix(line, "goarch: ") {
			parts = append(parts, strings.TrimPrefix(line, "goarch: "))
		} else if strings.HasPrefix(line, "cpu: ") {
			cpu := strings.TrimPrefix(line, "cpu: ")
			cpu = strings.TrimSpace(cpu)
			parts = append(parts, "cpu: "+cpu)
		}
	}

	if len(parts) == 0 {
		return "unknown environment"
	}

	return strings.Join(parts, " ")
}

// testEvent represents a single JSON event from `go test -json` output.
// See: https://pkg.go.dev/cmd/test2json
type testEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}
