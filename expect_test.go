package expect_test

import (
	"errors"
	"testing"

	"github.com/woodie/expect"
)

func TestToPassesAndFails(t *testing.T) {
	passed := t.Run("passing", func(t *testing.T) {
		expect.That(t, 2+2).To(expect.Equal(4))
	})
	if !passed {
		t.Fatal("expected a matching To to pass")
	}

	passed = t.Run("failing", func(t *testing.T) {
		expect.That(t, 2+2).To(expect.Equal(5))
	})
	if passed {
		t.Fatal("expected a mismatched To to fail")
	}
}

func TestNotToPassesAndFails(t *testing.T) {
	passed := t.Run("passing", func(t *testing.T) {
		expect.That(t, 2+2).NotTo(expect.Equal(5))
	})
	if !passed {
		t.Fatal("expected a mismatched NotTo to pass")
	}

	passed = t.Run("failing", func(t *testing.T) {
		expect.That(t, 2+2).NotTo(expect.Equal(4))
	})
	if passed {
		t.Fatal("expected a matching NotTo to fail")
	}
}

func TestEqual(t *testing.T) {
	if !expect.Equal(5).Match(5) {
		t.Error("Equal(5).Match(5) = false")
	}
	if expect.Equal(5).Match(6) {
		t.Error("Equal(5).Match(6) = true")
	}
}

func TestDeepEqual(t *testing.T) {
	want := []string{"a", "b"}
	if !expect.DeepEqual(want).Match([]string{"a", "b"}) {
		t.Error("DeepEqual matched slices with equal contents as unequal")
	}
	if expect.DeepEqual(want).Match([]string{"a", "c"}) {
		t.Error("DeepEqual matched slices with different contents as equal")
	}
}

func TestContain(t *testing.T) {
	if !expect.Contain("world").Match("hello world") {
		t.Error("Contain(\"world\") should match \"hello world\"")
	}
	if expect.Contain("world").Match("hello") {
		t.Error("Contain(\"world\") should not match \"hello\"")
	}
}

func TestSucceedAndHaveOccurred(t *testing.T) {
	if !expect.Succeed().Match(nil) {
		t.Error("Succeed() should match nil")
	}
	if expect.Succeed().Match(errors.New("boom")) {
		t.Error("Succeed() should not match a non-nil error")
	}
	if expect.HaveOccurred().Match(nil) {
		t.Error("HaveOccurred() should not match nil")
	}
	if !expect.HaveOccurred().Match(errors.New("boom")) {
		t.Error("HaveOccurred() should match a non-nil error")
	}
}

func TestBeTrueAndBeFalse(t *testing.T) {
	if !expect.BeTrue().Match(true) || expect.BeTrue().Match(false) {
		t.Error("BeTrue() matched incorrectly")
	}
	if !expect.BeFalse().Match(false) || expect.BeFalse().Match(true) {
		t.Error("BeFalse() matched incorrectly")
	}
}

func TestBeNumerically(t *testing.T) {
	if !expect.BeNumerically(">", 0).Match(1) {
		t.Error(`BeNumerically(">", 0) should match 1`)
	}
	if expect.BeNumerically(">", 0).Match(0) {
		t.Error(`BeNumerically(">", 0) should not match 0`)
	}
	if !expect.BeNumerically("<=", 5).Match(5) {
		t.Error(`BeNumerically("<=", 5) should match 5`)
	}
}

func TestBeNumericallyPanicsOnUnknownOp(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected an unknown operator to panic")
		}
	}()
	expect.BeNumerically("~=", 0).Match(0)
}

func TestBeAnExistingFile(t *testing.T) {
	if expect.BeAnExistingFile().Match("/definitely/does/not/exist") {
		t.Error("BeAnExistingFile() matched a nonexistent path")
	}
	if !expect.BeAnExistingFile().Match(t.TempDir()) {
		t.Error("BeAnExistingFile() should match an existing directory")
	}
}

func TestBeADirectory(t *testing.T) {
	if !expect.BeADirectory().Match(t.TempDir()) {
		t.Error("BeADirectory() should match a real directory")
	}
	if expect.BeADirectory().Match("/definitely/does/not/exist") {
		t.Error("BeADirectory() matched a nonexistent path")
	}
}

func TestPanic(t *testing.T) {
	if !expect.Panic().Match(func() { panic("boom") }) {
		t.Error("Panic() should match a func that panics")
	}
	if expect.Panic().Match(func() {}) {
		t.Error("Panic() should not match a func that returns normally")
	}
}
