---
name: fdleak — macOS support
description: Plan to add darwin implementation of internal/fdleak, refactoring the package around a Kind enum.
type: plan
---

# fdleak — macOS support

**Created:** 2026-04-17
**Status:** PLANNED
**Scope:** add darwin support to `internal/fdleak`. Windows deferred (possibly skipped entirely).

## Context

- `internal/fdleak` is currently **Linux-only**. It uses `/proc/self/fd` + `os.Readlink` to snapshot open FDs, and encodes kind (socket/pipe/anon_inode) into the target string via Linux-specific prefixes.
- Consumer: `internal/assertions/safety.go:NoFileDescriptorLeak`, which on non-Linux skips via `skipper`.
- The zero-dep policy rules out `golang.org/x/sys/unix`; cgo is also off the table. Everything must go through stdlib `syscall`.

## Decisions (locked with Fred, 2026-04-17)

1. **Introduce a `Kind` enum on `FDInfo`.** Filter on `Kind`, not on string prefix. This is a small API refactor but it's cleaner, and Windows — if ever — will need it too. Existing tests that assert filtering behaviour will be updated.
2. **macOS only in this pass.** Windows may be skipped entirely; revisit only if a user asks.
3. **Raw `syscall.Syscall(syscall.SYS_FCNTL, fd, F_GETPATH, buf)`** is acceptable. No cgo, no `golang.org/x/sys`.
4. **CI is the validation surface.** Fred does not run macOS locally. The CI macOS runner exercises the darwin path; we add tests that specifically cover the non-vnode paths (sockets, pipes) since those use the `fstat` fallback rather than `F_GETPATH`.

## macOS approach

- **Enumerate FDs:** `os.ReadDir("/dev/fd")`. `/dev/fd` is a devfs mount on darwin; entries have numeric names just like `/proc/self/fd`.
- **Resolve path:** raw syscall `fcntl(fd, F_GETPATH, buf)` where `F_GETPATH = 50` and `buf` is `MAXPATHLEN = 1024`. Returns the absolute path for vnode-backed FDs (regular files, devices).
- **Kind for non-vnode FDs:** when `F_GETPATH` returns an error (typical for sockets/pipes/kqueues), call `syscall.Fstat` and inspect `stat.Mode & syscall.S_IFMT`:
  - `S_IFSOCK` → `KindSocket`
  - `S_IFIFO` → `KindPipe`
  - `S_IFCHR` → `KindChar` (character device; `/dev/null`, `/dev/tty` etc. — usually resolvable via `F_GETPATH`, so this branch is a fallback only)
  - `S_IFREG`/`S_IFDIR`/`S_IFLNK`/`S_IFBLK` → `KindFile` (regular/dir/symlink/block device; path should be populated)
  - unknown → `KindOther`
- **kqueues and anon_inode analogues:** macOS doesn't expose a portable "anon inode" concept the way Linux does. kqueues show up through `fstat` as... not straightforward. We'll mark any FD where `F_GETPATH` fails **and** `fstat` doesn't match a known type as `KindOther` and filter it by default (same spirit as filtering `anon_inode:[…]` on Linux).

## New `FDInfo` shape

```go
type Kind int

const (
    KindUnknown Kind = iota
    KindFile        // regular file, directory, symlink, block device — anything with a real path
    KindSocket
    KindPipe
    KindChar        // character device without a resolvable path (edge case)
    KindOther       // kqueue, eventpoll, anon_inode — not user-opened
)

type FDInfo struct {
    FD     int
    Kind   Kind
    Target string // path if available, else a descriptive label like "socket:[inode]"
}

func (f FDInfo) isFiltered() bool {
    switch f.Kind {
    case KindSocket, KindPipe, KindOther:
        return true
    default:
        return false
    }
}
```

Target strings on darwin:
- `KindFile` → the `F_GETPATH` result.
- `KindSocket` → `"socket:[<st_ino>]"` (mimicking Linux for readability of leak reports).
- `KindPipe` → `"pipe:[<st_ino>]"`.
- `KindChar` → the `F_GETPATH` result, or `"char:[<st_rdev>]"` if unresolvable.
- `KindOther` → `"other:[<st_ino>]"` (or a more specific label if we can detect one cheaply).

