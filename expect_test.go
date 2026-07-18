package expect_test

import (
	"errors"
	"os"
	"testing"

	"github.com/sclevine/spec"
	. "github.com/woodie/expect"
)

// spyT is a test double for testing.TB: embed the real interface, then override only the methods you need to intercept -- the same technique works for mocking any interface, not just testing.TB.
type spyT struct {
	testing.TB
	failed bool
}

func (s *spyT) Helper() {}

// Errorf is overridden so a deliberately-mismatched matcher records failure here instead of failing this actual test run.
func (s *spyT) Errorf(format string, args ...interface{}) {
	s.failed = true
}

func TestExpect(t *testing.T) {
	spec.RunAliased(t, "expect", func(t *testing.T, describe, context spec.Describe, it spec.S, before, after func(func())) {
		describe("To/NotTo/ToNot", func() {
			var pass, fail *spyT

			before(func() {
				pass, fail = &spyT{}, &spyT{}
			})

			it("passes To when the matcher matches", func() {
				Expect(pass, 2+2).To(Equal(4))
				Expect(t, pass.failed).To(BeFalse())
			})

			it("fails To when the matcher doesn't match", func() {
				Expect(fail, 2+2).To(Equal(5))
				Expect(t, fail.failed).To(BeTrue())
			})

			it("passes NotTo when the matcher doesn't match", func() {
				Expect(pass, 2+2).NotTo(Equal(5))
				Expect(t, pass.failed).To(BeFalse())
			})

			it("fails NotTo when the matcher matches", func() {
				Expect(fail, 2+2).NotTo(Equal(4))
				Expect(t, fail.failed).To(BeTrue())
			})

			it("ToNot behaves exactly like NotTo", func() {
				Expect(pass, 2+2).ToNot(Equal(5))
				Expect(fail, 2+2).ToNot(Equal(4))
				Expect(t, pass.failed).To(BeFalse())
				Expect(t, fail.failed).To(BeTrue())
			})
		})

		describe("Equal", func() {
			it("matches equal values", func() {
				Expect(t, Equal(5).Match(5)).To(BeTrue())
			})

			it("does not match different values", func() {
				Expect(t, Equal(5).Match(6)).To(BeFalse())
			})
		})

		describe("DeepEqual", func() {
			want := []string{"a", "b"}

			it("matches slices with equal contents", func() {
				Expect(t, DeepEqual(want).Match([]string{"a", "b"})).To(BeTrue())
			})

			it("does not match slices with different contents", func() {
				Expect(t, DeepEqual(want).Match([]string{"a", "c"})).To(BeFalse())
			})
		})

		describe("Contain", func() {
			it("matches a substring", func() {
				Expect(t, Contain("world").Match("hello world")).To(BeTrue())
			})

			it("does not match a missing substring", func() {
				Expect(t, Contain("world").Match("hello")).To(BeFalse())
			})
		})

		describe("Succeed and HaveOccurred", func() {
			it("Succeed matches a nil error", func() {
				Expect(t, Succeed().Match(nil)).To(BeTrue())
			})

			it("Succeed does not match a non-nil error", func() {
				Expect(t, Succeed().Match(errors.New("boom"))).To(BeFalse())
			})

			it("HaveOccurred matches a non-nil error", func() {
				Expect(t, HaveOccurred().Match(errors.New("boom"))).To(BeTrue())
			})

			it("HaveOccurred does not match a nil error", func() {
				Expect(t, HaveOccurred().Match(nil)).To(BeFalse())
			})
		})

		describe("BeTrue and BeFalse", func() {
			it("BeTrue matches true and only true", func() {
				Expect(t, BeTrue().Match(true)).To(BeTrue())
				Expect(t, BeTrue().Match(false)).To(BeFalse())
			})

			it("BeFalse matches false and only false", func() {
				Expect(t, BeFalse().Match(false)).To(BeTrue())
				Expect(t, BeFalse().Match(true)).To(BeFalse())
			})
		})

		describe("BeNumerically", func() {
			it("supports > and <=", func() {
				Expect(t, BeNumerically(">", 0).Match(1)).To(BeTrue())
				Expect(t, BeNumerically(">", 0).Match(0)).To(BeFalse())
				Expect(t, BeNumerically("<=", 5).Match(5)).To(BeTrue())
			})

			context("given an unknown operator", func() {
				it("panics", func() {
					Expect(t, func() { BeNumerically("~=", 0).Match(0) }).To(Panic())
				})
			})
		})

		describe("BeAnExistingFile and BeADirectory", func() {
			var dir string

			before(func() {
				dir, _ = os.MkdirTemp("", "expect-*")
			})

			// after runs explicit teardown, since dir here isn't managed by t.TempDir().
			after(func() {
				os.RemoveAll(dir)
			})

			it("BeAnExistingFile matches a real path", func() {
				Expect(t, BeAnExistingFile().Match(dir)).To(BeTrue())
			})

			it("BeAnExistingFile does not match a missing path", func() {
				Expect(t, BeAnExistingFile().Match("/definitely/does/not/exist")).To(BeFalse())
			})

			it("BeADirectory matches a real directory", func() {
				Expect(t, BeADirectory().Match(dir)).To(BeTrue())
			})

			it("BeADirectory does not match a missing path", func() {
				Expect(t, BeADirectory().Match("/definitely/does/not/exist")).To(BeFalse())
			})
		})

		describe("Panic", func() {
			it("matches a func that panics", func() {
				Expect(t, Panic().Match(func() { panic("boom") })).To(BeTrue())
			})

			it("does not match a func that returns normally", func() {
				Expect(t, Panic().Match(func() {})).To(BeFalse())
			})
		})
	})
}
