package expect_test

import (
	"errors"
	"os"
	"testing"

	"github.com/sclevine/spec"
	. "github.com/woodie/expect"
)

// spyT is a test double for testing.TB. The interface can't be implemented
// from outside package testing on its own -- it carries an unexported method
// as a deliberate Go 1 compatibility guard -- but embedding testing.TB here
// satisfies that requirement through method promotion, without ever needing
// a real *testing.T. The embedded field is left nil; every method spyT
// actually calls (Helper, Errorf) is overridden below, so the nil value is
// never reached.
type spyT struct {
	testing.TB
	failed bool
}

// Helper is a no-op override. Expect calls t.Helper() on every assertion, and
// this keeps that call from falling through to the nil embedded testing.TB.
func (s *spyT) Helper() {}

// Errorf overrides the method a real *testing.T uses to report a failure.
// Instead of writing to a real test (which would fail the actual `go test`
// run), it just records that a failure happened, so a test here can assert
// that a mismatched matcher reports failure without the test process itself
// failing.
func (s *spyT) Errorf(format string, args ...interface{}) {
	s.failed = true
}

func TestExpect(t *testing.T) {
	spec.Run(t, "expect", func(t *testing.T, context spec.G, it spec.S) {
		describe := context

		describe("To ̸NotTo ̸ToNot", func() {
			var pass, fail *spyT

			it.BeforeEach(func() {
				pass, fail = &spyT{}, &spyT{}
			})

			it("passes To when the matcher matches", func() {
				expect(2+2, pass).To(Equal(4))
				expect(pass.failed, t).To(BeFalse())
			})

			it("fails To when the matcher doesn't match", func() {
				expect(2+2, fail).To(Equal(5))
				expect(fail.failed, t).To(BeTrue())
			})

			it("passes NotTo when the matcher doesn't match", func() {
				expect(2+2, pass).NotTo(Equal(5))
				expect(pass.failed, t).To(BeFalse())
			})

			it("fails NotTo when the matcher matches", func() {
				expect(2+2, fail).NotTo(Equal(4))
				expect(fail.failed, t).To(BeTrue())
			})

			it("ToNot behaves exactly like NotTo", func() {
				expect(2+2, pass).ToNot(Equal(5))
				expect(2+2, fail).ToNot(Equal(4))
				expect(pass.failed, t).To(BeFalse())
				expect(fail.failed, t).To(BeTrue())
			})
		})

		describe("Equal", func() {
			it("matches equal values", func() {
				expect(Equal(5).Match(5), t).To(BeTrue())
			})

			it("does not match different values", func() {
				expect(Equal(5).Match(6), t).To(BeFalse())
			})
		})

		describe("DeepEqual", func() {
			want := []string{"a", "b"}

			it("matches slices with equal contents", func() {
				expect(DeepEqual(want).Match([]string{"a", "b"}), t).To(BeTrue())
			})

			it("does not match slices with different contents", func() {
				expect(DeepEqual(want).Match([]string{"a", "c"}), t).To(BeFalse())
			})
		})

		describe("BeTrue and BeFalse", func() {
			it("BeTrue matches true and only true", func() {
				expect(BeTrue().Match(true), t).To(BeTrue())
				expect(BeTrue().Match(false), t).To(BeFalse())
			})

			it("BeFalse matches false and only false", func() {
				expect(BeFalse().Match(false), t).To(BeTrue())
				expect(BeFalse().Match(true), t).To(BeFalse())
			})
		})

		describe("HaveOccurred and Succeed", func() {
			it("HaveOccurred matches a non-nil error", func() {
				expect(HaveOccurred().Match(errors.New("boom")), t).To(BeTrue())
			})

			it("HaveOccurred does not match a nil error", func() {
				expect(HaveOccurred().Match(nil), t).To(BeFalse())
			})

			it("Succeed matches a nil error", func() {
				expect(Succeed().Match(nil), t).To(BeTrue())
			})

			it("Succeed does not match a non-nil error", func() {
				expect(Succeed().Match(errors.New("boom")), t).To(BeFalse())
			})
		})

		describe("BeAnExistingFile and BeADirectory", func() {
			var dir string

			it.BeforeEach(func() {
				dir, _ = os.MkdirTemp("", "expect-*")
			})

			// after runs explicit teardown, since dir here isn't managed by t.TempDir().
			it.AfterEach(func() {
				_ = os.RemoveAll(dir)
			})

			it("BeAnExistingFile matches a real path", func() {
				expect(BeAnExistingFile().Match(dir), t).To(BeTrue())
			})

			it("BeAnExistingFile does not match a missing path", func() {
				expect(BeAnExistingFile().Match("/definitely/does/not/exist"), t).To(BeFalse())
			})

			it("BeADirectory matches a real directory", func() {
				expect(BeADirectory().Match(dir), t).To(BeTrue())
			})

			it("BeADirectory does not match a missing path", func() {
				expect(BeADirectory().Match("/definitely/does/not/exist"), t).To(BeFalse())
			})
		})

		describe("Contain", func() {
			it("matches a substring", func() {
				expect(Contain("world").Match("hello world"), t).To(BeTrue())
			})

			it("does not match a missing substring", func() {
				expect(Contain("world").Match("hello"), t).To(BeFalse())
			})
		})

		describe("ContainSubstring", func() {
			it("is an alias for Contain", func() {
				expect(ContainSubstring("world").Match("hello world"), t).To(BeTrue())
				expect(ContainSubstring("world").Match("hello"), t).To(BeFalse())
			})
		})

		describe("BeEmpty", func() {
			it("matches an empty slice", func() {
				expect(BeEmpty[[]int]().Match(nil), t).To(BeTrue())
				expect(BeEmpty[[]int]().Match([]int{1}), t).To(BeFalse())
			})

			it("matches an empty string", func() {
				expect(BeEmpty[string]().Match(""), t).To(BeTrue())
				expect(BeEmpty[string]().Match("x"), t).To(BeFalse())
			})
		})

		describe("HaveLen", func() {
			it("matches a slice of the given length", func() {
				expect(HaveLen[[]int](3).Match([]int{1, 2, 3}), t).To(BeTrue())
				expect(HaveLen[[]int](3).Match([]int{1, 2}), t).To(BeFalse())
			})

			it("matches a string of the given length", func() {
				expect(HaveLen[string](5).Match("hello"), t).To(BeTrue())
			})
		})

		describe("BeNumerically", func() {
			it("supports > and <=", func() {
				expect(BeNumerically(">", 0).Match(1), t).To(BeTrue())
				expect(BeNumerically(">", 0).Match(0), t).To(BeFalse())
				expect(BeNumerically("<=", 5).Match(5), t).To(BeTrue())
			})

			context("given an unknown operator", func() {
				it("panics", func() {
					expect(func() { BeNumerically("~=", 0).Match(0) }, t).To(Panic())
				})
			})
		})

		describe("Panic", func() {
			it("matches a func that panics", func() {
				expect(Panic().Match(func() { panic("boom") }), t).To(BeTrue())
			})

			it("does not match a func that returns normally", func() {
				expect(Panic().Match(func() {}), t).To(BeFalse())
			})
		})
	})
}
