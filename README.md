# expect

A small, dependency-free, generics-based matcher library for Go's standard
`testing` package.

```go
expect.That(t, pkg.ImportPath).To(expect.Equal("example.com/math"))
expect.That(t, err).To(expect.Succeed())
expect.That(t, scans).To(expect.HaveLen(1))
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
`BeTrue`, `BeFalse`, `BeEmpty`, `HaveLen`, `BeNumerically`, `BeAnExistingFile`.

Each returns a `Matcher[T]`; use with `.To(...)` or `.NotTo(...)`.
