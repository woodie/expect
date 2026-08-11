# Working with expect

Cross-project conventions are in `~/workspace/woodie/docs/COWORK.md`.

## What this is

A generics-based matcher library, born out of migrating `gorderly`'s own
noisy `if x != y { t.Errorf(...) }` specs and `lambada`'s real Ginkgo/Gomega
suite onto `sclevine/spec` (see `~/workspace/gorderly`'s own
`docs/COWORK.md`, "Test-writing convention" section; `spec`'s own
`docs/COWORK.md` no longer exists -- its fork-only `docs/` folder was
wiped when `master` reset to plain `upstream/master`, see "Reversal"
below).

## How this came to exist

Started as a `gorderly`-only idea: pre-generics Go couldn't express
`expect(x).to(equal(y))` type-safely without `interface{}` and reflection,
which is exactly why testify/gomega read the way they do. Generics close
that gap directly. Originally scoped to live inside `gorderly` as an
internal package, since it was gorderly's own noise being cut down.

Promoted to its own module once `lambada` turned out to be a real second
consumer -- its Ginkgo/Gomega suite (`middleware_test.go`, `server_test.go`,
`scanfiles_test.go`, `main_test.go`) uses `Equal`, `HaveLen`, `BeEmpty`,
`HaveOccurred`, `Succeed`, `BeTrue`/`BeFalse`, `BeIdenticalTo`,
`BeNumerically`, `BeAnExistingFile`, and `ContainSubstring` -- a much wider
matcher surface than `gorderly` alone needed. This is the same shape
documented in `~/workspace/woodie/docs/COWORK.md`'s "Shared libraries across
sibling repos": one piece of logic, needed by more than one consumer, pulled
out rather than duplicated or left in the first repo that happened to need
it.

`cmd/lambada-mta`'s suite (`attachments_test.go`) added two more:
`BeADirectory` (distinct from `BeAnExistingFile` -- a real directory vs.
any resolvable path) and `Panic` (`Matcher[func()]`, calls the func and
recovers -- the one matcher whose `T` is a function rather than a plain
value). Every `BeNil()` call site in that file turned out to be checking an
`error`, so those became `Succeed()`, not a new general `BeNil` matcher --
another real-usage-first decision, same as dropping `HaveLen`/`BeEmpty`
below.

## Design

- `Expect(t, got)` returns an `Expectation[T]`; `.To(m)`/`.NotTo(m)` fail via
  `t.Errorf` (not `Fatal` -- matches spec's own it-blocks continuing to
  report every mismatch in one run, not stopping at the first).
- `Matcher[T]` is two methods, `Match(T) bool` and `String() string`. Adding
  a matcher is a constructor function plus a private struct; no shared base
  type or registration step.
- Dot-import is the recommended usage (see "Dot-imports: reversed" below) --
  a deliberate, scoped exception to the account's general avoidance of
  Ginkgo/Gomega's dot-import convention.
- `BeNumerically` takes the same `(op string, want T)` shape Gomega uses,
  deliberately, so porting a Gomega call site is close to a search-and-replace
  rather than a redesign.

## Dot-imports: reversed

Originally built with the account's usual no-dot-imports stance ("Every
call site reads `expect.That(...).To(expect.Equal(...))` in full"). Reversed
same session, per woodie's own framing: a real test file ends up full of
`expect.Expect(something).To(expect.Contain("whatever"))`, and if the goal
is a matcher library people actually want to reach for instead of
gomega/testify, the qualifier on every single line is exactly the clutter
working against that -- "if we take a safe route with a lot of clutter,
nobody will want to use our noisy expect package."

`That` renamed to `Expect` so porting a Gomega call site takes as little
effort as possible -- dot-imported, `Expect(t, got).To(Equal(want))` reads
the same way a Gomega call site does, `Expect(got).To(Equal(want))`, aside
from the explicit `t` (no `RegisterFailHandler`/ambient global here,
matching `spec`'s own no-global-state stance -- see `spec`'s
`docs/COWORK.md`).

