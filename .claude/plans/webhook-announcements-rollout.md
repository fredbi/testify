# Rollout: webhook-announcements workflow + ci-workflows v0.4.0 bump

Generalize the announcements webhook (landed in testify) to all public, unarchived
go-openapi repos, and bump shared `ci-workflows` pins to v0.4.0
(`af4c93f45481ea7d24ac2a9858272cc03daf424e`).

## Per-repo changes

1. Add `.github/workflows/webhook-announcements.yml` (verbatim copy of testify's).
2. Bump every `go-openapi/ci-workflows/.github/workflows/*.yml@<sha> # <ver>` pin to
   `@af4c93f45481ea7d24ac2a9858272cc03daf424e # v0.4.0`.

## Scope

| Group | Repos | Changes |
|-------|-------|---------|
| Both | strfmt, codescan, analysis, loads, errors, runtime, spec, swag, validate, inflect, jsonpointer, jsonreference | 1 + 2 |
| Announce only (no ci-workflows refs) | gh-actions | 1 |
| Seed `## Announcements` + both | codegen, doc-site | 1 + 2 + seed section |
| Skip | testify (done), ci-workflows (owner), .github (org meta), core (private) | — |

## Procedure (per repo)

- `git fetch origin`; worktree `ci/webhook-announcements` off `origin/master`.
- Apply changes; `git commit -s` (author Fred, co-author Claude); push `-u`.
- `gh pr create`; wait for CI; merge if token allows, else leave for manual.

## Verification gates

- v0.4.0 tag == `af4c93f…` ✓; all referenced shared workflows exist at v0.4.0 ✓.
- After bump: no residual non-v0.4.0 ci-workflows pins.
- Pilot one repo (strfmt) end-to-end before fanning out.

## Status

- [x] Pilot: strfmt (#272)
- [x] Fan-out: 15 PRs opened (12 both + gh-actions + codegen/doc-site seeded)
- [x] CI green on all 15
- [ ] Merge — BLOCKED: every repo has `REVIEW_REQUIRED` branch protection; PAT
      cannot self-approve. Awaiting manual approval/merge (or `gh pr merge --admin`).

## Gotchas hit

- `scanner.yml` used uppercase `# V0.x.x` comments → regex needed `[vV]`.
- PAT lacks `workflow` scope → HTTPS pushes to `.github/workflows/` are rejected;
  push over SSH instead (gh-actions, doc-site have HTTPS origins).

## PRs

strfmt#272 codescan#51 analysis#210 loads#151 errors#114 runtime#494 spec#282
swag#213 validate#266 inflect#57 jsonpointer#139 jsonreference#104 gh-actions#106
codegen#11 doc-site#24
