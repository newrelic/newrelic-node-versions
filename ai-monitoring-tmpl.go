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

// aiCompatBuildTmplData translates the parsed JSON data into a structure
// that our template is able to use.
func aiCompatBuildTmplData(input AiCompatJson) AiCompatTemplateData {
	result := AiCompatTemplateData{
		Abstractions: make([]AiCompatAbstraction, 0),
		Gateways:     make([]AiCompatGateway, 0),
		Sdks:         make([]AiCompatSdk, 0),
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
			sdk := aiCompatParseSdk(envelope)
			result.Sdks = append(result.Sdks, sdk)
		}
	}

	return result
}

// aiCompatParseGateway translates an envelope of `kind = "gateway"` into
// a narrow gateway type.
func aiCompatParseGateway(input AiCompatEnvelope) AiCompatGateway {
	return AiCompatGateway{
		Title:    input.Title,
		Preamble: input.Preamble,
		Footnote: input.Footnote,
		Models:   input.Models,
	}
}

// aiCompatParseAbstraction translates an envelope of `kind = "abstraction"`
// into a narrow abstraction type.
func aiCompatParseAbstraction(input AiCompatEnvelope) AiCompatAbstraction {
	return AiCompatAbstraction{
		Title:             input.Title,
		FeaturesPreamble:  input.FeaturesPreamble,
		ProvidersPreamble: input.ProvidersPreamble,
		Features:          input.Features,
		Providers:         input.Providers,
	}
}

// aiCompatParseSdk translates an envelope of `kind = "sdk"` into a narrow
// sdk type.
func aiCompatParseSdk(input AiCompatEnvelope) AiCompatSdk {
	return AiCompatSdk{
		Title:            input.Title,
		FeaturesPreamble: input.FeaturesPreamble,
		Features:         input.Features,
	}
}

// aiCompatLoadTemplate loads the templated Markdown into a [text/template]
// instance that has all of our custom utility methods attached to it.
func aiCompatLoadTemplate() (*template.Template, error) {
	tmpl := template.New("aiMonitoring")

	tmpl.Funcs(template.FuncMap{
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
func aiModelsToTable(input []AiCompatModel) string {
	result := strings.Builder{}

	featureTitles := make([]string, 0)
	for _, val := range input[0].Features {
		featureTitles = append(featureTitles, val.Title)
	}
	slices.Sort(featureTitles)
	featureTitles = append([]string{"Model"}, featureTitles...)
	result.WriteString(titlesToTableHeader(featureTitles))

	// Gateways have multiple models behind them. Each model usually has an
	// overlapping feature set, but each model may have a feature other models
	// in the list do not. Which means we have a mismatch in the number of
	// columns. So we need to build out our rows in a manner such that we can
	// add "gap" columns if a model does not have a feature found in some other
	// model.
	modelNames := make([]string, 0)
	rows := make(map[string]string)
	for _, title := range featureTitles {
		if title == "Model" {
			for _, model := range input {
				modelNames = append(modelNames, model.Name)
				rows[model.Name] = fmt.Sprintf("| %s |", model.Name)
			}
			continue
		}

		for _, model := range input {
			name := model.Name
			idx := slices.IndexFunc(model.Features, func(f AiCompatFeature) bool { return f.Title == title })
			if idx == -1 {
				rows[name] += " - |"
			} else {
				feature := model.Features[idx]
				rows[name] = fmt.Sprintf("%s %s |", rows[name], aiCompatBoolEmoji(feature.Supported))
			}
		}
	}

	slices.Sort(modelNames)
	for _, name := range modelNames {
		result.WriteString(rows[name] + "\n")
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
