// SPDX-FileCopyrightText: Copyright 2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

// Package difflib contains comparison tests for diff libraries.
//
// This package compares the output of four diff libraries:
//  1. github.com/go-openapi/testify/v2/internal/difflib (our internalized library - the yardstick)
//  2. github.com/hexops/gotextdiff (Myers algorithm implementation)
//  3. github.com/aymanbagabas/go-udiff (modern unified diff library)
//  4. github.com/sergi/go-diff/diffmatchpatch (Google's diff-match-patch port)
package difflib
