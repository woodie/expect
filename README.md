# expect

[![go.mod version](https://img.shields.io/github/go-mod/go-version/woodie/expect)](https://github.com/woodie/expect)
[![CI](https://github.com/woodie/expect/actions/workflows/ci.yml/badge.svg)](https://github.com/woodie/expect/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/woodie/expect.svg)](https://github.com/woodie/expect/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/woodie/expect.svg)](https://pkg.go.dev/github.com/woodie/expect)
[![License](https://img.shields.io/github/license/woodie/expect.svg)](LICENSE)

A small, dependency-free matcher library for Go's standard `testing` package. Gomega-style 
assertions, built on Go generics, with no test runner or framework to adopt. Add a one-line 
lowercase alias (see "Setup" below) so call sites read lowercase, as in this example using
[`spec`](https://github.com/woodie/spec).

```go
func TestCalculator(t *testing.T) {
    spec.Run(t, "Calculator", func(t *testing.T, context spec.G, it spec.S) {
        describe := context

        var calculator *Calculator
        it.BeforeEach(func() { calculator = NewCalculator() })

        context("with 5 entered", func() {
            it.BeforeEach(func() { calculator.Enter(5) })

            describe("#DivideBy", func() {
                var divisor int
                subject := func() int { return calculator.DivideBy(divisor) }

                context("when the divisor is 1", func() {
                    it.BeforeEach(func() { divisor = 1 })

                    it("has no remainder", func() {
                        expect(subject(), t).To(Equal(0))
                    })
                })

                context("when the divisor is 3", func() {
                    it.BeforeEach(func() { divisor = 3 })

                    it("has a remainder of 2", func() {
                        expect(subject(), t).To(Equal(2))
                    })
                })
            })
        })
    })
}
```

Every assertion is a plain function call against whatever `*testing.T`/
`*testing.B`/`testing.TB` your test already has -- drops into `go test`,
table-driven tests, [`spec`](https://github.com/woodie/spec) suites, or
anything else built on stdlib `testing`, nothing to install beyond the
import.

## Installation

```
go get github.com/woodie/expect
```
Then dot-import the package into your test files.

## Setup

Dot-import `expect` and declare a local lowercase alias once per test
package -- a real (non-closure) generic function declared inside your own
package can be lowercase with zero loss of type inference, unlike a
dot-imported name, which has to stay capitalized:

```go
package calculator_test

import (
    "testing"

    "github.com/sclevine/spec"
    . "github.com/woodie/expect"
)

func expect[T any](got T, t testing.TB) Expectation[T] { return Expect(got, t) } // declared once per package
```

Every `it` in the package can then call `expect(...)` lowercase, as in the
example above. See
[`config_test.go`](https://github.com/woodie/expect/blob/main/config_test.go)
for the real version.

The import path stays `github.com/sclevine/spec` -- [`woodie/spec`](https://github.com/woodie/spec)
keeps upstream's own module path unchanged, so picking up the fork's
`BeforeEach`/`AfterEach`/`JustBeforeEach` (used above) is a `go.mod` `replace`
directive, not an import change:

```
replace github.com/sclevine/spec => github.com/woodie/spec v0.2.0
```

## Matchers

Every matcher returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`
(`.ToNot(...)` is an alias for `.NotTo(...)`).

| Matcher | Signature | Matches |
|---|---|---|
| `Equal` | `Equal[T comparable](want T) Matcher[T]` | `got == want` |
| `DeepEqual` | `DeepEqual[T any](want T) Matcher[T]` | `reflect.DeepEqual(got, want)` -- slices, maps, structs |
| `BeIdenticalTo` | `BeIdenticalTo[T comparable](want T) Matcher[T]` | `got == want` -- named separately from `Equal` for pointer/interface identity checks |
| `BeTrue` | `BeTrue() Matcher[bool]` | `got == true` |
| `BeFalse` | `BeFalse() Matcher[bool]` | `got == false` |
| `HaveOccurred` | `HaveOccurred() Matcher[error]` | `got != nil` |
| `Succeed` | `Succeed() Matcher[error]` | `got == nil` |
| `BeAnExistingFile` | `BeAnExistingFile() Matcher[string]` | `os.Stat(got)` succeeds |
| `BeADirectory` | `BeADirectory() Matcher[string]` | `os.Stat(got)` succeeds and is a directory |
| `Contain` | `Contain(substr string) Matcher[string]` | `strings.Contains(got, substr)` |
| `ContainSubstring` | `ContainSubstring(substr string) Matcher[string]` | alias for `Contain` |
| `BeEmpty` | `BeEmpty[T any]() Matcher[T]` | `got`'s length is zero |
| `HaveLen` | `HaveLen[T any](n int) Matcher[T]` | `got`'s length (slice, array, map, channel, or string) equals `n` |
| `BeNumerically` | `BeNumerically[T cmp.Ordered](op string, want T) Matcher[T]` | `got` compared to `want` via `op` (`"=="`, `"!="`, `">"`, `">="`, `"<"`, `"<="`) |
| `Panic` | `Panic() Matcher[func()]` | calling `got` panics |

This list grows from real call sites, not speculatively -- most matchers
above came from an actual Gomega call site ported into a real Go codebase,
not a guess at future need. `ContainSubstring`, `HaveLen`, and `BeEmpty`
are included directly to match Gomega's own names, so a ported call site
never has to change. Adding a new matcher is four lines, no registry or
base type to touch:

```go
type fooMatcher struct{ want Bar }

func (m fooMatcher) Match(got Bar) bool { return /* ... */ }
func (m fooMatcher) String() string     { return "be foo" }

func BeFoo(want Bar) Matcher[Bar] { return fooMatcher{want} }
```

## Where this differs from Gomega

The pitch is "you pass in `t`, otherwise it's Gomega" -- true for most
call sites, with two real differences:

- **`Equal`/`DeepEqual` are two matchers, not one.** Go's `==` isn't
  defined on slices/maps, so a `comparable`-constrained `Equal` can't also
  do deep comparison the way Gomega's reflection-based `Equal` does.
- **Occasional explicit `[T]`.** `BeIdenticalTo[http.Handler](mux)`,
  `BeNumerically[time.Duration](">", 0)`, `HaveLen[[]scan](1)` -- needed
  when `want`'s type doesn't let Go infer `T`. Gomega never needs this,
  since it isn't generic.

Everything else -- `To`/`NotTo`/`ToNot`, matcher names, overall call
shape -- matches Gomega's own vocabulary on purpose: a team can start on
plain `go test`, adopt `expect` and [`spec`](https://github.com/woodie/spec)
for the RSpec-style structure and generics-safe assertions, and still have
a clear path to Ginkgo/Gomega later if they ever need it -- porting a call
site either direction is close to search-and-replace. Gomega itself works
standalone too (no Ginkgo required), but still asks for a per-test
wrapper; `expect` skips that step and threads `t` straight into the call.
Same assertion, three ways, fewest lines last:

```go
// plain go test -- no matcher library at all.
if !f.HasCow() {
    t.Error("expected a cow")
}

// Gomega, standalone -- wrap *testing.T once, assert off the wrapper.
g := NewWithT(t)
g.Expect(f.HasCow()).To(BeTrue())

// expect -- no wrapper, t goes straight into the call.
Expect(f.HasCow(), t).To(BeTrue())
```

## Development

```
make build    # go build ./... -- expect is a library, nothing to install
make test     # verbose, dogfoods spec + gorderly on expect's own suite
make lint     # golangci-lint
make check    # terse: silent on success, full log on failure
```

`make test` pipes through [`gorderly`](https://github.com/woodie/gorderly)
for RSpec-style documentation output; without it installed, run
`go test -v ./...` or `go test ./...` directly instead.

## Learn more

- [`gorderly`](https://github.com/woodie/gorderly)'s [FRAMEWORK](https://github.com/woodie/gorderly/blob/main/docs/FRAMEWORK.md) --
  full suites combining `spec` + `expect`: context nesting, the `subject`
  pattern, stubbing, `httptest`, and interface test doubles.
- [`woodie/spec`](https://github.com/woodie/spec) -- the structural/lifecycle
  half of the pairing; `expect` has no dependency on it. Our fork of
  [`sclevine/spec`](https://github.com/sclevine/spec) (see that fork's own
  README for why it exists), picked up via the `go.mod` `replace` directive
  above for `BeforeEach`/`AfterEach`/`JustBeforeEach`.
