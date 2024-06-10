package main

const AiCompatKindGateway = "gateway"
const AiCompatKindAbstraction = "abstraction"
const AiCompatKindSdk = "sdk"

type AiCompatJson []AiCompatEnvelope

// AiCompatEnvelope represents the generic structure of objects found within
// the AI compatability JSON descriptor document.
type AiCompatEnvelope struct {
	Kind              string             `json:"kind"`
	Title             string             `json:"title"`
	Preamble          string             `json:"preamble,omitempty"`
	Footnote          string             `json:"footnote,omitempty"`
	FeaturesPreamble  string             `json:"featuresPreamble,omitempty"`
	ProvidersPreamble string             `json:"providersPreamble,omitempty"`
	Models            []AiCompatModel    `json:"models,omitempty"`
	Features          []AiCompatFeature  `json:"features,omitempty"`
	Providers         []AiCompatProvider `json:"providers,omitempty"`
}

// AiCompatGateway is a narrow type derived from an [AiCompatEnvelope] with
// `kind = "gateway"`.
type AiCompatGateway struct {
	Title    string
	Preamble string
	Footnote string
	Models   []AiCompatModel
}

// AiCompatAbstraction is a narrow type derived from an [AiCompatEnvelope] with
// `kind = "abstraction"`.
type AiCompatAbstraction struct {
	Title             string
	FeaturesPreamble  string
	ProvidersPreamble string
	Features          []AiCompatFeature
	Providers         []AiCompatProvider
}

// AiCompatSdk is a narrow type derived from an [AiCompatEnvelope] with
// `kind = "sdk"`.
type AiCompatSdk struct {
	Title            string
	FeaturesPreamble string
	Features         []AiCompatFeature
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

// AiCompatTemplateData represents the context that will be utilized by the
// template engine to render the human readable Markdown document.
type AiCompatTemplateData struct {
	Abstractions []AiCompatAbstraction
	Gateways     []AiCompatGateway
	Sdks         []AiCompatSdk
}
