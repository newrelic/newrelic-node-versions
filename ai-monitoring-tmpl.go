package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
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
	result := AiCompatTemplateData{
		Gateways: make([]AiCompatGateway, 0),
	}

	for _, envelope := range input {
		switch strings.ToLower(envelope.Kind) {
		case AiCompatKindGateway:
			gateway := aiCompatParseGateway(envelope)
			result.Gateways = append(result.Gateways, gateway)
		case AiCompatKindAbstraction:
			result.Langchain = aiCompatParseLangchainData(envelope)
		case AiCompatKindSdk:
			result.Openai = aiCompatParseOpenaiData(envelope)
		}
	}

	return result
}

func aiCompatParseGateway(input AiCompatEnvelope) AiCompatGateway {
	result := AiCompatGateway{
		Title:    input.Title,
		Preamble: input.Preamble,
		Footnote: input.Footnote,
	}

	models := make([]AiCompatGatewayModel, 0)
	for _, model := range input.Models {
		models = append(models, AiCompatGatewayModel{model.Name, model.Features})
	}
	result.Models = models

	return result
}

func aiCompatLoadTemplate() (*template.Template, error) {
	tmpl := template.New("aiMonitoring")

	tmpl.Funcs(template.FuncMap{
		"boolEmoji":            aiCompatBoolEmoji,
		"gatewayModelsToTable": aiModelsToTable,
	})

	tmpl, err := tmpl.Parse(aiMonitoringTmplString)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
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

func aiModelsToTable(input []AiCompatGatewayModel) string {
	result := strings.Builder{}

	featureTitles := make([]string, 0)
	for _, val := range input[0].Features {
		featureTitles = append(featureTitles, val.Title)
	}
	slices.Sort(featureTitles)

	header := "| Model |"
	separator := "| --- |"
	for _, title := range featureTitles {
		header = fmt.Sprintf("%s %s |", header, title)
		separator = fmt.Sprintf("%s --- |", separator)
	}
	result.WriteString(header + "\n")
	result.WriteString(separator + "\n")

	for _, model := range input {
		row := fmt.Sprintf("| %s |", model.Title)
		slices.SortFunc(model.Features, func(a AiCompatFeature, b AiCompatFeature) int {
			if a.Title < b.Title {
				return -1
			}
			return 1
		})
		for _, feature := range model.Features {
			row = fmt.Sprintf("%s %s |", row, aiCompatBoolEmoji(feature.Supported))
		}
		result.WriteString(row + "\n")
	}

	return strings.TrimSpace(result.String())
}
