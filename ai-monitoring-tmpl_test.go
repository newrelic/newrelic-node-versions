package main

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_RenderAiCompatDoc(t *testing.T) {
	t.Run("successfully renders document", func(t *testing.T) {
		expected, err := os.ReadFile("testdata/ai-compat.expected.md")
		require.Nil(t, err)

		builder := strings.Builder{}
		err = RenderAiCompatDoc("testdata/ai-compat.json", &builder)
		assert.Nil(t, err)

		assert.Equal(t, string(expected), builder.String())
	})

	t.Run("errors if cannot open json file", func(t *testing.T) {
		err := RenderAiCompatDoc("testdata/bad-file.does.not.exist", io.Discard)
		assert.ErrorContains(t, err, "could not read descriptor json file")
	})

	t.Run("errors if json is bad", func(t *testing.T) {
		currFS := appFS
		t.Cleanup(func() {
			appFS = currFS
		})

		appFS = afero.NewMemMapFs()
		file, err := appFS.Create("foo.json")
		require.Nil(t, err)
		io.WriteString(file, `{"bad":json`)

		err = RenderAiCompatDoc("foo.json", io.Discard)
		assert.ErrorContains(t, err, "could not parse descriptor json file")
	})

	t.Run("errors if template is invalid", func(t *testing.T) {
		curTmplString := aiMonitoringTmplString
		t.Cleanup(func() {
			aiMonitoringTmplString = curTmplString
		})
		aiMonitoringTmplString = "{{bad}"

		err := RenderAiCompatDoc("testdata/ai-compat.json", io.Discard)
		assert.ErrorContains(t, err, "could not load template")
	})

	t.Run("errors if rendering fails", func(t *testing.T) {
		curTmplString := aiMonitoringTmplString
		t.Cleanup(func() {
			aiMonitoringTmplString = curTmplString
		})
		aiMonitoringTmplString = "{{.Text}}"

		err := RenderAiCompatDoc("testdata/ai-compat.json", io.Discard)
		assert.ErrorContains(t, err, "failed to render template")
	})
}
