## AI Monitoring Support

The Node.js agent supports the following AI platforms and integrations.

{{range .Gateways -}}
### {{.Title}}

{{if .Preamble}}{{.Preamble}}{{end}}

{{gatewayModelsToTable .Models}}

{{if .Footnote}}{{.Footnote}}{{end}}
{{end}}

{{range .Abstractions -}}
### {{.Title}}

{{if .FeaturesPreamble}}{{.FeaturesPreamble}}{{end}}

{{featuresToTable .Features}}

{{if .ProvidersPreamble}}{{.ProvidersPreamble}}{{end}}

{{providersToTable .Providers}}
{{end}}

### OpenAI

Through the `openai` module, we support:

| Completions | Chat | Embeddings | Files | Images | Audio |
| --- | --- | --- | --- | --- | --- |
{{- with .Openai}}
| {{boolEmoji .Completions}} | {{boolEmoji .Chat}} | {{boolEmoji .Embeddings}} | {{boolEmoji .Files}} | {{boolEmoji .Images}} | {{boolEmoji .Audio}} |
{{end}}
