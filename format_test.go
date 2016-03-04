package lorg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// placeholderFabric can fabricate a function which has same signature
// as Placeholder, but fabricated function logs entry about running
// placeholder to fabric's log and returns same entry.
type placeholderFabric struct {
	log []string
}

func (fabric *placeholderFabric) fabricate(name string) Placeholder {
	// create a function and pass placeholder name by reference
	return func(_ Level, value string) string {
		message := fmt.Sprintf("[%s@%s]", name, value)
		fabric.log = append(fabric.log, message)
		return message
	}
}

func TestNewFormat_ReturnsFormatWithDefaultFiends(t *testing.T) {
	format := NewFormat(``)

	assert.Len(t, format.placeholders, len(defaultPlaceholders))

	keys := []string{}
	for key := range defaultPlaceholders {
		keys = append(keys, key)
	}

	for _, key := range keys {
		assert.Contains(t, format.placeholders, key)
	}
}

func TestFormat_ImplementsFormatterInterface(t *testing.T) {
	assert.Implements(t, (*Formatter)(nil), &Format{})
}

func TestFormat_GetPlaceholders_ReturnsPlaceholdersField(t *testing.T) {
	rawFormat := `${place_foo} ${place_bar:arg} ${unknown}`
	format := NewFormat(rawFormat)

	placeholders := map[string]Placeholder{
		"place_foo": func(_ Level, _ string) string { return "holder foo" },
		"place_bar": func(_ Level, _ string) string { return "holder bar" },
	}

	format.placeholders = placeholders

	assert.Equal(t, placeholders, format.GetPlaceholders())
}

func TestFormat_SetPlaceholders_ChangesPlaceholdersField(t *testing.T) {
	// https://golang.org/doc/devel/weekly.html#2011-11-18
	//
	// Map and function value comparisons are now disallowed (except for
	// comparison with nil) as per the Go 1 plan. Function equality was
	// problematic in some contexts and map equality compares pointers, not the
	// maps' content.
	//
	// Okay, go can't compare functions, so we can't test that SetPlaceholder
	// sets exactly specified Placeholder into placeholders[key], we can check
	// only key existing.
	format := NewFormat(``)

	placeholders := map[string]Placeholder{
		"place_foo": func(_ Level, _ string) string { return "foo" },
	}

	format.SetPlaceholders(placeholders)
	assert.Contains(t, format.placeholders, "place_foo")

	anotherPlaceholders := map[string]Placeholder{
		"place_bar": func(_ Level, _ string) string { return "bar" },
	}

	format.SetPlaceholders(anotherPlaceholders)
	assert.NotContains(t, format.placeholders, "place_foo")
	assert.Contains(t, format.placeholders, "place_bar")
}

func TestFormat_SetPlaceholder_ChangesPlaceholdersField(t *testing.T) {
	// please see comment about testing SetPlaceholders function:
	// TestFormat_SetPlaceholders_ChangesPlaceholdersField
	format := NewFormat(``)

	placeholderFoo := func(_ Level, _ string) string { return "foo" }
	placeholderBar := func(_ Level, _ string) string { return "bar" }

	format.SetPlaceholder("place_foo", placeholderFoo)

	assert.Len(t, format.placeholders, len(defaultPlaceholders)+1)
	assert.Contains(t, format.placeholders, "place_foo")

	format.SetPlaceholder("place_bar", placeholderBar)

	assert.Len(t, format.placeholders, len(defaultPlaceholders)+2)
	assert.Contains(t, format.placeholders, "place_foo")
	assert.Contains(t, format.placeholders, "place_bar")
}

func TestFormat_Render_UsesPlaceholdersFieldForMatchingPlaceholders(
	t *testing.T,
) {
	rawFormat := `plain text ${place_foo} foo ${place_bar:barvalue}`
	format := NewFormat(rawFormat)

	placeholders := map[string]Placeholder{
		"place_foo": func(_ Level, _ string) string { return "foo" },
	}

	format.SetPlaceholders(placeholders)
	format.Render(LevelWarning)

	assert.Equal(
		t,
		[]string{`${place_foo}`},
		getReplacementsValues(format.replacements),
	)

	format.SetPlaceholders(
		map[string]Placeholder{
			"place_foo": func(_ Level, _ string) string { return "foo" },
			"place_bar": func(_ Level, _ string) string { return "bar" },
		},
	)

	format.Reset()

	format.Render(LevelWarning)

	assert.Equal(
		t,
		[]string{`${place_foo}`, `${place_bar:barvalue}`},
		getReplacementsValues(format.replacements),
	)
}

