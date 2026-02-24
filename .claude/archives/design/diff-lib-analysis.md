# Comparative analysis of various diff libs

Two aspects to consider:
1. how to render go values
2. how to show their differences

## Rendering

1. github.com/go-openapi/testify/internal/spew (from github.com/davecgh/go-spew)
2. github.com/alecthomas/repr
3. https://github.com/kr/pretty
4. github.com/k0kubun/pp (colorized)
   -> ansi term code

### What to colorize?

* Error messages
* Expected vs Actual in diffs

No when output is not a tty

### Pending PRs at stretchr/testify

> * 🔍 **#1467** - Colorized output with terminal detection (most mature)
> * 🔍 **#1480** - Colorized diffs via TESTIFY_COLORED_DIFF env var
> * 🔍 **#1232** - Colorized output for expected/actual/errors
> * 🔍 **#994**  - Colorize expected vs actual values

1. #1467: golang.org/x/term v0.12.0 (<- IsTerminal)
2. #1480: github.com/sergi/go-diff
    -> https://github.com/kr/pretty
3. #1232: raw term codes. No deps
4. #994: raw term codes. No deps

### Proposed solution

* enable package that activates:
  * extra env & go test flags
  * detects terminal (internalized for golang.org)
  * no options: just import and defaults

* difflib adapted to support printer options
* internal/enable/colors contains basic ansi colors

Challenges:
* how to make it flexible, so users may inject options (like we have for yaml)?

## Show diff

The problem: unified diff

1. github.com/go-openapi/testify/internal/difflib
2. "github.com/hexops/gotextdiff"
   "github.com/hexops/gotextdiff/myers"
3. https://github.com/sergi/go-diff
   github.com/sergi/go-diff/diffmatchpatch


### Repr vs go-spew
Compare go-spew:

```go
(parser.expression) (len=1 cap=1) {
 (parser.alternative) (len=1 cap=1) {
  ([]interface {}) (len=1 cap=1) {
   (*parser.repitition)(0xc82000b220)({
     expression: (parser.expression) (len=2 cap=2) {
      (parser.alternative) (len=1 cap=1) {
      ([]interface {}) (len=1 cap=1) {
       (parser.str) (len=1) "a"
      }
     },
     (parser.alternative) (len=1 cap=1) {
      ([]interface {}) (len=1 cap=1) {
       (*parser.self)(0x593ef0)({
       })
      }
     }
    }
   })
  }
 }
}
```

To repr:

```go
parser.expression{
  parser.alternative{
    []interface {}{
      &parser.repitition{
        expression: parser.expression{
          parser.alternative{
            []interface {}{
              parser.str("a"),
            },
          },
          parser.alternative{
            []interface {}{
              &parser.self{              },
            },
          },
        },
      },
    },
  },
}
```

### gotextdiff vs difflib

TODO
