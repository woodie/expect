# expect

A small, dependency-free, generics-based matcher library for Go's standard
`testing` package.

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

### Why not gomega/testify?

Generics (Go 1.18+) make `Expect(x).To(Equal(y))`-style matchers writable
without `interface{}` and reflection for the comparable cases, and without a
runner/reporting layer of their own -- `Expect(t, ...)` is a plain
`testing.TB` adapter, nothing more. Not reimplementing `testing`, same
principle `sclevine/spec` already applies to test organization.

### Matchers

Every matcher returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`.

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

`BeIdenticalTo`/`BeNumerically` sometimes need an explicit type argument,
when `want` is an untyped constant or concrete type but `got` is an
interface (`BeIdenticalTo[http.Handler](mux)`,
`BeNumerically[time.Duration](">", 0)`) -- Go's generic inference doesn't
look at how the result is used afterward, only at the arguments
themselves.

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
