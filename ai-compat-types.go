package main

const AiCompatKindGateway = "gateway"
const AiCompatKindAbstraction = "abstraction"
const AiCompatKindSdk = "sdk"

type AiCompatJson []AiCompatEnvelope

type AiCompatEnvelope struct {
	Kind      string             `json:"kind"`
	Title     string             `json:"title"`
	Models    []AiCompatModel    `json:"models,omitempty"`
	Features  []AiCompatFeature  `json:"features,omitempty"`
	Providers []AiCompatProvider `json:"providers,omitempty"`
}

type AiCompatFeature struct {
	Title     string `json:"title"`
	Supported bool   `json:"supported"`
}

type AiCompatModel struct {
	Name     string            `json:"name"`
	Features []AiCompatFeature `json:"features"`
}

type AiCompatProvider struct {
	Name         string `json:"name"`
	Supported    bool   `json:"supported"`
	Transitively bool   `json:"transitively"`
}

type AiCompatTemplateData struct {
	Bedrock   AiCompatBedrockData
	Langchain AiCompatLangchainData
	Openai    AiCompatOpenaiData
}

type AiCompatBedrockData struct {
	Title  string
	Models []struct {
		Name  string
		Text  bool
		Image bool
	}
}

type AiCompatLangchainData struct {
	Title    string
	Features struct {
		Agents       bool
		Chains       bool
		Vectorstores bool
		Tools        bool
	}
	Providers []struct {
		Name         string
		Supported    bool
		Transitively bool
	}
}

type AiCompatOpenaiData struct {
	Completions bool
	Chat        bool
	Embeddings  bool
	Files       bool
	Images      bool
	Audio       bool
}
