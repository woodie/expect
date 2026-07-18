# expect

Gomega, rethought for the current state of Go.

Gomega's shape (`Expect(x).To(Equal(y))`) predates generics -- every matcher
takes `interface{}` and leans on reflection, and getting the failure back to
the right `*testing.T` needs either a global `RegisterFailHandler(Fail)` or a
`NewWithT(t)` wrapper. Go 1.18+ generics remove most of the reason for either:
a matcher can be statically typed to what it actually compares, and an
assertion can just take `*testing.T` as a normal argument, no ambient state
required. `expect` is what Gomega's own call shape looks like once you can
assume generics -- which makes it, in one sense, a *more* conventional Go
library than Gomega itself: no reflection where a type parameter already
does the job, no package-level registration step, just `testing.TB` in and
`t.Errorf` out.

```go
import . "github.com/woodie/expect"

Expect(t, pkg.ImportPath).To(Equal("example.com/math"))
Expect(t, err).To(Succeed())
Expect(t, len(scans)).To(Equal(1))
Expect(t, rec.Body.String()).To(Contain("Available Scans"))
```

Dot-import is the intended, recommended usage -- a file full of
`expect.Expect(...).To(expect.Contain(...))` is exactly the clutter this
package exists to cut. `Expect`/`Equal`/`Contain`/etc. are distinctive
enough, and few enough, that collisions with anything else in a typical
test file are unlikely. (This is a deliberate exception to this account's
general dot-import avoidance -- see `docs/COWORK.md` for the reasoning.
`sclevine/spec`/`~/workspace/spec` itself stays un-dot-imported; its own
exports -- `Run`, `Before`, `After`, `G`, `S` -- are generic enough that
dot-importing it would be a real collision risk in a way `expect`'s aren't.)

If you'd rather not dot-import, every name still works qualified:
`expect.Expect(t, x).To(expect.Equal(y))`.

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
- **No `HaveLen`/`BeEmpty`.** `Expect(x).To(HaveLen(n))` becomes `Expect(t,
  len(x)).To(Equal(n))` -- Go's builtin `len()` already does this generically,
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
`Expect(t, len(x)).To(Equal(n))` needs no new matcher at all. A
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
