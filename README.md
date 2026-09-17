# llm-mock-api

A mock LLM server that simulates OpenAI, AWS Bedrock, and Google Vertex AI (Gemini) APIs with
high wire-level fidelity, for load testing, component testing, and integration testing.

## Architecture (Hexagonal / Ports & Adapters)

```
                         ┌────────────────────────────┐
  openai client  ──HTTP──▶  adapters/inbound/http/openai │
  bedrock client ──HTTP──▶  adapters/inbound/http/bedrock├──▶ ports.Inbound  ──▶ core/usecase ──▶ ports.Outbound ──▶ adapters/outbound/scenario (in-memory + YAML)
  vertex client  ──HTTP──▶  adapters/inbound/http/vertex  │                         (Chat, Embedding,             adapters/outbound/clock
                         └────────────────────────────┘                         ModelCatalog)
                                                                                        ▲
  admin client   ──HTTP──▶  adapters/admin/http ─────────────────────────────────────────┘ (Put/List scenarios)
```

- `internal/core/domain` — generic, provider-agnostic types (ChatRequest/Response, Scenario, TokenUsage, ...). Zero imports of any SDK, `net/http`, or adapter code.
- `internal/core/ports` — interfaces only. `inbound.go` = driving ports the usecases expose (ChatCompletionUseCase, EmbeddingUseCase, ModelCatalogUseCase). `outbound.go` = driven ports the usecases need (ScenarioRepository, ModelRepository, Clock).
- `internal/core/usecase` — pure business logic implementing the driving ports, depending only on the driven ports. Unit-tested with in-memory fakes, no HTTP, no SDKs (`internal/core/usecase/*_test.go`).
- `internal/adapters/inbound/http/{openai,bedrock,vertex}` — one HTTP adapter per provider. All provider-specific wire knowledge (JSON field names, SSE framing, AWS eventstream binary framing, error envelopes) lives here, and only here.
- `internal/adapters/outbound/scenario` — in-memory `ScenarioRepository`/`ModelRepository` implementation, plus a YAML loader for seeding it at startup.
- `internal/adapters/outbound/clock` — real-time `Clock` implementation.
- `internal/adapters/admin/http` — a small non-provider API (`/admin/scenarios`) for configuring canned responses, latency, and error injection at runtime.
- `cmd/server/main.go` — the only place that constructs concrete adapters and wires them into usecases via the ports (manual dependency injection).

Verify the core stays decoupled at any time:
```bash
grep -rl 'openai-go\|aws-sdk-go\|google-api-go-client\|genai\|net/http' internal/core
# (expected: no output)
```

## Fidelity strategy per provider

- **OpenAI** (`internal/adapters/inbound/http/openai`): decodes/encodes directly with `openai-go`'s own request/response Go structs (`ChatCompletionNewParams`, `ChatCompletion`, `ChatCompletionChunk`, `EmbeddingNewParams`, `CreateEmbeddingResponse`, `Model`, ...) via standard `encoding/json` — these types carry real JSON tags and custom `(Un)MarshalJSON`, so reusing them directly is exact wire fidelity. SSE streaming uses OpenAI's `data: {...}\n\n` framing, terminated by `data: [DONE]\n\n`.
- **Bedrock** (`internal/adapters/inbound/http/bedrock`): the `bedrockruntime` Go SDK's request/response structs (`ConverseInput/Output`, `InvokeModelInput/Output`, ...) have **no** `encoding/json` tags — the real SDK serializes them through AWS's smithy restjson1 protocol machinery, not `encoding/json`. Reusing those Go structs for our own marshal/unmarshal isn't viable, so `internal/adapters/inbound/http/bedrock/types.go` hand-rolls small JSON-tagged mirror structs matching Bedrock's documented wire shape field-for-field. Fidelity is validated the same way regardless: `test/contract/bedrock_contract_test.go` points a real `bedrockruntime.Client` at the mock and lets the SDK's own smithy deserializer parse the response.
  - `InvokeModel`/`InvokeModelWithResponseStream` bodies use the Anthropic Claude Messages API shape (the most common Bedrock provider family) — the body itself is an opaque `[]byte` as far as Bedrock's envelope is concerned, so this is exactly Anthropic's own documented format.
  - `ConverseStream`/`InvokeModelWithResponseStream` use AWS's real `application/vnd.amazon.eventstream` binary framing, built with `aws-sdk-go-v2/aws/protocol/eventstream`'s own `Encoder` — not a simplified substitute.
