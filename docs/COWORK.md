# Working with expect

Cross-project conventions are in `~/workspace/woodie/docs/COWORK.md`.

## What this is

A generics-based matcher library, born out of migrating `gorderly`'s own
noisy `if x != y { t.Errorf(...) }` specs and `lambada`'s real Ginkgo/Gomega
suite onto `sclevine/spec` (see `~/workspace/spec`'s own `docs/COWORK.md` and
`~/workspace/gorderly`'s, "Test-writing convention" section).

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

`That` renamed to `Expect` to match -- dot-imported, `Expect(t,
got).To(Equal(want))` reads identically to Gomega's own
`Expect(got).To(Equal(want))`, the one difference being the explicit `t`
(no `RegisterFailHandler`/ambient global here, matching `spec`'s own
no-global-state stance -- see `spec`'s `docs/COWORK.md`).

This is scoped to `expect` specifically, not a reversal for `spec` too:
`spec`'s own exports (`Run`, `Before`, `After`, `G`, `S`) are common enough
words that dot-importing it would risk real collisions with other test
helpers in a way `expect`'s distinctive matcher names (`Equal`, `Contain`,
`Succeed`, `BeADirectory`, ...) don't. `expect`'s own test suite
(`expect_test.go`) now dot-imports itself too, both to dogfood the
recommended usage and to confirm it actually compiles under dot-import, not
just in the README's prose.

Consumers not yet updated to this: `lambada`'s migrated test files still
call `expect.That(...)` as of this writing -- see `lambada`'s own
`docs/COWORK.md` for that follow-up.

## Not built

- No async/eventually-style matchers (Gomega's `Eventually`/`Consistently`).
  Nothing in `gorderly` or `lambada` currently needs polling assertions.
- No custom-matcher combinators (`And`/`Or`/`Not` wrapping other Matchers).
  Every real call site so far only needed one matcher at a time.
- No `HaveLen`/`BeEmpty`. Drafted as `HaveLen[T any](n int) Matcher[T]`,
  then dropped while porting `lambada`'s real `Expect(scans).To(HaveLen(1))`
  call sites: neither argument is of type `T`, so every call would need an
  explicit type argument Go can't infer, e.g. `expect.HaveLen[[]scan](1)` --
  exactly the ceremony the builtin `len()` already avoids. Real lesson from
  porting real call sites, not a hypothetical: `expect.That(t,
  len(scans)).To(expect.Equal(1))` reads just as well with zero new
  matcher code.

## Verification

No Go toolchain in this sandbox (same situation as `gorderly`/`lambada`) --
`expect.go`/`expect_test.go` written by inspection, not yet run for real.
Needs, on your Mac:

```
cd ~/workspace/expect
go test -v ./...
```
