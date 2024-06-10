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
		Abstractions: make([]AiCompatAbstraction, 0),
		Gateways:     make([]AiCompatGateway, 0),
	}

	for _, envelope := range input {
		switch strings.ToLower(envelope.Kind) {
		case AiCompatKindGateway:
			gateway := aiCompatParseGateway(envelope)
			result.Gateways = append(result.Gateways, gateway)
		case AiCompatKindAbstraction:
			abstraction := aiCompatParseAbstraction(envelope)
			result.Abstractions = append(result.Abstractions, abstraction)
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

func aiCompatParseAbstraction(input AiCompatEnvelope) AiCompatAbstraction {
	return AiCompatAbstraction{
		Title:             input.Title,
		FeaturesPreamble:  input.FeaturesPreamble,
		ProvidersPreamble: input.ProvidersPreamble,
		Features:          input.Features,
		Providers:         input.Providers,
	}
}

func aiCompatLoadTemplate() (*template.Template, error) {
	tmpl := template.New("aiMonitoring")

	tmpl.Funcs(template.FuncMap{
		"boolEmoji":            aiCompatBoolEmoji,
		"featuresToTable":      aiFeaturesToTable,
		"gatewayModelsToTable": aiModelsToTable,
		"providersToTable":     aiProvidersToTable,
	})

	tmpl, err := tmpl.Parse(aiMonitoringTmplString)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
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
// respective value.
func aiCompatBoolEmoji(input bool) string {
	if input == true {
		return "✅"
	}
	return "❌"
}

// aiModelsToTable renders a set of gateway objects into a Markdown table.
// It is added to the AI Monitoring template as a convenience function.
func aiModelsToTable(input []AiCompatGatewayModel) string {
	result := strings.Builder{}

	featureTitles := make([]string, 0)
	for _, val := range input[0].Features {
		featureTitles = append(featureTitles, val.Title)
	}
	slices.Sort(featureTitles)
	featureTitles = append([]string{"Model"}, featureTitles...)
	result.WriteString(titlesToTableHeader(featureTitles))

	for _, model := range input {
		row := fmt.Sprintf("| %s |", model.Title)
		slices.SortFunc(model.Features, sortFeaturesFn)
		for _, feature := range model.Features {
			row = fmt.Sprintf("%s %s |", row, aiCompatBoolEmoji(feature.Supported))
		}
		result.WriteString(row + "\n")
	}

	return strings.TrimSpace(result.String())
}

// aiFeaturesToTable renders a set of feature objects into a Markdown table.
// It is added to the AI Monitoring template as a convenience function.
func aiFeaturesToTable(input []AiCompatFeature) string {
	result := strings.Builder{}

	titles := make([]string, 0)
	for _, val := range input {
		titles = append(titles, val.Title)
	}
	slices.Sort(titles)
	result.WriteString(titlesToTableHeader(titles))

	slices.SortFunc(input, sortFeaturesFn)
	for _, feature := range input {
		col := fmt.Sprintf("| %s ", aiCompatBoolEmoji(feature.Supported))
		result.WriteString(col)
	}
	result.WriteString("|\n")

	return strings.TrimSpace(result.String())
}

// aiProvidersToTable renders a set of provider objects into a Markdown table.
// It is added to the AI Monitoring template as a convenience function.
func aiProvidersToTable(input []AiCompatProvider) string {
	result := strings.Builder{}

	result.WriteString("| Provider | Supported | Transitively |\n")
	result.WriteString("| --- | --- | --- |\n")

	for _, provider := range input {
		row := fmt.Sprintf(
			"| %s | %s | %s |\n",
			provider.Name,
			aiCompatBoolEmoji(provider.Supported),
			aiCompatBoolEmoji(provider.Transitively),
		)
		result.WriteString(row)
	}

	return strings.TrimSpace(result.String())
}

// titlesToTableHeader converts a set of strings into a Markdown table heading.
// That is, a row of names followed by a row of heading markers.
func titlesToTableHeader(input []string) string {
	header := "|"
	separator := "|"

	for _, val := range input {
		header = fmt.Sprintf("%s %s |", header, val)
		separator = fmt.Sprintf("%s --- |", separator)
	}

	return header + "\n" + separator + "\n"
}

// sortFeaturesFn is provided to [slices.SortFunc] as the function that
// determines the ordering of features.
func sortFeaturesFn(a AiCompatFeature, b AiCompatFeature) int {
	if a.Title < b.Title {
		return -1
	}
	return 1
}
