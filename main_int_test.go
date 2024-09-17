package main_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"strings"
	"testing"
)
import "nrversions"

func Test_RenderCompatDoc(t *testing.T) {
	outFilePath := path.Join(os.TempDir(), "compat.md")
	outFile, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	require.Nil(t, err)

	t.Cleanup(func() {
		os.Remove(outFilePath)
	})

	_, err = outFile.WriteString("{/* begin: compat-table */}\n{/* end: compat-table */}")
	require.Nil(t, err)

	args := []string{
		"--no-externals",
		"--repo-dir", ".",
		"--test-dir", "testdata/versioned",
		"--ai-compat-json", "testdata/ai-compat.json",
		"--replace-in-file", outFilePath,
	}
	err = main.Run(args)
	require.Nil(t, err)

	fileData, err := os.ReadFile(outFilePath)
	require.Nil(t, err)
	err = outFile.Close()
	require.Nil(t, err)

	found := string(fileData)

	// Verify that all of the modules are found and their minimum versions
	// are listed correctly. We can't compare against a static document because
	// we can't override private methods in a `_test` package. As a result, this
	// test does hit npmjs.com to find the latest version of each package. So
	// the result will change according to whatever versions are available each
	// time the test is run.
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/client-bedrock-runtime` | 3.474.0"))
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/client-dynamodb` | 3.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/client-sns` | 3.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/client-sqs` | 3.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/lib-dynamodb` | 3.377.0"))
	assert.Equal(t, true, strings.Contains(found, "`@aws-sdk/smithy-client` | 3.47.0"))
	assert.Equal(t, true, strings.Contains(found, "`@elastic/elasticsearch` | 7.16.0"))
	assert.Equal(t, true, strings.Contains(found, "`@koa/router` | 8.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`@langchain/core` | 0.1.17"))
	assert.Equal(t, true, strings.Contains(found, "`@smithy/smithy-client` | 2.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`koa` | 2.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`koa-route` | 3.0.0"))
	assert.Equal(t, true, strings.Contains(found, "`koa-router` | 7.1.0"))
	assert.Equal(t, true, strings.Contains(found, "`mongodb` | 2.1.0"))
}
