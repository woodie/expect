# expect

[![go.mod version](https://img.shields.io/github/go-mod/go-version/woodie/expect)](https://github.com/woodie/expect)
[![CI](https://github.com/woodie/expect/actions/workflows/ci.yml/badge.svg)](https://github.com/woodie/expect/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/woodie/lambada.svg)](https://github.com/woodie/lambada/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/woodie/expect.svg)](https://pkg.go.dev/github.com/woodie/expect)
[![License](https://img.shields.io/github/license/woodie/expect.svg)](LICENSE)

With `expect`, you get a small, dependency-free matcher library for Go's
standard `testing` package, built around Go 1.18+ generics.

This isn't a new test framework -- there's no runner to adopt, no global
registration, no config file. Every assertion is a plain function call
against whatever `*testing.T`/`*testing.B`/`testing.TB` your test already
has, so it drops straight into `go test`, table-driven tests, `spec`
suites, or anything else built on stdlib `testing`, with nothing new to
install and nothing to wire up beyond the import.

Before generics, a type-safe `Expect(x).To(Equal(y))`-style assertion
wasn't really possible without leaning on `interface{}` and reflection for
every matcher, plus a global fail handler or wrapper to route failures back
to the right `*testing.T` -- roughly the shape libraries like Gomega
settled on, for good reason at the time. Generics remove most of that need:
a matcher can be statically typed to what it actually compares, and an
assertion can just take `*testing.T` as a normal argument, no ambient state
required. That's the simpler shape `expect` takes, now that Go has caught
up -- no reflection where a type parameter already does the job, no
package-level registration step, just `testing.TB` in and `t.Errorf` out.

```go
import . "github.com/woodie/expect"

Expect(resp.StatusCode, t).To(Equal(200))
Expect(resp.StatusCode, t).NotTo(Equal(404))
Expect(tags, t).To(DeepEqual([]string{"go", "testing"}))
Expect(body, t).To(Contain("Available Scans"))
Expect(err, t).To(Succeed())
Expect(elapsed, t).To(BeNumerically[time.Duration]("<", time.Second))
Expect(srv.Handler, t).To(BeIdenticalTo[http.Handler](mux))
Expect(func() { mustParse("bad") }, t).To(Panic())
```

Dot-import is the intended, recommended usage -- a file full of
`expect.Expect(...).To(expect.Contain(...))` is exactly the clutter this
package exists to cut. `Expect`/`Equal`/`Contain`/etc. are distinctive
enough, and few enough, that collisions with anything else in a typical
test file are unlikely. (This is a deliberate exception to this account's
general dot-import avoidance -- see `docs/COWORK.md` for the reasoning,
which doesn't extend to `sclevine/spec`/`~/workspace/spec` itself: it stays
un-dot-imported, since its own exports -- `Run`, `Before`, `After`, `G`,
`S` -- are generic enough that dot-importing it would be a real collision
risk in a way `expect`'s aren't.)

If you'd rather not dot-import, every name still works qualified:
`expect.Expect(x, t).To(expect.Equal(y))`.

`got` comes first and `t` second, deliberately: in a `spec`/Ginkgo-style
file where `describe`/`context`/`it`/`before`/`after` already read as
lowercase structural keywords, `Expect(t, x)` put the one thing that looks
like `self` first on the line, ahead of the actual subject under test.
`Expect(x, t)` puts the subject first, `t` becomes a quiet trailing detail.

### Lowercase call sites

`Expect` has to stay capitalized -- Go only lets a dot-imported name be used
unqualified if it's exported, and exported means capitalized. But nothing
stops the *consuming* test package from declaring its own lowercase
pass-through, since that capitalization rule only applies across the
package boundary:

```go
func expect[T any](got T, t testing.TB) Expectation[T] { return Expect(got, t) }
```

One line, written once per test package (see `expect_test.go`'s own copy,
which every `it` in this repo's suite calls). It's a real generic function
declaration, not a closure, so it keeps full compile-time type inference --
nothing is traded away for the lowercase spelling. Every call site then
reads `expect(x, t).To(Equal(y))`, blending in with `describe`/`context`/
`it`/`before`/`after` instead of standing out as the one capitalized word
in the block.

`describe`/`context`/`it`/`before`/`after` get to be lowercase for a
different reason, worth not conflating with the above: they're just
parameter names in your own suite function's signature (see
[`spec`](https://github.com/woodie/spec)'s README), never crossing a
package boundary the way a called function like `Expect` does, so no
alias trick is needed there at all. See
[`gorderly`](https://github.com/woodie/gorderly)'s README for `spec`,
`expect`, and this alias all used together in one real suite, piped
through a real RSpec-style renderer.

### Where this differs from Gomega

The pitch is "you pass in `t`, otherwise it's Gomega" -- true for the large
majority of call sites, but not the whole story. Four real differences, kept
rather than papered over:

- **`Equal`/`DeepEqual` are two matchers, not one.** Gomega's `Equal` handles
  scalars and deep structural comparison together (it's reflection-based
  regardless of which). Go's `==` isn't defined on slices/maps, so a
  `comparable`-constrained `Equal` can't also do deep comparison -- `DeepEqual`
  exists for that case. Permanent, not a gap.
- **`Contain`, not `ContainSubstring`.** A deliberate shortening, not a miss.
- **No `HaveLen`/`BeEmpty`.** `Expect(x).To(HaveLen(n))` becomes `Expect(
  len(x), t).To(Equal(n))` -- Go's builtin `len()` already does this generically,
  so a wrapping matcher would only add ceremony (see "Matchers" below for why
  one was drafted and dropped).
- **Occasional explicit `[T]`.** `BeIdenticalTo[http.Handler](mux)`,
  `BeNumerically[time.Duration](">", 0)` -- needed when `want`'s own type
  doesn't match `got`'s closely enough for Go to infer `T`. Gomega never
  needs this, since it isn't generic.

Everything else -- `To`/`NotTo`/`ToNot`, matcher names, the overall call
shape -- is a deliberate match to Gomega's own vocabulary, so porting a real
Gomega call site is close to search-and-replace.

### Matchers

Every matcher returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`
(`.ToNot(...)` is an alias for `.NotTo(...)`, matching Gomega's own naming).

| Matcher | Signature | Matches |
|---|---|---|
| `Equal` | `Equal[T comparable](want T) Matcher[T]` | `got == want` |
| `DeepEqual` | `DeepEqual[T any](want T) Matcher[T]` | `reflect.DeepEqual(got, want)` -- slices, maps, structs |
| `BeIdenticalTo` | `BeIdenticalTo[T comparable](want T) Matcher[T]` | `got == want` -- named separately from `Equal` for pointer/interface identity checks |
| `Contain` | `Contain(substr string) Matcher[string]` | `strings.Contains(got, substr)` |
| `Succeed` | `Succeed() Matcher[error]` | `got == nil` |
| `HaveOccurred` | `HaveOccurred() Matcher[error]` | `got != nil` |
| `BeTrue` | `BeTrue() Matcher[bool]` | `got == true` |
| `BeFalse` | `BeFalse() Matcher[bool]` | `got == false` |
| `BeNumerically` | `BeNumerically[T cmp.Ordered](op string, want T) Matcher[T]` | `got` compared to `want` via `op` (`"=="`, `"!="`, `">"`, `">="`, `"<"`, `"<="`) |
| `BeAnExistingFile` | `BeAnExistingFile() Matcher[string]` | `os.Stat(got)` succeeds |
| `BeADirectory` | `BeADirectory() Matcher[string]` | `os.Stat(got)` succeeds and is a directory |
| `Panic` | `Panic() Matcher[func()]` | calling `got` panics |

No `HaveLen`/`BeEmpty` -- Go's builtin `len()` already works generically
over slices, maps, strings, arrays, and channels, so
`Expect(len(x), t).To(Equal(n))` needs no new matcher at all. A
`HaveLen[T any](n int) Matcher[T]` was drafted and dropped: since neither
of its arguments is of type T, every call site would need an explicit
`HaveLen[[]scan](1)` type argument, exactly the ceremony `len()` already
avoids.

### Adding a matcher

This list is expected to grow -- add matchers as real call sites need them,
not speculatively (every matcher above came from an actual Ginkgo/Gomega
call site in `gorderly` or `lambada`, not a guess at future need). The
pattern is always the same, four lines:

```go
type fooMatcher struct{ want Bar }

func (m fooMatcher) Match(got Bar) bool { return /* ... */ }
func (m fooMatcher) String() string     { return "be foo" }

func BeFoo(want Bar) Matcher[Bar] { return fooMatcher{want} }
```

No shared registry, no base type, no other file to touch. If the matcher
is generic, add `[T any]`/`[T comparable]`/`[T cmp.Ordered]` (whichever
constraint the `Match` body actually needs) to both the struct and the
constructor.

### Using a subject

Go has no `subject`/`let` keyword, but the pattern translates directly
once you have `spec`'s `before` (re-run fresh before every `it`, parent
before child): declare whatever `subject` depends on as plain local
variables in the enclosing `describe`, define `subject` itself as a
closure over them, and let a `before` at whichever nesting level actually
needs to change one set it.

```go
describe("DistanceInTime", func() {
	var at *time.Time
	var base time.Time
	var opts humane.TimeOptions
	subject := func() string { return humane.DistanceInTime(at, base, opts) }

	before(func() {
		base = time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
		opts = humane.TimeOptions{}
	})

	context("just now", func() {
		before(func() { at = ptr(base) })

		it("displays less than a minute ago", func() {
			expect(subject(), t).To(Equal("less than a minute ago"))
		})
	})

	context("45 seconds ago", func() {
		before(func() { at = ptr(base.Add(-45 * time.Second)) })

		it("rounds up to 1 minute ago", func() {
			expect(subject(), t).To(Equal("1 minute ago"))
		})
	})
})
```

`subject` is just a closure -- the same mechanism `before`/`after`
themselves are built on -- so it doesn't run until called, and
`subject()` inside each `it` reads whatever the `before` chain most
recently set. See [`humane`](https://github.com/woodie/humane)'s own
`size_test.go`/`time_test.go` for this pattern used across a real,
45-spec suite, covering both shapes that come up in practice: a subject
closing over a single varying input (`size_test.go`'s `subject := func()
string { return humane.HumanSize(bytes) }`), and one closing over several
independently-overridable inputs (`time_test.go`'s version above, where
different `context`s override `at`, `opts`, or both).

### Test doubles: mocking a Go interface

`expect_test.go` needs to verify that a *mismatched* assertion actually
fails -- but calling the real `t.Errorf` would fail the real test run,
which isn't what's being tested. Go has no built-in mocking, and
`testing.TB` specifically can't be implemented from scratch outside
package `testing` (it carries an unexported method on purpose). Embedding
sidesteps that: the embedded field gets you the interface for free via
method promotion, then you override just the methods you need to
intercept.

```go
type spyT struct {
    testing.TB
    failed bool
}

func (s *spyT) Helper() {}

func (s *spyT) Errorf(format string, args ...interface{}) {
    s.failed = true
}
```

`Helper` is a no-op so it doesn't fall through to the nil embedded value;
`Errorf` records the failure instead of reporting it. Pass a `*spyT`
anywhere a `testing.TB` is expected (`Expect(got, spy).To(...)`) and a
test can assert on `spy.failed` afterward. The same shape -- embed,
override, inspect -- works for stubbing any interface, not just
`testing.TB`. See `expect_test.go` itself for the fuller, more heavily
commented version of this exact pattern.
