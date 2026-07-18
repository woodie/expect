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

`Equal`, `DeepEqual`, `BeIdenticalTo`, `Contain`, `Succeed`, `HaveOccurred`,
`BeTrue`, `BeFalse`, `BeNumerically`, `BeAnExistingFile`, `BeADirectory`,
`Panic`.

Each returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`.

No `HaveLen`/`BeEmpty` -- Go's builtin `len()` already works generically
over slices, maps, strings, arrays, and channels, so
`Expect(t, len(x)).To(Equal(n))` needs no new matcher at all. A
`HaveLen[T any](n int) Matcher[T]` was drafted and dropped: since neither
of its arguments is of type T, every call site would need an explicit
`HaveLen[[]scan](1)` type argument, exactly the ceremony `len()` already
avoids.

`BeIdenticalTo`/`BeNumerically` sometimes need an explicit type argument
too, when `want` is an untyped constant or concrete type but `got` is an
interface (`BeIdenticalTo[http.Handler](mux)`,
`BeNumerically[time.Duration](">", 0)`) -- Go's generic inference doesn't
look at how the result is used afterward, only at the arguments
themselves.
