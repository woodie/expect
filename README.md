# expect

A small, dependency-free, generics-based matcher library for Go's standard
`testing` package.

```go
expect.That(t, pkg.ImportPath).To(expect.Equal("example.com/math"))
expect.That(t, err).To(expect.Succeed())
expect.That(t, len(scans)).To(expect.Equal(1))
expect.That(t, rec.Body.String()).To(expect.Contain("Available Scans"))
```

### Why not gomega/testify?

Generics (Go 1.18+) make `expect(x).to(equal(y))`-style matchers writable
without `interface{}` and reflection for the comparable cases, and without a
runner/reporting layer of their own -- `expect.That(t, ...)` is a plain
`testing.TB` adapter, nothing more. Not reimplementing `testing`, same
principle `sclevine/spec` already applies to test organization.

### Matchers

`Equal`, `DeepEqual`, `BeIdenticalTo`, `Contain`, `Succeed`, `HaveOccurred`,
`BeTrue`, `BeFalse`, `BeNumerically`, `BeAnExistingFile`.

Each returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`.

No `HaveLen`/`BeEmpty` -- Go's builtin `len()` already works generically
over slices, maps, strings, arrays, and channels, so
`expect.That(t, len(x)).To(expect.Equal(n))` needs no new matcher at all.
A `HaveLen[T any](n int) Matcher[T]` was drafted and dropped: since neither
of its arguments is of type T, every call site would need an explicit
`expect.HaveLen[[]scan](1)` type argument, exactly the ceremony `len()`
already avoids.

`BeIdenticalTo`/`BeNumerically` sometimes need an explicit type argument
too, when `want` is an untyped constant or concrete type but `got` is an
interface (`expect.BeIdenticalTo[http.Handler](mux)`,
`expect.BeNumerically[time.Duration](">", 0)`) -- Go's generic inference
doesn't look at how the result is used afterward, only at the arguments
themselves.
