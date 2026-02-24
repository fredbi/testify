# CI Release Notes Fixes — Work Plan

**Date:** 2026-02-08 (updated 2026-02-09)
**Branch (ci-workflows):** fix/monorepo-release-notes
**Fork for testing (round 1):** fredbi/testify
**Fork for testing (round 2, planned):** fork of go-openapi/swag

## Status: Rounds 1 & 2 complete + markdown polish done

Both testify (flat, v2) and swag (hierarchical, 15 modules, no v2) validated.
Markdown presentation polished. Temporary refs reverted. Ready for rebase and merge.

## Completed Fixes

### Structural fixes

- [x] **Shallow clone issue**: `git fetch --depth=1` was corrupting the tag graph after `actions/checkout`
  with `fetch-depth: 0`. Fixed by adding `git fetch --unshallow 2>/dev/null || true` before the tag ref fetch.
- [x] **Nested workflow refs**: `bump-release-monorepo.yml` and `bump-release.yml` use
  `./.github/workflows/` relative refs which resolve against the *caller* repo, not ci-workflows.
  Temporarily changed to absolute refs with branch for testing.
- [x] **Config URL chain**: `bump-release-monorepo.yml` and `bump-release.yml` have their own
  defaults for `cliff-config-url` and `monorepo-cliff-template-url` which get passed through to
  `release.yml`, overriding its defaults. All three levels temporarily pointed to the branch.
- [x] **Module sections "unreleased"**: Per-module `--tag-pattern` required previous module-specific
  tags (e.g., `codegen/v2.2.0`) to exist for git-cliff to resolve a bounded range. Fixed by using
  root tag pattern for all modules with `--include-path` for path filtering.
- [x] **Root module duplication**: Root module appeared in both Part 1 (full changelog) and Part 2
  (module sections). Fixed by skipping root module in the module loop.
- [x] **Broken --include-path from module names**: Module name stripping used `GITHUB_REPO` prefix
  which fails on forks (and had v2 path issues for internal/testintegration). Fixed by deriving
  include paths from filesystem paths (`ALL_FOLDERS`) instead of module names.
- [x] **Broken --exclude-path absolute paths**: `--exclude-path` used absolute filesystem paths
  which never matched git's repo-root-relative paths. Fixed to use relative paths with `/**` globs.
- [x] **Bash glob expansion**: `--include-path codegen/**` was glob-expanded by bash before reaching
  git-cliff. Fixed with `set -f` / `set +f` around the git-cliff invocation.
- [x] **Module version header**: `--include-path` filtering excluded tag commits from the filtered
  graph, so git-cliff couldn't resolve version boundaries. Fixed by passing `--tag "${RELEASE_TAG}"`.
- [x] **Tag message rendering**: Single newline between title and body meant git's
  `%(contents:subject)` joined everything as one line. Fixed with double newline separator.
  Body `|` separators now produce paragraph breaks (`\n\n`) instead of single newlines.

### .cliff.toml categorization fixes

- [x] **Merge commits**: Changed from `skip = true` to `group = "Other (technical)"`.
- [x] **Documentation garbled heading**: Removed overlapping `(documentation)` regex alternative
  that caused git-cliff to create separate "Documentationnumentation" groups.
- [x] **CI pattern**: Added `\bCI\b` for mid-message matches (e.g., "pass secrets in CI").
- [x] **Refactor precedence**: Removed `^` anchor from `refact` pattern so "chore(...): refactored"
  matches Refactor before the chore catch-all.
- [x] **"Other" renamed**: "Other" → "Other (technical)" (mostly merge commits).
- [x] **Bot contributor exclusion**: `bot-go-openapi[bot]` excluded from contributors and new
  contributors lists (commits still included). Copilot stays visible.
- [x] **Tag annotation in module sections**: Added `--with-tag-message ""` to suppress inherited
  tag annotations from module git-cliff calls.

## Round 2: swag fork — complete

**Fork:** fredbi/swag — **Release:** [v0.25.1](https://github.com/fredbi/swag/releases/tag/v0.25.1)

**Why**: Different characteristics exercise different code paths:
- Hierarchical module structure (parent/child nesting)
- No v2 module paths
- Less cross-module commit overlap
- Tests `--exclude-path` child module exclusion logic more thoroughly

**Results:** All 15 non-root modules rendered correctly on first run. No structural issues.
- Hierarchical exclusion works: `jsonutils` parent excludes child module commits
- All module sections have versioned `[0.25.1]` headers
- Tag message with markdown content (lists, block quotes) renders correctly
- Commit categorization correct across different commit history
- Only remaining area: optional markdown presentation polish — **done**

### Markdown presentation polish

- [x] Module headers simplified: `## module-name (version)` instead of separate `## name` + `### [version]`
- [x] `---` separator moved before each module section (between modules, not orphaned after header)
- [x] Removed sed header indentation (clean `#` → `##` → `###` hierarchy)
- [x] Empty modules (no commits) automatically skipped
- [x] Heading renamed: "Module-specific release notes" → "Per-module changes"
- [x] `.cliff-monorepo.toml` stripped of version header and `---` (now handled by release.yml assembly)
- [ ] License footer position (between Part 1 and Part 2): deferred — would require parsing git-cliff output

## Temporary changes to revert before merge

**In ci-workflows:** (all done)
- [x] `bump-release-monorepo.yml`: 3 nested `uses:` refs → reverted to relative refs
- [x] `bump-release-monorepo.yml`: 2 cliff config URL defaults → reverted to `refs/heads/master`
- [x] `bump-release.yml`: 1 nested `uses:` ref → reverted to relative ref
- [x] `bump-release.yml`: 1 cliff config URL default → reverted to `refs/heads/master`
- [x] `release.yml`: 2 cliff config URL defaults → reverted to `refs/heads/master`

**In testify:** (after ci-workflows merge + release)
- [ ] All 7 `.github/workflows/*.yml` files: refs → update to new SHA/tag after ci-workflows release
- [ ] `bump-release.yml`: remove `enable-tag-signing: 'false'` and `enable-commit-signing: 'false'`

## Files Modified (ci-workflows)

- `.github/workflows/release.yml` — main release notes generation (major rework of module loop)
- `.github/workflows/bump-release-monorepo.yml` — tag message construction, config URL defaults
- `.github/workflows/bump-release.yml` — tag message construction, config URL default
- `.cliff.toml` — commit categorization rules, contributor filtering

## Big Picture: go-openapi mono-repo landscape

These CI workflows are shared across multiple mono-repos. The design must handle
varying module counts, topologies, and naming conventions.

| Repo | Modules | Characteristics |
|------|---------|-----------------|
| **testify** | ~4 | v2 paths, flat structure, high cross-module commit overlap |
| **swag** | ~15 | hierarchical parent/child nesting, no v2, less overlap |
| **runtime** | 2 (maybe 3) | simple, potential otel opt-in module later |
| **strfmt** (planned) | 2-3 | internalized mongo v1 shim, opt-in mongo v2 driver |
| **core** (fredbi → go-openapi) | ~20 (post-restructure, down from ~50) | needs restructuring before transfer |

**Performance note:** Release notes generation (sequential git-cliff per module) is not the
bottleneck — the time-consuming step in `bump-release-monorepo` is waiting for CI on the
`prepare-release-monorepo` PR that updates go.mod files. Sequential git-cliff is fine up to ~20 modules.
