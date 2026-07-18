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

## Design

- `That(t, got)` returns an `Expectation[T]`; `.To(m)`/`.NotTo(m)` fail via
  `t.Errorf` (not `Fatal` -- matches spec's own it-blocks continuing to
  report every mismatch in one run, not stopping at the first).
- `Matcher[T]` is two methods, `Match(T) bool` and `String() string`. Adding
  a matcher is a constructor function plus a private struct; no shared base
  type or registration step.
- No dot-imports. Every call site reads `expect.That(...).To(expect.Equal(...))`
  in full -- consistent with the account's existing avoidance of Ginkgo/Gomega's
  dot-import convention.
- `BeNumerically` takes the same `(op string, want T)` shape Gomega uses,
  deliberately, so porting a Gomega call site is close to a search-and-replace
  rather than a redesign.

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
