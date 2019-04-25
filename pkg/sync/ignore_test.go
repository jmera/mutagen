package sync

import (
	"testing"
)

// TestIgnoreVCSModeUnmarshal tests that unmarshaling from a string
// specification succeeeds for IgnoreVCSMode.
func TestIgnoreVCSModeUnmarshal(t *testing.T) {
	// Set up test cases.
	testCases := []struct {
		Text          string
		ExpectedMode  IgnoreVCSMode
		ExpectFailure bool
	}{
		{"", IgnoreVCSMode_IgnoreVCSModeDefault, true},
		{"asdf", IgnoreVCSMode_IgnoreVCSModeDefault, true},
		{"true", IgnoreVCSMode_IgnoreVCSModeIgnore, false},
		{"false", IgnoreVCSMode_IgnoreVCSModePropagate, false},
	}

	// Process test cases.
	for _, testCase := range testCases {
		var mode IgnoreVCSMode
		if err := mode.UnmarshalText([]byte(testCase.Text)); err != nil {
			if !testCase.ExpectFailure {
				t.Errorf("unable to unmarshal text (%s): %s", testCase.Text, err)
			}
		} else if testCase.ExpectFailure {
			t.Error("unmarshaling succeeded unexpectedly for text:", testCase.Text)
		} else if mode != testCase.ExpectedMode {
			t.Errorf(
				"unmarshaled mode (%s) does not match expected (%s)",
				mode,
				testCase.ExpectedMode,
			)
		}
	}
}

// TestIgnoreVCSModeSupported tests that IgnoreVCSMode support detection works
// as expected.
func TestIgnoreVCSModeSupported(t *testing.T) {
	// Set up test cases.
	testCases := []struct {
		Mode            IgnoreVCSMode
		ExpectSupported bool
	}{
		{IgnoreVCSMode_IgnoreVCSModeDefault, false},
		{IgnoreVCSMode_IgnoreVCSModeIgnore, true},
		{IgnoreVCSMode_IgnoreVCSModePropagate, true},
		{(IgnoreVCSMode_IgnoreVCSModePropagate + 1), false},
	}

	// Process test cases.
	for _, testCase := range testCases {
		if supported := testCase.Mode.Supported(); supported != testCase.ExpectSupported {
			t.Errorf(
				"mode support status (%t) does not match expected (%t)",
				supported,
				testCase.ExpectSupported,
			)
		}
	}
}

// TestIgnoreVCSModeDescription tests that IgnoreVCSMode description generation
// works as expected.
func TestIgnoreVCSModeDescription(t *testing.T) {
	// Set up test cases.
	testCases := []struct {
		Mode                IgnoreVCSMode
		ExpectedDescription string
	}{
		{IgnoreVCSMode_IgnoreVCSModeDefault, "Default"},
		{IgnoreVCSMode_IgnoreVCSModeIgnore, "Ignore"},
		{IgnoreVCSMode_IgnoreVCSModePropagate, "Propagate"},
		{(IgnoreVCSMode_IgnoreVCSModePropagate + 1), "Unknown"},
	}

	// Process test cases.
	for _, testCase := range testCases {
		if description := testCase.Mode.Description(); description != testCase.ExpectedDescription {
			t.Errorf(
				"mode description (%s) does not match expected (%s)",
				description,
				testCase.ExpectedDescription,
			)
		}
	}
}

type ignoreTestValue struct {
	path      string
	directory bool
	expected  bool
}

type ignoreTestCase struct {
	ignores []string
	tests   []ignoreTestValue
}

func (c *ignoreTestCase) run(t *testing.T) {
	// Ensure that all patterns are valid.
	for _, p := range c.ignores {
		if !ValidIgnorePattern(p) {
			t.Fatal("invalid ignore pattern detected:", p)
		}
	}

	// Create an ignorer.
	ignorer, err := newIgnorer(c.ignores)
	if err != nil {
		t.Fatal("unable to create ignorer:", err)
	}

	// Verify test values.
	for _, p := range c.tests {
		if ignorer.ignored(p.path, p.directory) != p.expected {
			t.Error("ignore behavior not as expected for", p.path)
		}
	}
}

func TestIgnoreNone(t *testing.T) {
	test := &ignoreTestCase{
		ignores: nil,
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"something", false, false},
			{"something", true, false},
			{"some/path", false, false},
			{"some/path", true, false},
		},
	}
	test.run(t)
}

