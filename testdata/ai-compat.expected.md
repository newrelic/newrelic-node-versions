## AI Monitoring Support

The Node.js agent supports the following AI platforms and integrations.

### Amazon Bedrock

Through the `@aws-sdk/client-bedrock-runtime` module, we support:

| Model | Image | Text |
| --- | --- | --- |
| Claude | ❌ | ✅ |
| Cohere | ❌ | ✅ |

Note: if a model supports streaming, we also instrument the streaming variant.
### Foo Gateway



| Model | Four | One | Three | Two |
| --- | --- | --- | --- | --- |
| Foo Model | ✅ | ✅ | ❌ | ❌ |




### Langchain

The following general features of Langchain are supported:

| Agents | Chains | Tools | Vectorstores |
| --- | --- | --- | --- |
| ✅ | ✅ | ✅ | ✅ |

Models/providers are generally supported transitively by our instrumentation of the provider's module.

| Provider | Supported | Transitively |
| --- | --- | --- |
| Azure OpenAI | ❌ | ❌ |
| OpenAI | ✅ | ✅ |


### OpenAI

Through the `openai` module, we support:

| Completions | Chat | Embeddings | Files | Images | Audio |
| --- | --- | --- | --- | --- | --- |
| ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |

