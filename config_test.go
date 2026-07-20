package expect_test

import (
	"testing"

	. "github.com/woodie/expect"
)

// expect is the recommended lowercase call-site alias: a one-line generic
// pass-through, declared locally per test package rather than exported from
// expect itself, since Go only allows a lowercase identifier to skip the
// dot-import's capitalization requirement within its own package.
func expect[T any](got T, t testing.TB) Expectation[T] { return Expect(got, t) }

// Improve readability with vars set for structural functions and lifecycle hooks
/*
func TestCalculator(t *testing.T) {
	spec.Run(t, "Calculator", func(t *testing.T, describe spec.G, it spec.S) {
		context, before, after := describe, it.Before, it.After // HERE

		var calculator *Calculator
		before(func() { calculator = NewCalculator() })
		after(func() { calculator.Clear() })

		describe("#add", func() {
			context("with positive numbers", func() {
				it("returns the correct sum", func() {
					expect(calculator.Add(2, 3), t).To(Equal(5))
				})
			})
		})
	})
} */