func TestFormat_Render_PlaceholderRegexpMatching(t *testing.T) {
	type testcase struct {
		format               string
		expectedReplacements []string
	}

	testcases := []testcase{
		{
			`text`,
			[]string{},
		},
		{
			`${place_foo}`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`before${place_foo}after`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`before space ${place_foo} space after`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`before${place_foo}`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`before space ${place_foo}`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`${place_foo}after`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`${place_foo} space after`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`${place_foo} ${unknown}`,
			[]string{
				`${place_foo}`,
			},
		},
		{
			`${place_foo} ${place_foo:1}`,
			[]string{
				`${place_foo}`, `${place_foo:1}`,
			},
		},
		{
			`${place_foo} ${place_foo:1:2}`,
			[]string{
				`${place_foo}`, `${place_foo:1:2}`,
			},
		},
		{
			`${place_foo} ${unknown} ${place_foo:1}`,
			[]string{
				`${place_foo}`, `${place_foo:1}`,
			},
		},
		{
			`${unknown}${place_foo} ${place_foo:1}after`,
			[]string{
				`${place_foo}`, `${place_foo:1}`,
			},
		},
		{
			`${place_foo} ${place_foo:1} ${place_foo}`,
			[]string{
				`${place_foo}`, `${place_foo:1}`, `${place_foo}`,
			},
		},
		{
			`${place_foo} ${place_foo:1} ${place_foo:1:2:3:4}`,
			[]string{
				`${place_foo}`, `${place_foo:1}`, `${place_foo:1:2:3:4}`,
			},
		},
		{
			`${place_foo} ${place_foo:1} ${place_bar}}`,
			[]string{
				`${place_foo}`, `${place_foo:1}`, `${place_bar}`,
			},
		},
	}

	placeholders := map[string]Placeholder{
		"place_foo": func(_ Level, _ string) string { return "holder foo" },
		"place_bar": func(_ Level, _ string) string { return "holder bar" },
	}

	for _, testcase := range testcases {
		format := NewFormat(testcase.format)
		format.SetPlaceholders(placeholders)
		format.Render(LevelError)

		assert.Equal(
			t,
			testcase.expectedReplacements,
			getReplacementsValues(format.replacements),
			"format: %s", testcase.format,
		)
	}
}

func TestFormat_Reset_DesolatesReplacementAndUnsetsCompiledFlag(
	t *testing.T,
) {
	rawFormat := `plain text ${place_foo} foo ${place_bar:barvalue}`
	format := NewFormat(rawFormat)

	format.SetPlaceholders(
		map[string]Placeholder{
			"place_foo": func(_ Level, _ string) string { return "foo" },
		},
	)

	format.Render(LevelWarning)

	assert.NotEmpty(t, format.replacements)
	assert.True(t, format.compiled)

	format.Reset()

	assert.Empty(t, format.replacements)
	assert.False(t, format.compiled)
}

func TestFormat_Render_CallsSettedPlaceholders(t *testing.T) {
	fabric := new(placeholderFabric)

	format := NewFormat(
		`fmt: ${place_foo} ${place_foo:1} ${place_bar:a b c:d e f}`,
	)

	format.SetPlaceholders(
		map[string]Placeholder{
			"place_foo": fabric.fabricate("place_foo"),
			"place_bar": fabric.fabricate("place_bar"),
		},
	)

	rendered := format.Render(LevelWarning)

	assert.Equal(
		t,
		"fmt: [place_foo@] [place_foo@1] [place_bar@a b c:d e f]",
		rendered,
	)

	expectedFabricLog := []string{
		"[place_foo@]",
		"[place_foo@1]",
		"[place_bar@a b c:d e f]",
	}

	assert.Equal(
		t, expectedFabricLog, fabric.log,
		"Format runs placeholders in wrong order",
	)

	// now let's fake placeholder for place_bar and check that placeholder
	// was launched, and old place_bar placeholder wasn't launched
	format.SetPlaceholder("place_bar", fabric.fabricate("fakebar"))
	format.Reset()

	rendered = format.Render(LevelWarning)

	assert.Equal(
		t,
		"fmt: [place_foo@] [place_foo@1] [fakebar@a b c:d e f]",
		rendered,
	)

	expectedFabricLog = append(
		expectedFabricLog,
		"[place_foo@]",
		"[place_foo@1]",
		"[fakebar@a b c:d e f]",
	)

	assert.Equal(
		t, expectedFabricLog, fabric.log,
		"Format runs placeholders in wrong order",
	)
}

func TestFormat_Render_CallsSettedPlaceholdersAndPassesLogLevel(t *testing.T) {
	format := NewFormat(`${place_foo}`)

	var placeholderLogLevel Level = -1

	format.SetPlaceholder(
		"place_foo",
		func(logLevel Level, _ string) string {
			placeholderLogLevel = logLevel
			return ""
		},
	)

	format.Render(LevelDebug)

	assert.Equal(
		t, LevelDebug, placeholderLogLevel,
		"log level doesn't passed to placeholder",
	)
}
