package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
)

//go:embed tmpl/ai-monitoring-support.md
var aiMonitoringTmplString string

// RenderAiCompatDoc renders the specified AI Monitoring support JSON
// descriptor file into Markdown text that is readable and understandable (🤞)
// by humans.
func RenderAiCompatDoc(descriptorJsonFile string, writer io.Writer) error {
	reader, err := appFS.Open(descriptorJsonFile)
	if err != nil {
		return fmt.Errorf("could not read descriptor json file: %w", err)
	}

	parsedJson, err := aiCompatReadJson(reader)
	if err != nil {
		return fmt.Errorf("could not parse descriptor json file: %w", err)
	}

	tmpl, err := aiCompatLoadTemplate()
	if err != nil {
		return fmt.Errorf("could not load template: %w", err)
	}

	tmplData := aiCompatBuildTmplData(parsedJson)
	err = tmpl.Execute(writer, tmplData)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return nil
}

func aiCompatReadJson(file io.Reader) (AiCompatJson, error) {
	var result AiCompatJson
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read json from file reader: %w", err)
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func aiCompatBuildTmplData(input AiCompatJson) AiCompatTemplateData {
	result := AiCompatTemplateData{}

	for _, envelope := range input {
		switch strings.ToLower(envelope.Kind) {
		case AiCompatKindGateway:
			result.Bedrock = aiCompatParseBedrockData(envelope)
		case AiCompatKindAbstraction:
			result.Langchain = aiCompatParseLangchainData(envelope)
		case AiCompatKindSdk:
			result.Openai = aiCompatParseOpenaiData(envelope)
		}
	}

	return result
}

func aiCompatLoadTemplate() (*template.Template, error) {
	tmpl := template.New("aiMonitoring")

	tmpl.Funcs(template.FuncMap{
		"boolEmoji": aiCompatBoolEmoji,
	})

	tmpl, err := tmpl.Parse(aiMonitoringTmplString)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func aiCompatParseBedrockData(envelope AiCompatEnvelope) AiCompatBedrockData {
	result := AiCompatBedrockData{Title: envelope.Title}

	result.Models = make([]struct {
		Name  string
		Text  bool
		Image bool
	}, 0)
	for _, model := range envelope.Models {
		modelData := struct {
			Name  string
			Text  bool
			Image bool
		}{}
		modelData.Name = model.Name
		for _, feature := range model.Features {
			if strings.ToLower(feature.Title) == "text" {
				modelData.Text = feature.Supported
			}
			if strings.ToLower(feature.Title) == "Image" {
				modelData.Image = feature.Supported
			}
		}
		result.Models = append(result.Models, modelData)
	}

	return result
}

func aiCompatParseLangchainData(envelope AiCompatEnvelope) AiCompatLangchainData {
	result := AiCompatLangchainData{Title: envelope.Title}

	result.Features = struct {
		Agents       bool
		Chains       bool
		Vectorstores bool
		Tools        bool
	}{}
	for _, feature := range envelope.Features {
		switch strings.ToLower(feature.Title) {
		case "agents":
			result.Features.Agents = feature.Supported
		case "chains":
			result.Features.Chains = feature.Supported
		case "vectorstores":
			result.Features.Vectorstores = feature.Supported
		case "Tools":
			result.Features.Tools = feature.Supported
		}
	}

	result.Providers = make([]struct {
		Name         string
		Supported    bool
		Transitively bool
	}, 0)
	for _, provider := range envelope.Providers {
		providerData := struct {
			Name         string
			Supported    bool
			Transitively bool
		}{provider.Name, provider.Supported, provider.Transitively}
		result.Providers = append(result.Providers, providerData)
	}

	return result
}

func aiCompatParseOpenaiData(envelope AiCompatEnvelope) AiCompatOpenaiData {
	result := AiCompatOpenaiData{}

	for _, feature := range envelope.Features {
		switch strings.ToLower(feature.Title) {
		case "completions":
			result.Completions = feature.Supported
		case "chat":
			result.Chat = feature.Supported
		case "embeddings":
			result.Embeddings = feature.Supported
		case "files":
			result.Files = feature.Supported
		case "images":
			result.Images = feature.Supported
		case "audio":
			result.Audio = feature.Supported
		}
	}

	return result
}

// aiCompatBoolEmoji converts a boolean into emoji text representing the
// respective value. It is added to the AI compat document template as a
// convenience method so that if/else blocks do not need to be repeated
// throughout the template source.
func aiCompatBoolEmoji(input bool) string {
	if input == true {
		return "✅"
	}
	return "❌"
}
