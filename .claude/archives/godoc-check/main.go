package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	var cfg config

	flag.StringVar(&cfg.pkg, "package", "./...", "Package pattern")
	flag.BoolVar(&cfg.fix, "fix", false, "Apply fixes")
	flag.BoolVar(&cfg.updateWordlist, "update-wordlist", false, "Add false positives to wordlist without code fixes")
	flag.BoolVar(&cfg.reportOnly, "report-only", false, "Skip AI classification, just print raw findings")
	flag.StringVar(&cfg.model, "model", "claude-haiku-4-5-20251001", "Model for judgment calls")
	flag.StringVar(&cfg.workspace, "workspace", ".", "Workspace root")
	flag.StringVar(&cfg.wordlist, "wordlist", ".github/wordlist.txt", "Path to wordlist file relative to workspace")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Print detailed progress")
	flag.Parse()

	if cfg.reportOnly && (cfg.fix || cfg.updateWordlist) {
		fmt.Fprintln(os.Stderr, "Error: --report-only is mutually exclusive with --fix and --update-wordlist")
		os.Exit(2)
	}

	os.Exit(run(context.Background(), cfg))
}
