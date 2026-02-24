package main

import (
	"os"
	"sort"
	"strings"
	"unicode"
)

func readWordlist(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// updateWordlist adds new words to the wordlist, preserving the sort convention:
// uppercase words first (sorted), then lowercase (sorted).
func updateWordlist(path string, newWords []string) error {
	existing := make(map[string]struct{})

	data, err := os.ReadFile(path)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			w := strings.TrimSpace(line)
			if w != "" {
				existing[w] = struct{}{}
			}
		}
	}

	for _, w := range newWords {
		existing[w] = struct{}{}
	}

	var upper, lower []string
	for w := range existing {
		if len(w) > 0 && unicode.IsUpper(rune(w[0])) {
			upper = append(upper, w)
		} else {
			lower = append(lower, w)
		}
	}

	sort.Strings(upper)
	sort.Strings(lower)

	all := append(upper, lower...)
	return os.WriteFile(path, []byte(strings.Join(all, "\n")+"\n"), 0o644)
}