This is scoped to `expect` specifically, not a reversal for `spec` too:
`spec`'s own exports (`Run`, `Before`, `After`, `G`, `S`) are common enough
words that dot-importing it would risk real collisions with other test
helpers in a way `expect`'s distinctive matcher names (`Equal`, `Contain`,
`Succeed`, `BeADirectory`, ...) don't. `expect`'s own test suite
(`expect_test.go`) now dot-imports itself too, both to dogfood the
recommended usage and to confirm it actually compiles under dot-import, not
just in the README's prose.

Update: resolved long since -- `lambada`'s test files call `Expect(...)`
(now `Expect(got, t)` as of `v0.2.0`, see below), not `That`. `humane` and
`gorderly` have since adopted the same dot-import convention too.

## This session: a real matcher reference, and a stated growth policy

The README's matcher list was a bare comma-separated name-drop -- no
signature, no description of what each one actually matches. Replaced with
a table (matcher, signature, matches) plus an "Adding a matcher" section
naming the pattern explicitly (constructor function + private struct,
`Match`/`String`, no shared registry), prompted by woodie asking for a
"clear list of matchers" and confirming this library should stay "open to
adding more."

Stated policy, made explicit rather than left implicit: grow this list
real-usage-first, not speculatively. Every matcher that exists today came
from an actual Ginkgo/Gomega call site in `gorderly` or `lambada` -- not a
guess at what Gomega has that might eventually be needed. `HaveLen`/`BeEmpty`
already went through the speculative-then-dropped cycle once (see "Not
built" above); the growth policy is the generalized version of that lesson,
not a new one.

## Not built

- No async/eventually-style matchers (Gomega's `Eventually`/`Consistently`).
  Nothing in `gorderly` or `lambada` currently needs polling assertions.
- No custom-matcher combinators (`And`/`Or`/`Not` wrapping other Matchers).
  Every real call site so far only needed one matcher at a time.
- ~~No `HaveLen`/`BeEmpty`.~~ Superseded -- see "Reversal: `ContainSubstring`,
  `HaveLen`, `BeEmpty` added after all" below. Drafted as `HaveLen[T
  any](n int) Matcher[T]`, then dropped while porting `lambada`'s real
  `Expect(scans).To(HaveLen(1))` call sites: neither argument is of type
  `T`, so every call would need an explicit type argument Go can't infer,
  e.g. `expect.HaveLen[[]scan](1)` -- exactly the ceremony the builtin
  `len()` already avoids. Real lesson from porting real call sites, not a
  hypothetical: `expect.That(t, len(scans)).To(expect.Equal(1))` reads just
  as well with zero new matcher code -- true as far as it went, but not the
  whole story; see the reversal below.

## expect_test.go now dogfoods spec too

Rewritten from plain `Test*`/`t.Run` functions to one `spec.RunAliased`
suite, using `describe`/`context`/`it`/`before`/`after` throughout -- not
just `expect`'s own dot-import, per woodie's ask to "go out of our way"
(within reason) to show off both libraries being used together, since this
file is the most-read example of the pairing. `context` is used once, for
`BeNumerically`'s unknown-operator case; every other matcher gets its own
`describe`. The `BeAnExistingFile`/`BeADirectory` block deliberately uses
`os.MkdirTemp`/`os.RemoveAll` instead of `t.TempDir()` specifically so
`after` has a real teardown to run, rather than being included for its own
sake with nothing to do.

Added `github.com/sclevine/spec` as a real `require` (pinned to `v1.4.0`,
matching `lambada`'s pin) plus a local `replace => ../spec`, test-only.
This doesn't weaken `expect`'s "dependency-free" pitch: Go only resolves a
module's test dependencies when building that module's own tests, not for
downstream code that imports `expect` as a library -- a consumer importing
`github.com/woodie/expect` never pulls in `spec` transitively.

The `spyT` test double (a `testing.TB` mock via embed-then-override) was
already here from the earlier CI fix; per the same ask, promoted from an
unexplained implementation detail to a documented pattern -- one-line
comments added at each override, plus a new README section ("Test doubles:
mocking a Go interface") generalizing the technique beyond `testing.TB`,
since it's broadly useful and Go has no built-in mocking library to point
to instead.

## v0.1.0 shipped; the local `../spec` replace was a real CI bug

First pass at the `spec` replace pointed at the local sibling checkout
(`replace github.com/sclevine/spec => ../spec`). That works on a Mac with
both repos checked out side by side, but GitHub Actions only checks out
`expect` itself -- CI failed with "replacement directory ../spec does not
exist" the first time it actually ran. Separately, `go test` also failed
locally at first: `spec`'s own `go.mod` still said `go 1.13` from 2019,
never bumped for `Var[T]`'s generics or `it.Context()`'s dependency on
`t.Context()` (1.24) -- both genuinely required `go 1.18`+/`1.24`+ to
compile at all.

Both fixed at the source: `spec`'s `go.mod` bumped to `go 1.24` and tagged
as its own `v0.1.0`, then `expect`'s replace repointed at the real tag
(`replace github.com/sclevine/spec => github.com/woodie/spec v0.1.0`)
instead of a local path -- works identically on a Mac and in CI, since it's
a real network fetch either way.

## Verification

Confirmed for real, not just by inspection: `go test ./...` passes on the
user's Mac (`ok github.com/woodie/expect 0.553s`), CI is green, and
`v0.1.0` is tagged and pushed. `lambada` (see its own `docs/COWORK.md`)
resolving `github.com/woodie/expect v0.1.0` as a real published module and
passing its full suite is the actual end-to-end proof this works outside
a local checkout.

## This session: `Expect(got, t)`, and a lowercase local-alias convention

Prompted by looking at `lambada`'s `cmd/lambada-web/main_test.go` once its
`spec` blocks were already lowercase `describe`/`context`/`it`/`before`/
`after` -- `Expect(t, got)` put the argument that reads like `self` first on
the line, ahead of the actual subject under test, which read as visually
"in your face" against the surrounding lowercase DSL. Explored in
`github.com/woodie/expect` issue #1 ("Consider `t.Expect()`"); worth
recording why neither of its two proposed paths survived contact with the
real file, and what shipped instead.

`t.Expect(got)` via wrapping `testing.T` in a own type doesn't just read
oddly -- it doesn't compile for a file like `main_test.go`, which asserts
`int`, `string`, `error`, and `bool` all under the same `t` in one
`spec.RunAliased` body. Go methods can't introduce their own type
parameters (only free functions can be generic), so a `T`-typed wrapper
struct can only ever fix one `T` for its lifetime -- it can't flex across
mixed-type assertions the way the current free `Expect[T any](...)`
does. `spec.Run`/`spec.RunAliased` also hardcode `*testing.T` in their own
signatures (`~/workspace/spec`'s `spec.go`/`aliases.go`) -- issue #1's other
path, a generic `spec.Run[T]`, doesn't exist and would ripple into every
other `spec` consumer, not just `expect`.

What actually shipped: `Expect[T any](t testing.TB, got T)` flipped to
`Expect[T any](got T, t testing.TB)` -- subject first, `t` a quiet trailing
detail instead of the lead argument -- plus a documented convention (see
README's "Lowercase call sites") for consumers to declare their own local,
one-line generic alias:

```go
func expect[T any](got T, t testing.TB) Expectation[T] { return Expect(got, t) }
```

This isn't a workaround, it's the actual unlock: Go's capitalize-to-export
rule only applies across the package boundary, so a real (non-closure)
generic function declared inside the *consuming* test package can be
lowercase with zero loss of compile-time type inference -- unlike a
closure, which can't declare its own type parameters at all, and unlike
the wrapped-`testing.T` approach above, which hits the one-`T`-per-value
ceiling. `expect_test.go` now declares and uses this alias itself, both to
dogfood the recommended shape and to confirm it actually compiles under
dot-import. This is a breaking change to `Expect`'s argument order --
`lambada`'s five test files (see its own `docs/COWORK.md`) were updated in
the same session. Tagged and published as `v0.2.0`, confirmed against
`lambada`'s real suite (55 specs) on the user's own Mac; `humane` (45
specs) and `gorderly` (29 specs) have since bumped their own pins and
confirmed clean too -- see each repo's own `docs/COWORK.md`.

## Same session: the `spec` suite body moved out of the inline closure

A follow-up to the `t` discussion above, this time about
`spec.RunAliased`/`spec.Run`'s own third argument -- the function itself,
not what's inside it. `TestExpect` used to read `spec.RunAliased(t,
"expect", func(t *testing.T, describe, context spec.Describe, it spec.S,
before, after func(func())) {...})`, the whole suite body inline inside
that one call. Go's function types are structural, so nothing about
`spec`'s own API needed to change to pull that closure out into a named,
top-level function with a matching signature:

```go
func TestExpect(t *testing.T) {
	spec.RunAliased(t, "expect", expectSuite)
}

func expectSuite(t *testing.T, describe, context spec.Describe, it spec.S, before, after func(func())) {
	describe("To/NotTo/ToNot", func() { ... })
	...
}
```

`TestExpect` is now one line; `expectSuite` holds the actual suite,
declared right below it in the same file. The parameter list itself still
has to be written out somewhere -- this doesn't shrink it, it just moves
it from an inline closure header to a named function's header, which reads
a little cleaner and keeps the `Test*` entry point trivial to scan. Not
done to any other file yet -- see `lambada`'s own `docs/COWORK.md` for
where this went next.

## Reversal: moved off the `woodie/spec` fork, back to plain upstream

Superseding the section above: `expectSuite` as a named top-level function
was itself walked back once `gorderly` and `lambada` both confirmed the
underlying fork wasn't pulling its weight. `spec.RunAliased`'s six-parameter
signature (`describe, context spec.Describe, it spec.S, before, after
func(func())`) exists to hand `before`/`after`/`context` in as bound
parameters -- but the one-line alternative, `context, before, after :=
describe, it.Before, it.After` written by hand as the first line inside a
plain `spec.Run` closure, needs nothing from the fork and reads more
plainly than either the six-parameter signature or a separate named suite
function does. `go.mod`'s `replace github.com/sclevine/spec =>
github.com/woodie/spec v0.1.0` is dropped; `expect_test.go` now calls
`spec.Run(t, "expect", func(t *testing.T, describe spec.G, it spec.S)
{...})` directly, with the destructuring line as the closure's first line
and a blank line under it before the real specs start (`spyT` and
`TestExpect` stay in `expect_test.go`; the lowercase `expect` alias moved
to a new `config_test.go`, matching the shape `gorderly` and `lambada`
both settled on -- see either repo's own `docs/COWORK.md`, "Reversal:
moved off the `woodie/spec` fork" section, for the fuller why). README
updated to point at `config_test.go` instead of `expect_test.go` for the
alias, and to drop its `github.com/woodie/spec` cross-reference.

## Second reversal: back on `woodie/spec`, this time for BeforeEach/AfterEach/JustBeforeEach

Different shape from the reversal above -- no `RunAliased`, no six-
parameter signature. `go.mod` picks up `woodie/spec` v0.2.0 via a plain
`replace` directive (`replace github.com/sclevine/spec =>
github.com/woodie/spec v0.2.0`), since the fork keeps upstream's module
path unchanged. The fork adds `BeforeEach`/`AfterEach`/`JustBeforeEach`
(`Before`/`After` still work, deprecated via `staticcheck`'s `SA1019`,
not removed) -- see `woodie/spec`'s own `docs/COWORK.md` for the full
history.

`expect_test.go`'s `before(...)`/`after(...)` calls (via the
`it.Before`/`it.After` method values) became `it.BeforeEach(...)`/
`it.AfterEach(...)`, called qualified rather than aliased -- three hook
names made the old one-line destructuring cluttered. Separately: the
`describe, it.Before, it.After` destructuring's first element flipped
direction. `expect_test.go` genuinely calls both `describe(...)` (for
each matcher: `describe("Equal", ...)`, `describe("Contain", ...)`, etc.)
and `context(...)` (once, for "given an unknown operator"), so the
`spec.Run` parameter is now named `context` and `describe := context` is
declared right below it -- matching the account-wide rule that the alias
only gets declared where a file actually calls the aliased name (see
`gorderly`'s `docs/COWORK.md` and `docs/FRAMEWORK.md` for the fuller
reasoning and the repos where no alias is needed at all).

Not yet verified against a real Go toolchain -- no Go in this sandbox.
`go mod tidy && make check` on the user's own Mac is the next step,
remembering to commit `go.sum` alongside `go.mod` before pushing.

## Reversal: `ContainSubstring`, `HaveLen`, `BeEmpty` added after all

Both "Contain, not ContainSubstring" and "no HaveLen/BeEmpty" were
documented as deliberate differences from Gomega -- true, but each one put
a footnote in the README that a person migrating a real Gomega call site
had to stop and read before they could tell whether their code still
compiled. Woodie's framing: `NotTo`/`ToNot` already establish the pattern
of shipping the Gomega spelling as a plain alias rather than making people
learn a shortened name, so the same treatment should apply here instead of
explaining the difference away.

`ContainSubstring` is now a one-line alias for `Contain`, same shape as
`ToNot`/`NotTo`. `HaveLen`/`BeEmpty` are real additions, not aliases -- no
existing matcher already covers them -- implemented via
`reflect.ValueOf(got).Len()` against `T any` rather than a
`comparable`/`cmp.Ordered`-style constraint, since length applies across
slices, arrays, maps, channels, and strings, none of which share a builtin
constraint. Both fall into the same "occasional explicit `[T]`" bucket
`BeIdenticalTo`/`BeNumerically` already established -- `HaveLen[T
any](n int)` can't infer `T` from `n` any more than the original draft
could, but that's no longer treated as disqualifying, since the README
already documents (and accepts) that ceremony for two other matchers.

This is a real reversal of the "grow real-usage-first, not speculatively"
growth policy for these three specifically -- neither came from an actual
`gorderly`/`lambada` call site this time. The reasoning stands for future
matchers; `ContainSubstring`/`HaveLen`/`BeEmpty` are the deliberate, named
exception, added for naming parity with Gomega rather than a call site
that needed them.

## Makefile added: build/test/lint/check, matching gorderly/gomeleon/humane

`expect` had no Makefile at all -- reviewed `gorderly`'s and `gomeleon`'s
(both binaries: `build`/`install`/`lint`/`test`/`check`) and `humane`'s
(a pure library like `expect`, so `lint`/`test`/`check` only, no
`build`/`install`) to find the right shape. `expect` has no `main`
package, so no binary to build or install -- `build` here is
`go build ./...`, a compile-only sanity check, with no `install` target
(nothing to copy to `~/go/bin`).

`test` pipes through `gorderly -fd`, matching `humane`'s own Makefile
exactly rather than plain `go test -v`: `expect`'s suite already dogfoods
`spec` (`describe`/`context`/`it`), and `spec`'s subtests join with `/` in
`go test -v`'s flat output -- `gorderly` is what turns that back into a
real indented RSpec-style tree, the same reason `humane` (also spec+expect)
reaches for it. `check` stays terse (silent on pass, full log on failure),
using plain `go test ./...` so it doesn't depend on `gorderly` being
installed -- same split every other repo's Makefile uses.

README's new "Development" section documents the four targets plus a
fallback (`go test -v ./...` / `go test ./...`) for a reader without
`gorderly` on `PATH`, matching `humane`'s README precedent.

## Matcher order now follows Gomega's own docs, not addition history

The README's matcher table (and, following it, `expect.go`'s function
order and `expect_test.go`'s `describe` order) used to read in the order
each matcher was added -- real-usage-first, chronological, no relation to
how Gomega documents its own matchers. Woodie's ask: line the list up with
Gomega's own category order (`Asserting Equivalence` -> `Presence` ->
`Truthiness` -> `Errors` -> `Channels` -> `Files` -> `Strings` ->
`Collections` -> `Structs` -> `Numbers and Times` -> `Values` -> `HTTP` ->
`Panics` -> `Composing`, per <https://onsi.github.io/gomega/#provided-matchers>)
so a reader comparing the two lists side by side isn't tripped up by
`expect`'s own history showing through.

Filtered down to just the matchers `expect` actually has, that order is:
`Equal`, `DeepEqual` (no direct Gomega equivalent -- kept next to `Equal`
since together they cover what Gomega's single reflection-based `Equal`
does), `BeIdenticalTo`, `BeTrue`, `BeFalse`, `HaveOccurred`, `Succeed`,
`BeAnExistingFile`, `BeADirectory`, `Contain`, `ContainSubstring`,
`BeEmpty`, `HaveLen`, `BeNumerically`, `Panic`. Two real reorderings
against the old chronological list: `HaveOccurred`/`Succeed` were swapped
(Gomega documents `HaveOccurred` first), and `BeTrue`/`BeFalse` moved from
after the Errors matchers to before them (Gomega's `Asserting Truthiness`
section precedes `Asserting on Errors`).

All three surfaces -- README table, `expect.go`'s matcher definitions, and
`expect_test.go`'s `describe` blocks -- were reordered together, a pure
code-motion pass with no behavior change, so a reader jumping between any
two of them sees the same sequence. Along the way, noticed `BeIdenticalTo`
has no test coverage in `expect_test.go` at all -- pre-existing gap, not
introduced by this reorder, left as-is since fixing it wasn't part of the
ask.

## v0.3.0 shipped: `ContainSubstring`/`HaveLen`/`BeEmpty`, all three consumers bumped

Tagged and released as `v0.3.0` (`docs/releases/v0.3.0.md`), confirmed via
`make check` on the user's own Mac before tagging, CI green on the push.
`gorderly`, `lambada`, and `humane` all bumped their `go.mod` pin from
`v0.2.0` to `v0.3.0` in the same session -- no breaking changes, so none of
the three needed their own version bump, just the pin update.

Real gotcha hit rolling out all three at once: the hand-off for each was
"run `go mod tidy`, then `make check`/`go test`," which regenerates
`go.sum` locally but doesn't commit it -- `git push`ing just the `go.mod`
bump commit (already staged/committed before the tidy ran) ships a stale
`go.sum`, and a clean CI checkout fails with `missing go.sum entry for
module providing package github.com/woodie/expect`. `gorderly` caught it
first; fixed with a follow-up commit adding the `go.sum` diff (not an
amend -- the bump was already pushed), same fix applied to `lambada` and
`humane` pre-emptively. Generalized into
`~/workspace/woodie/docs/COWORK.md`'s "Shared libraries across sibling
repos" section so the next library bump's hand-off states the commit step
explicitly instead of assuming it.

## `woodie/spec` module rename: dropped the `replace` directive

Picked up `woodie/spec` v0.3.0 (see that fork's own `docs/COWORK.md` and
`woodie/spec#3`), which renamed its module declaration to its own path.
`go.mod`'s `require github.com/sclevine/spec v1.4.0` plus `replace
github.com/sclevine/spec => github.com/woodie/spec v0.2.0` became a single
plain `require github.com/woodie/spec v0.3.0`. `expect_test.go`'s import
switched to match; README's "Setup" and "Learn more" sections rewritten
since both described the now-gone `replace`-based story.
