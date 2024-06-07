## AI Monitoring Support

The Node.js agent supports the following AI platforms and integrations.

{{range .Gateways -}}
### {{.Title}}

{{if .Preamble}}{{.Preamble}}{{end}}

{{gatewayModelsToTable .Models}}

{{if .Footnote}}{{.Footnote}}{{end}}
{{end}}

### Langchain

The following general features of Langchain are supported:

| Agents | Chains | Vectorstores | Tools |
| --- | --- | --- | --- |
{{- with .Langchain.Features}}
| {{boolEmoji .Agents}} | {{boolEmoji .Chains}} | {{boolEmoji .Vectorstores}} | {{boolEmoji .Tools}} |
{{end}}

Models/providers are generally supported transitively by our instrumentation of
the provider's module.

| Provider | Supported | Transitively |
| --- | --- | --- |
{{range .Langchain.Providers -}}
| {{.Name}} | {{boolEmoji .Supported}} | {{boolEmoji .Transitively}} |
{{end}}

### OpenAI

Through the `openai` module, we support:

| Completions | Chat | Embeddings | Files | Images | Audio |
| --- | --- | --- | --- | --- | --- |
{{- with .Openai}}
| {{boolEmoji .Completions}} | {{boolEmoji .Chat}} | {{boolEmoji .Embeddings}} | {{boolEmoji .Files}} | {{boolEmoji .Images}} | {{boolEmoji .Audio}} |
{{end}}
