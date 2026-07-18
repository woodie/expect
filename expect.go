// Package expect is a small, dependency-free, generics-based matcher library for the standard testing package.
package expect

import (
	"cmp"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Matcher checks got and describes the expectation it failed to meet.
type Matcher[T any] interface {
	Match(got T) bool
	String() string
}

// Expectation wraps a value under test, obtained via Expect.
type Expectation[T any] struct {
	t   testing.TB
	got T
}

// Expect begins an assertion against got -- dot-import this package so call sites read Expect(t, x).To(Equal(y)).
func Expect[T any](t testing.TB, got T) Expectation[T] {
	t.Helper()
	return Expectation[T]{t: t, got: got}
}

// To fails the test via t.Errorf if m does not match.
func (e Expectation[T]) To(m Matcher[T]) {
	e.t.Helper()
	if !m.Match(e.got) {
		e.t.Errorf("got %v, want to %s", e.got, m.String())
	}
}

// NotTo fails the test via t.Errorf if m matches.
func (e Expectation[T]) NotTo(m Matcher[T]) {
	e.t.Helper()
	if m.Match(e.got) {
		e.t.Errorf("got %v, want to not %s", e.got, m.String())
	}
}

// ToNot is an alias for NotTo, matching Gomega's own To/ToNot/NotTo naming exactly.
func (e Expectation[T]) ToNot(m Matcher[T]) {
	e.t.Helper()
	e.NotTo(m)
}

type equalMatcher[T comparable] struct{ want T }

func (m equalMatcher[T]) Match(got T) bool { return got == m.want }
func (m equalMatcher[T]) String() string   { return fmt.Sprintf("equal %v", m.want) }

// Equal matches got == want.
func Equal[T comparable](want T) Matcher[T] { return equalMatcher[T]{want} }

type deepEqualMatcher[T any] struct{ want T }

func (m deepEqualMatcher[T]) Match(got T) bool { return reflect.DeepEqual(got, m.want) }
func (m deepEqualMatcher[T]) String() string   { return fmt.Sprintf("deep-equal %v", m.want) }

// DeepEqual matches reflect.DeepEqual(got, want) -- for slices, maps, and structs, which Equal can't compare.
func DeepEqual[T any](want T) Matcher[T] { return deepEqualMatcher[T]{want} }

type identicalMatcher[T comparable] struct{ want T }

func (m identicalMatcher[T]) Match(got T) bool { return got == m.want }
func (m identicalMatcher[T]) String() string   { return fmt.Sprintf("be identical to %v", m.want) }

// BeIdenticalTo matches got == want -- named separately from Equal for pointer/interface identity checks.
func BeIdenticalTo[T comparable](want T) Matcher[T] { return identicalMatcher[T]{want} }

type containMatcher struct{ substr string }

func (m containMatcher) Match(got string) bool { return strings.Contains(got, m.substr) }
func (m containMatcher) String() string        { return fmt.Sprintf("contain %q", m.substr) }

// Contain matches strings.Contains(got, substr).
func Contain(substr string) Matcher[string] { return containMatcher{substr} }

type succeedMatcher struct{}

func (succeedMatcher) Match(got error) bool { return got == nil }
func (succeedMatcher) String() string       { return "succeed (nil error)" }

// Succeed matches a nil error.
func Succeed() Matcher[error] { return succeedMatcher{} }

type occurredMatcher struct{}

func (occurredMatcher) Match(got error) bool { return got != nil }
func (occurredMatcher) String() string       { return "have occurred (non-nil error)" }

// HaveOccurred matches a non-nil error.
func HaveOccurred() Matcher[error] { return occurredMatcher{} }

type trueMatcher struct{}

func (trueMatcher) Match(got bool) bool { return got }
func (trueMatcher) String() string      { return "be true" }

// BeTrue matches got == true.
func BeTrue() Matcher[bool] { return trueMatcher{} }

type falseMatcher struct{}

func (falseMatcher) Match(got bool) bool { return !got }
func (falseMatcher) String() string      { return "be false" }

// BeFalse matches got == false.
func BeFalse() Matcher[bool] { return falseMatcher{} }

type numericMatcher[T cmp.Ordered] struct {
	op   string
	want T
}

func (m numericMatcher[T]) Match(got T) bool {
	switch m.op {
	case "==":
		return got == m.want
	case "!=":
		return got != m.want
	case ">":
		return got > m.want
	case ">=":
		return got >= m.want
	case "<":
		return got < m.want
	case "<=":
		return got <= m.want
	default:
		panic("expect: unknown BeNumerically operator " + m.op)
	}
}
func (m numericMatcher[T]) String() string { return fmt.Sprintf("be %s %v", m.op, m.want) }

// BeNumerically matches got against want using op ("==", "!=", ">", ">=", "<", "<=").
func BeNumerically[T cmp.Ordered](op string, want T) Matcher[T] {
	return numericMatcher[T]{op: op, want: want}
}

type existingFileMatcher struct{}

func (existingFileMatcher) Match(got string) bool {
	_, err := os.Stat(got)
	return err == nil
}
func (existingFileMatcher) String() string { return "be an existing file" }

// BeAnExistingFile matches a path that os.Stat can resolve.
func BeAnExistingFile() Matcher[string] { return existingFileMatcher{} }

type directoryMatcher struct{}

func (directoryMatcher) Match(got string) bool {
	info, err := os.Stat(got)
	return err == nil && info.IsDir()
}
func (directoryMatcher) String() string { return "be a directory" }

// BeADirectory matches a path that os.Stat resolves to a directory.
func BeADirectory() Matcher[string] { return directoryMatcher{} }

type panicMatcher struct{}

func (panicMatcher) Match(got func()) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	got()
	return false
}
func (panicMatcher) String() string { return "panic" }

// Panic matches a func() that panics when called.
func Panic() Matcher[func()] { return panicMatcher{} }
