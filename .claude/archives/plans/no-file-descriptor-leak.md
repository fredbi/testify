# Plan: NoFileDescriptorLeak assertion

## Context

The v2.4 roadmap includes a `NoFileDescriptorLeak` assertion in the "safety" domain alongside the existing `NoGoRoutineLeak`. This is a Linux-only feature (via `/proc/self/fd`) that detects file descriptor leaks in tested code. It follows the same snapshot-before/after architecture, lives in the same domain, and uses the same codegen pipeline for all 8 variants.

Key constraints from the user:
- Linux only, `runtime.GOOS` check (not build tags)
- Not compatible with parallel tests (process-wide `/proc/self/fd`)
- Sockets filtered out by default (safer)
- Ships directly in `internal/assertions`, no "enable" gate
- Tests skipped on non-Linux via `t.Skip()`

## Architecture

Mirrors `NoGoRoutineLeak` / `internal/leak/`:

```
internal/fdleak/          ← new package (core logic)
├── doc.go
├── fdleak.go
└── fdleak_test.go

internal/assertions/      ← existing (assertion layer)
├── ifaces.go             ← add skipper interface
├── safety.go             ← add NoFileDescriptorLeak
├── safety_test.go        ← add tests
└── doc.go                ← update safety domain description
```

## Implementation

### 1. `internal/fdleak/doc.go` (new)

Package doc explaining the `/proc/self/fd` snapshot-and-diff approach, the Linux-only limitation, and the filtering strategy.

### 2. `internal/fdleak/fdleak.go` (new)

**Types:**

```go
// FDInfo describes an open file descriptor.
type FDInfo struct {
    FD     int
    Target string // readlink target (e.g. "/tmp/foo.txt", "socket:[12345]")
}
```

**Core functions:**

- `Snapshot() (map[int]FDInfo, error)` — reads `/proc/self/fd`, calls `os.Readlink` on each entry. FDs that close between `ReadDir` and `Readlink` are silently skipped (the transient ReadDir FD gets dropped this way naturally). Returns error if `runtime.GOOS != "linux"` (safety net).

- `Leaked(tested func()) (string, error)` — takes "before" snapshot, runs `tested()`, takes "after" snapshot, computes diff excluding filtered FD types, formats result. Returns empty string if clean. Uses a `sync.Mutex` to serialize concurrent calls (prevents false positives from parallel tests using this assertion, though tests will be serialized).

- `Diff(before, after map[int]FDInfo) []FDInfo` — returns FDs in `after` but not `before`, excluding sockets (`socket:[`), pipes (`pipe:[`), and anon inodes (`anon_inode:[`). These are runtime-internal FDs that tests shouldn't worry about.

- `FormatLeaked(leaked []FDInfo) string` — formats the error message:
  ```
  found 2 leaked file descriptor(s):
    fd 7: /tmp/testdata/unclosed.txt
    fd 9: /dev/null
  ```

**Filtering rules (applied in `Diff`):**
- `socket:[...]` — network connections, filtered by default
- `pipe:[...]` — Go runtime signal handling pipes
- `anon_inode:[...]` — epoll, eventfd, timerfd (Go runtime internals)

Only regular files, devices, and named pipes are reported as leaks.

### 3. `internal/fdleak/fdleak_test.go` (new)

Tests (Linux-only, skip on other platforms):

- `TestSnapshot` — verifies snapshot returns stdin/stdout/stderr (FDs 0, 1, 2)
- `TestLeaked_NoLeak` — clean function, empty result
- `TestLeaked_WithLeak` — opens a temp file without closing, expects detection
- `TestLeaked_SocketsFiltered` — opens a net.Listener, expects it to be filtered out
- `TestDiff` — unit test for the diff logic with mock data
- `TestFormatLeaked` — unit test for formatting

### 4. `internal/assertions/ifaces.go` (modify)

Add private `skipper` interface (mirrors existing `failNower`, `namer`, `contextualizer`):

```go
type skipper interface {
    Skip(args ...any)
}
```

### 5. `internal/assertions/safety.go` (modify)

Add `NoFileDescriptorLeak` after `NoGoRoutineLeak`:

```go
// NoFileDescriptorLeak ensures that no file descriptor leaks from inside the tested function.
//
// This assertion works on Linux only (via /proc/self/fd).
// On other platforms, the test is skipped.
//
// NOTE: this assertion is not compatible with parallel tests.
// File descriptors are a process-wide resource; concurrent tests
// opening files would cause false positives.
//
// Sockets, pipes, and anonymous inodes are filtered out by default,
// as these are typically managed by the Go runtime.
//
// # Concurrency
//
// [NoFileDescriptorLeak] serializes its snapshots with a mutex.
// Parallel tests using this assertion will not produce false positives,
// but they will run the tested function sequentially.
//
// # Usage
//
//	NoFileDescriptorLeak(t, func() {
//		// code that should not leak file descriptors
//	})
//
// # Examples
//
//	success: func() {}
func NoFileDescriptorLeak(t T, tested func(), msgAndArgs ...any) bool {
    // Domain: safety
}
```

Implementation:
1. Helper() call
2. `runtime.GOOS != "linux"` → type-assert to `skipper`, call `t.Skip(...)`, return `true`
3. Delegate to `fdleak.Leaked(tested)`
4. Empty string → return `true`
5. Non-empty → `Fail(t, msg, msgAndArgs...)`

Only `success` example — the failure case can't run on non-Linux generated tests. Real failure testing is in `safety_test.go`.

### 6. `internal/assertions/safety_test.go` (modify)

Add tests after existing `TestNoGoRoutineLeak_*`:

- `TestNoFileDescriptorLeak_Success` — clean function, passes on all platforms (returns true / skips)
- `TestNoFileDescriptorLeak_Failure` — Linux only (`if runtime.GOOS != "linux" { t.Skip(...) }`): opens temp file without closing, verifies assertion returns false and mockT reports failure
- `TestNoFileDescriptorLeak_SocketFiltered` — Linux only: creates net.Listener (leaks socket FD), verifies assertion returns true (socket filtered)

### 7. `internal/assertions/doc.go` (modify)

Update safety domain description to mention both goroutine and file descriptor leak detection.

### 8. Codegen + docs

Run `go generate ./...` to produce all 8 variants and documentation updates. The generated `docs/doc-site/api/safety.md` will include the new assertion.

## Verification

1. `cd internal/fdleak && go test -v ./...` — unit tests for core logic
2. `cd internal/assertions && go test -v -run TestNoFileDescriptorLeak` — assertion tests
3. `go generate ./...` — regenerate all variants
4. `go test ./...` — full test suite (generated tests must pass)
5. `golangci-lint run ./internal/fdleak/... ./internal/assertions/...` — no lint issues