func TestIgnorerBasic(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"something",
			"otherthing",
			"!something",
			"somedir/",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"something", false, false},
			{"something", true, false},
			{"subpath/something", false, false},
			{"subpath/something", true, false},
			{"otherthing", false, true},
			{"otherthing", true, true},
			{"subpath/otherthing", false, true},
			{"subpath/otherthing", true, true},
			{"random", false, false},
			{"random", true, false},
			{"subpath/random", false, false},
			{"subpath/random", true, false},
			{"somedir", false, false},
			{"somedir", true, true},
			{"subpath/somedir", false, false},
			{"subpath/somedir", true, true},
		},
	}
	test.run(t)
}

func TestIgnoreGroup(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"*.py[cod]",
			"*.dir[cod]/",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"run.py", false, false},
			{"run.pyc", false, true},
			{"run.pyc", true, true},
			{"subpath/run.pyd", false, true},
			{"subpath/run.pyd", true, true},
			{"run.dir", false, false},
			{"run.dir", true, false},
			{"run.dirc", false, false},
			{"run.dirc", true, true},
			{"subpath/run.dird", false, false},
			{"subpath/run.dird", true, true},
		},
	}
	test.run(t)
}

func TestIgnoreRootRelative(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"/abspath",
			"/absdir/",
			"/name",
			"!*/**/name",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"abspath", false, true},
			{"abspath", true, true},
			{"subpath/abspath", false, false},
			{"subpath/abspath", true, false},
			{"absdir", false, false},
			{"absdir", true, true},
			{"subpath/absdir", false, false},
			{"subpath/absdir", true, false},
			{"name", false, true},
			{"name", true, true},
			{"subpath/name", false, false},
			{"subpath/name", true, false},
		},
	}
	test.run(t)
}

func TestIgnoreDoublestar(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"some/*",
			"some/**/*",
			"!some/other",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"something", false, false},
			{"some", false, false},
			{"some/path", false, true},
			{"some/other", false, false},
			{"some/other/path", false, true},
		},
	}
	test.run(t)
}

func TestIgnoreNegateOrdering(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"!something",
			"otherthing",
			"something",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"something", false, true},
			{"something/other", false, false},
			{"otherthing", false, true},
			{"some/path", false, false},
		},
	}
	test.run(t)
}

func TestIgnoreWildcard(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"some*",
			"!someone",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"som", false, false},
			{"some", false, true},
			{"something", false, true},
			{"someone", false, false},
			{"some/path", false, false},
		},
	}
	test.run(t)
}

func TestIgnorePathWildcard(t *testing.T) {
	test := &ignoreTestCase{
		ignores: []string{
			"some/*",
			"some/**/*",
			"!some/other",
		},
		tests: []ignoreTestValue{
			{"", false, false},
			{"", true, false},
			{"something", false, false},
			{"some", false, false},
			{"some/path", false, true},
			{"some/other", false, false},
			{"some/other/path", false, true},
			{"subdir/some/other/path", false, false},
		},
	}
	test.run(t)
}

func TestIgnoreEmptyPatternsInvalid(t *testing.T) {
	if ValidIgnorePattern("") {
		t.Error("empty pattern should be invalid")
	}
	if ValidIgnorePattern("!") {
		t.Error("negated empty pattern should be invalid")
	}
	if ValidIgnorePattern("/") {
		t.Error("root pattern should be invalid")
	}
	if ValidIgnorePattern("!/") {
		t.Error("negated root pattern should be invalid")
	}
	if ValidIgnorePattern("//") {
		t.Error("root directory pattern should be invalid")
	}
	if ValidIgnorePattern("!//") {
		t.Error("negated root directory pattern should be invalid")
	}
}

func TestIgnoreInvalidPatternInvalid(t *testing.T) {
	if ValidIgnorePattern("\\") {
		t.Fatal("invalid pattern should be invalid")
	}
}

func TestIgnoreInvalidPatternOnIgnorerConstruction(t *testing.T) {
	if ignorer, err := newIgnorer([]string{"\\"}); err == nil {
		t.Error("ignorer creation should fail on invalid pattern")
	} else if ignorer != nil {
		t.Error("ignorer should be nil on failed creation")
	}
}
