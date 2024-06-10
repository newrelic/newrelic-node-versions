package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_AiCompatJson(t *testing.T) {
	fileData, err := os.ReadFile("testdata/ai-compat.json")
	require.Nil(t, err)

	var parsed AiCompatJson
	err = json.Unmarshal(fileData, &parsed)
	require.Nil(t, err)
	assert.Equal(t, 4, len(parsed))

	expectedGateway := AiCompatEnvelope{
		Kind:     AiCompatKindGateway,
		Title:    "Amazon Bedrock",
		Preamble: "Through the `@aws-sdk/client-bedrock-runtime` module, we support:",
		Footnote: "Note: if a model supports streaming, we also instrument the streaming variant.",
		Models: []AiCompatModel{
			{
				Name: "Claude",
				Features: []AiCompatFeature{
					{"Text", true},
					{"Image", false},
				},
			},
			{
				Name: "Cohere",
				Features: []AiCompatFeature{
					{"Text", true},
					{"Image", false},
				},
			},
		},
	}
	assert.Equal(t, expectedGateway, parsed[0])

	expectedAbstraction := AiCompatEnvelope{
		Kind:              AiCompatKindAbstraction,
		Title:             "Langchain",
		FeaturesPreamble:  "The following general features of Langchain are supported:",
		ProvidersPreamble: "Models/providers are generally supported transitively by our instrumentation of the provider's module.",
		Features: []AiCompatFeature{
			{"Agents", true},
			{"Chains", true},
			{"Vectorstores", true},
			{"Tools", true},
		},
		Providers: []AiCompatProvider{
			{"Azure OpenAI", false, false},
			{"OpenAI", true, true},
		},
	}
	assert.Equal(t, expectedAbstraction, parsed[2])

	expectedSdk := AiCompatEnvelope{
		Kind:  AiCompatKindSdk,
		Title: "OpenAI",
		Features: []AiCompatFeature{
			{"Completions", true},
			{"Chat", true},
			{"Embeddings", true},
			{"Files", false},
			{"Images", false},
			{"Audio", false},
		},
	}
	assert.Equal(t, expectedSdk, parsed[3])
}
