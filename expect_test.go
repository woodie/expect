package expect_test

import (
	"errors"
	"testing"

	. "github.com/woodie/expect"
)

type spyT struct {
	testing.TB
	failed bool
}

func (s *spyT) Helper() {}

func (s *spyT) Errorf(format string, args ...interface{}) {
	s.failed = true
}

func TestExpectToPassesAndFails(t *testing.T) {
	pass := &spyT{}
	Expect(pass, 2+2).To(Equal(4))
	if pass.failed {
		t.Error("expected a matching To to pass")
	}

	fail := &spyT{}
	Expect(fail, 2+2).To(Equal(5))
	if !fail.failed {
		t.Error("expected a mismatched To to fail")
	}
}

func TestNotToPassesAndFails(t *testing.T) {
	pass := &spyT{}
	Expect(pass, 2+2).NotTo(Equal(5))
	if pass.failed {
		t.Error("expected a mismatched NotTo to pass")
	}

	fail := &spyT{}
	Expect(fail, 2+2).NotTo(Equal(4))
	if !fail.failed {
		t.Error("expected a matching NotTo to fail")
	}
}

func TestToNotIsAnAliasForNotTo(t *testing.T) {
	pass := &spyT{}
	Expect(pass, 2+2).ToNot(Equal(5))
	if pass.failed {
		t.Error("expected a mismatched ToNot to pass")
	}

	fail := &spyT{}
	Expect(fail, 2+2).ToNot(Equal(4))
	if !fail.failed {
		t.Error("expected a matching ToNot to fail")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(5).Match(5) {
		t.Error("Equal(5).Match(5) = false")
	}
	if Equal(5).Match(6) {
		t.Error("Equal(5).Match(6) = true")
	}
}

func TestDeepEqual(t *testing.T) {
	want := []string{"a", "b"}
	if !DeepEqual(want).Match([]string{"a", "b"}) {
		t.Error("DeepEqual matched slices with equal contents as unequal")
	}
	if DeepEqual(want).Match([]string{"a", "c"}) {
		t.Error("DeepEqual matched slices with different contents as equal")
	}
}

func TestContain(t *testing.T) {
	if !Contain("world").Match("hello world") {
		t.Error("Contain(\"world\") should match \"hello world\"")
	}
	if Contain("world").Match("hello") {
		t.Error("Contain(\"world\") should not match \"hello\"")
	}
}

func TestSucceedAndHaveOccurred(t *testing.T) {
	if !Succeed().Match(nil) {
		t.Error("Succeed() should match nil")
	}
	if Succeed().Match(errors.New("boom")) {
		t.Error("Succeed() should not match a non-nil error")
	}
	if HaveOccurred().Match(nil) {
		t.Error("HaveOccurred() should not match nil")
	}
	if !HaveOccurred().Match(errors.New("boom")) {
		t.Error("HaveOccurred() should match a non-nil error")
	}
}

func TestBeTrueAndBeFalse(t *testing.T) {
	if !BeTrue().Match(true) || BeTrue().Match(false) {
		t.Error("BeTrue() matched incorrectly")
	}
	if !BeFalse().Match(false) || BeFalse().Match(true) {
		t.Error("BeFalse() matched incorrectly")
	}
}

func TestBeNumerically(t *testing.T) {
	if !BeNumerically(">", 0).Match(1) {
		t.Error(`BeNumerically(">", 0) should match 1`)
	}
	if BeNumerically(">", 0).Match(0) {
		t.Error(`BeNumerically(">", 0) should not match 0`)
	}
	if !BeNumerically("<=", 5).Match(5) {
		t.Error(`BeNumerically("<=", 5) should match 5`)
	}
}

func TestBeNumericallyPanicsOnUnknownOp(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected an unknown operator to panic")
		}
	}()
	BeNumerically("~=", 0).Match(0)
}

func TestBeAnExistingFile(t *testing.T) {
	if BeAnExistingFile().Match("/definitely/does/not/exist") {
		t.Error("BeAnExistingFile() matched a nonexistent path")
	}
	if !BeAnExistingFile().Match(t.TempDir()) {
		t.Error("BeAnExistingFile() should match an existing directory")
	}
}

func TestBeADirectory(t *testing.T) {
	if !BeADirectory().Match(t.TempDir()) {
		t.Error("BeADirectory() should match a real directory")
	}
	if BeADirectory().Match("/definitely/does/not/exist") {
		t.Error("BeADirectory() matched a nonexistent path")
	}
}

func TestPanic(t *testing.T) {
	if !Panic().Match(func() { panic("boom") }) {
		t.Error("Panic() should match a func that panics")
	}
	if Panic().Match(func() {}) {
		t.Error("Panic() should not match a func that returns normally")
	}
}
