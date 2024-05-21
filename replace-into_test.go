package main

import (
	"bytes"
	"errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

// errorWriter is an [io.Writer] that always returns an error
// when attempting to write data.
type errorWriter struct{}

func (ew *errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

func Test_ReplaceInFile(t *testing.T) {
	origFS := appFS
	t.Cleanup(func() {
		appFS = origFS
	})

	inputFileContent, e := os.ReadFile("testdata/replace-into.input.md")
	require.Nil(t, e)
	expectedFileContent, e := os.ReadFile("testdata/replace-into.expected.md")
	require.Nil(t, e)

	t.Run("returns error if file cannot be opened", func(t *testing.T) {
		appFS = afero.NewMemMapFs()
		err := ReplaceInFile("foo.md", "bar", "a", "b")
		assert.ErrorContains(t, err, "file does not exist")
	})

	t.Run("replaces as expected", func(t *testing.T) {
		appFS = afero.NewMemMapFs()
		err := afero.WriteFile(appFS, "foo.md", inputFileContent, 0o777)
		require.Nil(t, err)
		appFS.Chmod("foo.md", os.ModePerm)

		err = ReplaceInFile(
			"foo.md",
			"\nfoobar\n",
			"{/* begin: compat-table */}",
			"{/* end: compat-table */}",
		)
		assert.Nil(t, err)

		found, err := afero.ReadFile(appFS, "foo.md")
		assert.Equal(t, expectedFileContent, found)
	})
}

func Test_getParts(t *testing.T) {
	t.Run("returns error for bad read", func(t *testing.T) {
		found, err := getParts(&errorReader{}, "a", "b")
		assert.Nil(t, found)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("returns error if cannot find start marker", func(t *testing.T) {
		reader := strings.NewReader("foo bar")
		found, err := getParts(reader, "~~~", "---")
		assert.Nil(t, found)
		assert.ErrorContains(t, err, "unable to find start marker: `~~~`")
	})

	t.Run("returns error if cannot find end marker", func(t *testing.T) {
		reader := strings.NewReader("foo ~~~ bar")
		found, err := getParts(reader, "~~~", "---")
		assert.Nil(t, found)
		assert.ErrorContains(t, err, "unable to find end marker: `---`")
	})

	t.Run("splits into parts correctly", func(t *testing.T) {
		doc := []string{
			"header line 1",
			"",
			"header line 3",
			"",
			"{start_marker}",
			"replace me",
			"",
			"{end_marker}",
			"",
			"tail line 1",
		}
		input := strings.Join(doc, "\n")
		expected := &fileParts{
			// TODO: there's a bug in strings.Join that is skipping the final ""
			//head: []byte(strings.Join(doc[0:4], "\n")),
			head: []byte("header line 1\n\nheader line 3\n\n"),
			//tail: []byte(strings.Join(doc[8:], "\n")),
			tail: []byte("\n\ntail line 1"),
		}

		found, err := getParts(strings.NewReader(input), "{start_marker}", "{end_marker}")
		assert.Nil(t, err)
		assert.Equal(t, expected, found)
	})
}

func Test_writeDoc(t *testing.T) {
	t.Run("returns error for bad write", func(t *testing.T) {
		writer := &errorWriter{}
		err := writeDoc(writer, [][]byte{})
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("writes a doc", func(t *testing.T) {
		writer := &bytes.Buffer{}
		input := [][]byte{
			[]byte("foo"),
			[]byte("bar"),
		}
		expected := []byte("foobar")
		err := writeDoc(writer, input)
		assert.Nil(t, err)
		assert.Equal(t, expected, writer.Bytes())
	})
}