- **Vertex AI / Gemini** (`internal/adapters/inbound/http/vertex`): uses `google.golang.org/genai`'s types (`Content`, `Part`, `Candidate`, `GenerateContentResponse`, ...) directly — these do carry standard JSON tags, so no hand-rolled mirrors are needed for the response side. `google.golang.org/genai` was chosen over the legacy generated `aiplatform/v1` client because its types map cleanly onto the REST wire shape with plain JSON tags. Vertex's Go SDK client has limited support for pointing at a custom base URL, so `test/contract/vertex_contract_test.go` uses a plain `http.Client` against the mock and unmarshals into `genai.GenerateContentResponse` directly — proving the JSON shape round-trips through Google's own generated struct, which is the fidelity property that matters.
  - `streamGenerateContent` uses Vertex's real framing: a single JSON **array** of response objects sent incrementally over chunked transfer encoding (not SSE, unlike OpenAI).

## Running the server

```bash
make run
# or: go run ./cmd/server
```

Env vars (all optional):
- `MOCK_HTTP_PORT` (default `8080`) — provider-facing mock API port.
- `MOCK_ADMIN_PORT` (default `8081`) — admin API port.
- `MOCK_SCENARIO_FILE` — path to a YAML scenario file (see `admin/scenarios.example.yaml`) loaded at startup.

## Pointing each provider's SDK at the mock

- **OpenAI**: `option.WithBaseURL("http://localhost:8080/v1/")` on `openai.NewClient(...)`, or set `OPENAI_BASE_URL=http://localhost:8080/v1`.
- **AWS Bedrock**: `bedrockruntime.New(bedrockruntime.Options{BaseEndpoint: aws.String("http://localhost:8080"), Region: "us-east-1", Credentials: ...})`, or `AWS_ENDPOINT_URL_BEDROCK_RUNTIME=http://localhost:8080`.
- **Vertex AI**: point requests at `http://localhost:8080/v1/projects/<project>/locations/<location>/publishers/google/models/<model>:generateContent` (the Go genai/aiplatform SDKs have limited custom-endpoint support; a plain `http.Client` works reliably, as the contract test demonstrates).

## Configuring mock behavior

At startup, seed `admin/scenarios.example.yaml`-style definitions via `MOCK_SCENARIO_FILE`. At
runtime, use the admin API:

```bash
curl -X PUT localhost:8081/admin/scenarios -d '{
  "model": "gpt-4o-mini",
  "operation": "chat",
  "response_text": "custom reply",
  "latency_ms": 500
}'

curl localhost:8081/admin/scenarios
```

## Running tests

```bash
make test            # unit tests (usecases with fakes) + contract tests
make test-contract    # contract tests only, verbose
```

The contract tests in `test/contract/` spin up the real server with `httptest.NewServer` and
call it with each provider's real SDK client (OpenAI, Bedrock) or by unmarshaling into the
SDK's own generated structs (Vertex), proving round-trip fidelity rather than asserting on
hand-written expectations of the wire format.

## Adding a new provider or endpoint

1. Add a domain type to `internal/core/domain` only if the existing generic types
   (`ChatRequest`, `EmbeddingRequest`, `ModelInfo`, ...) don't already cover it.
2. If new business behavior is needed, add it to `internal/core/usecase`, tested with a fake
   outbound port — no HTTP, no SDK.
3. Add a new inbound adapter package under `internal/adapters/inbound/http/<provider>` with its
   own `router.go` and handler files, translating wire format to/from the domain types via the
   provider's own SDK types (or hand-rolled mirrors if the SDK's types can't be marshaled with
   `encoding/json` directly, as with Bedrock).
4. Wire the new router into `cmd/server/main.go`.
5. Add a contract test under `test/contract/`.

## Roadmap / not yet covered (intentionally deferred)

- **Fine-tuning, batch, and file APIs** (all three providers) — out of scope for this pass.
- **OpenAI moderations** — a minimal stub is implemented (`always not flagged`), not wired to any Scenario.
- **OpenAI legacy `/v1/completions`** — not implemented; the modern chat/messages surface was prioritized.
- **Tool/function calling, multimodal (image) content, and audio** — domain types have room for
  multimodal `ContentPart`, but the usecases and adapters currently only exercise `text` parts.
- **Bedrock non-Anthropic provider bodies** (Amazon Titan text, Meta Llama, Mistral, etc. on
  `InvokeModel`) — only the Anthropic Claude Messages body shape is implemented, per the spec's
  "at minimum" requirement; Titan *embeddings* are the one Bedrock embedding surface stubbed
  (see below).
- **Embeddings**: OpenAI is fully implemented (deterministic pseudo-embeddings). Bedrock Titan
  and Vertex embeddings endpoints are not implemented in this pass — OpenAI chat/embeddings,
  Bedrock/Vertex chat, and model listing were prioritized first, per the requested order.
- **Vertex `aiplatform/v1` generated client** — not used; `google.golang.org/genai` was chosen
  instead (see Fidelity strategy above) since its types have plain JSON tags and map directly
  onto the REST wire contract.
- **Guardrails, prompt management, async invoke, and other Bedrock control-plane surfaces**
  beyond `foundation-models` listing — out of scope.
- **Admin API** has no authentication — it's a test-control surface, not meant to be exposed
  publicly.