Linux path adapts to produce the same Kind values from the existing prefix-based logic. Target strings stay backwards-compatible with today's Linux output.

## API shape (stdlib-style)

Exported API stays single-OS-agnostic. Platform differences are hidden behind an unexported
`snapshot()` whose body is selected at build time — same pattern as `os/file_unix.go` vs.
`os/file_windows.go`. `Snapshot` does no `runtime.GOOS` branching; build tags do that job.

```go
// fdleak.go — no build tag
type Kind int
type FDInfo struct { FD int; Kind Kind; Target string }
func (FDInfo) isFiltered() bool

var snapshotMu sync.Mutex

func Snapshot() (map[int]FDInfo, error) { return snapshot() }
func Leaked(tested func()) (string, error)       // unchanged orchestration
func Diff(before, after map[int]FDInfo) []FDInfo
func FormatLeaked(leaked []FDInfo) string

// fdleak_linux.go      //go:build linux
func snapshot() (map[int]FDInfo, error) { /* /proc/self/fd */ }

// fdleak_darwin.go     //go:build darwin
func snapshot() (map[int]FDInfo, error) { /* /dev/fd + fcntl F_GETPATH + fstat */ }

// fdleak_unsupported.go  //go:build !linux && !darwin
func snapshot() (map[int]FDInfo, error) {
    return nil, errors.New("file descriptor leak detection is not supported on " + runtime.GOOS)
}
```

## File layout

```
internal/fdleak/
  doc.go                  — update: no longer "Linux only"; describes supported platforms
  fdleak.go               — shared: FDInfo, Kind, isFiltered, Snapshot, Leaked, Diff, FormatLeaked, snapshotMu
  fdleak_linux.go         — //go:build linux              — unexported snapshot() via /proc/self/fd
  fdleak_darwin.go        — //go:build darwin             — unexported snapshot() via /dev/fd + fcntl + fstat
  fdleak_unsupported.go   — //go:build !linux && !darwin  — unexported snapshot() returning "not supported on <GOOS>"
  fdleak_test.go          — shared portable tests (Diff, FormatLeaked)
  fdleak_linux_test.go    — //go:build linux              — Snapshot/Leaked tests on Linux
  fdleak_darwin_test.go   — //go:build darwin             — Snapshot/Leaked tests on darwin, incl. socket/pipe filtering
```

File-naming note: the package is `fdleak`, the existing shared file is `fdleak.go`, so platform files follow the `fdleak_<GOOS>.go` convention (not `leak_<GOOS>.go`) — matches stdlib's `file_unix.go`/`file_windows.go` pattern scoped to this package's name.

Rationale: keeping tests guarded per-GOOS means CI runs exactly the right set on each platform without `t.Skip` noise. The portable subset (Diff/FormatLeaked) stays in the shared file.

## Consumer adjustments

- `internal/assertions/safety.go`:
  - Drop the `linuxOS` constant. Replace the gate with `supported := runtime.GOOS == "linux" || runtime.GOOS == "darwin"`.
  - Skip message becomes `"NoFileDescriptorLeak is not supported on <GOOS>"`.
  - Doc comment updated: now says "Linux and macOS" instead of "Linux only".
- `internal/assertions/safety_test.go`: remove any Linux-only skips on `NoFileDescriptorLeak` tests; they should now run on darwin CI too.

## Test plan

**Portable (fdleak_test.go):**
- `TestDiff` — already platform-agnostic. Update its fixtures to construct `FDInfo` with explicit `Kind` (breaking change inside the test; not a big deal).
- `TestDiff_NoLeaks`, `TestFormatLeaked`, `TestFormatLeaked_Empty` — no change beyond Kind-on-fixtures.

**Linux (fdleak_linux_test.go):**
- Move `TestSnapshot`, `TestLeaked_NoLeak`, `TestLeaked_WithLeak`, `TestLeaked_SocketsFiltered` here. Remove `skipIfNotLinux` — the build tag replaces it.

**Darwin (fdleak_darwin_test.go):**
- `TestSnapshot_Darwin` — asserts FDs 0/1/2 present, and Kind is something sensible (likely `KindChar` since stdin/stdout/stderr are ttys/devices on macOS).
- `TestLeaked_NoLeak_Darwin` — same as Linux version.
- `TestLeaked_WithLeak_Darwin` — temp file, expect it to be reported as `KindFile` with its path (validates `F_GETPATH`).
- `TestLeaked_SocketsFiltered_Darwin` — TCP listener, expect filtered (validates `fstat` fallback producing `KindSocket`).
- `TestLeaked_PipesFiltered_Darwin` — `os.Pipe()`, expect filtered (validates `KindPipe`).

The pipe test is new — we don't have its Linux equivalent today, but it's cheap and both platforms benefit from it.

## Risks / unknowns

- **`F_GETPATH` behaviour for `/dev/fd/N` itself.** If a test happens to be traversing `/dev/fd` while we snapshot, fcntl on the scanning directory FD should return its own directory; we already close it before returning, so not an issue in practice.
- **kqueue leaking from runtime.** Go's netpoller uses kqueue on darwin; the FD is long-lived and should be in `before` + `after` both, so `Diff` naturally ignores it. If any test accidentally triggers a fresh kqueue between snapshots, it'll be labelled `KindOther` and filtered — consistent with Linux filtering `anon_inode:[eventpoll]`.
- **darwin `SYS_FCNTL` trap number.** Provided by stdlib `syscall` package on darwin (`syscall.SYS_FCNTL`); defining `F_GETPATH = 50` as a local const is fine. Verify via `go doc syscall.SYS_FCNTL` on darwin, but this is well-established.

## Implementation checklist

Step 1: ascertain the inner working on package fdleak only
- [ ] Define `Kind` enum + refactor `FDInfo`, `isFiltered`, `Diff`, `FormatLeaked` in `fdleak.go`
- [ ] Move current Linux `/proc/self/fd` logic into unexported `snapshot()` in `fdleak_linux.go`, producing `Kind` from prefix detection
- [ ] Add `Snapshot()` wrapper in `fdleak.go` that simply calls `snapshot()`
- [ ] Split tests: portable (`fdleak_test.go`) vs linux (`fdleak_linux_test.go`) vs darwin (`fdleak_darwin_test.go`)
- [ ] Implement darwin `snapshot()` in `fdleak_darwin.go` (`/dev/fd` + `fcntl F_GETPATH` + `fstat`)
- [ ] Add darwin tests (Snapshot, NoLeak, WithLeak, SocketsFiltered, PipesFiltered)
- [ ] Add unsupported-platform stub `snapshot()` in `fdleak_unsupported.go` (`!linux && !darwin`)
- [ ] Ask Fred to push to CI and provide feedback from the MacOS runner on the unit tests for fdleak

Stage 2: wire the internal/assertions package

- [ ] Update `safety.go` to accept darwin; update doc comment
- [ ] Update `safety_test.go` to remove Linux-only skips on `NoFileDescriptorLeak`
- [ ] Update `fdleak/doc.go` wording
- [ ] Local: `go test work ./...` (exercises the portable + linux paths)
- [ ] Local: `golangci-lint run --new-from-rev master`
- [ ] CI: confirm the macOS runner exercises the darwin tests before merge (Fred)

Stage 3: final wrap-up: generate code & documentation
- [ ] generate code & docs / run tests in require & assert (Fred)
- [ ] update doc site pages to tout macos support
- [ ] final push & hopefully merge (Fred)

## Out of scope

- Windows support. NT handle enumeration would need `NtQuerySystemInformation` via `syscall.Syscall` into `ntdll.dll`. Tracked as a future consideration; may be skipped permanently depending on demand.
- Any change to `internal/leak` (goroutine leak detector) — already fully portable.
- Any change to `Leaked()` concurrency contract; still single-mutex process-wide.
