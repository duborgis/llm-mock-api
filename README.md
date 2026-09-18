# llm-mock-api

A mock LLM server that simulates OpenAI, AWS Bedrock, and Google Vertex AI (Gemini) APIs with
high wire-level fidelity, for load testing, component testing, and integration testing.

## Architecture (Hexagonal / Ports & Adapters)

```
                         ┌────────────────────────────┐
  openai client  ──HTTP──▶  adapters/inbound/http/openai │
  bedrock client ──HTTP──▶  adapters/inbound/http/bedrock├──▶ ports.Inbound  ──▶ core/usecase ──▶ ports.Outbound ──▶ adapters/outbound/scenario (in-memory + YAML)
  vertex client  ──HTTP──▶  adapters/inbound/http/vertex  │                    (Chat, Embedding, Image,           adapters/outbound/clock
                         └────────────────────────────┘                    AudioTranscription, AudioSpeech,
                                                                              Video, ModelCatalog)
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

The core idea of this project is that each provider's **official Go SDK is the source of truth**
for wire shape, not hand-copied documentation. In practice that holds fully for two of the three
providers and only partially for the third, because of a constraint in AWS's own SDK. The table
below says exactly which files to open to check this for yourself.

| Provider | Uses SDK structs directly for JSON? | Where to look | Why |
|---|---|---|---|
| OpenAI — chat/responses/embeddings/images/audio | **Yes** | `internal/adapters/inbound/http/openai/*.go` | `openai-go` structs have real JSON tags |
| OpenAI — video (Sora) | **No SDK exists** | `internal/adapters/inbound/http/openai/video_handler.go` | `openai-go` has no video support at all yet (see below) |
| Vertex | **Yes** (request envelope is a thin wrapper) | `internal/adapters/inbound/http/vertex/*.go` | `genai` structs have real JSON tags |
| Bedrock | **No — hand-rolled mirrors** | `internal/adapters/inbound/http/bedrock/types.go` | `bedrockruntime` structs have no JSON tags (see below) |

Find every hand-rolled mirror struct (the exception to "SDK is the source of truth") with:
```bash
grep -rln "[Hh]and-rolled mirror" internal/adapters/inbound/http
# internal/adapters/inbound/http/bedrock/types.go
# internal/adapters/inbound/http/vertex/types.go   (request envelope only, see below)
```
And confirm the honest, direct SDK usage in the other two providers with:
```bash
grep -rn '"github.com/openai/openai-go"' internal/adapters/inbound/http/openai/*.go
grep -rn '"google.golang.org/genai"' internal/adapters/inbound/http/vertex/*.go
```

- **OpenAI** (`internal/adapters/inbound/http/openai`): decodes/encodes directly with `openai-go`'s own request/response Go structs (`ChatCompletionNewParams`, `ChatCompletion`, `ChatCompletionChunk`, `EmbeddingNewParams`, `CreateEmbeddingResponse`, `ImageGenerateParams`, `ImagesResponse`, `AudioTranscriptionNewParams`, `Transcription`, `AudioSpeechNewParams`, `Model`, ...) via standard `encoding/json` — these types carry real JSON tags and custom `(Un)MarshalJSON`, so reusing them directly is exact wire fidelity. SSE streaming uses OpenAI's `data: {...}\n\n` framing, terminated by `data: [DONE]\n\n`.
  - **Video generation (Sora) is the one documented exception**, not a Bedrock-style deliberate choice: `openai-go` (currently v1.12.0) has **no video support at all** — no `video.go`, no `Video`/`VideoCreateParams` types, nothing. The Python SDK (`openai` on PyPI, which LiteLLM itself depends on) does have `client.videos` fully modeled (`create`/`retrieve`/`download_content`/...), so `internal/adapters/inbound/http/openai/video_handler.go` hand-keeps small Go structs (`videoCreateRequest`, `videoResponse`) matching `openai-python`'s `video_create_params.VideoCreateParams` / `openai.types.video.Video` field-for-field, with a comment on each explaining the source. If/when `openai-go` adds video support, these should be replaced with the real SDK structs to restore full "SDK is the source of truth" fidelity — check `go list -m github.com/openai/openai-go@latest` periodically for this. Video generation is also asynchronous in the real API (create a job → poll `GET /v1/videos/{id}` for status → `GET /v1/videos/{id}/content` once `completed`); this mock always completes the job synchronously on create, since it has nothing to actually render.
- **Bedrock — the exception** (`internal/adapters/inbound/http/bedrock`): the `bedrockruntime` Go SDK's request/response structs (`ConverseInput/Output`, `InvokeModelInput/Output`, ...) have **no** `encoding/json` tags — the real SDK serializes them through AWS's smithy restjson1 protocol machinery, not `encoding/json`. Reusing those Go structs for our own marshal/unmarshal isn't viable, so `internal/adapters/inbound/http/bedrock/types.go` hand-rolls small JSON-tagged mirror structs matching Bedrock's documented wire shape field-for-field (every such struct in that file carries a `// Hand-rolled mirror of ...` comment, which is what the `grep` above matches). Fidelity is validated a different way instead of by construction: `test/contract/bedrock_contract_test.go` points a real `bedrockruntime.Client` at the mock and lets the SDK's own smithy deserializer parse the response — if the real client parses it cleanly, the hand-rolled mirror is faithful.
  - `InvokeModel`/`InvokeModelWithResponseStream` bodies use the Anthropic Claude Messages API shape (the most common Bedrock provider family) — the body itself is an opaque `[]byte` as far as Bedrock's envelope is concerned, so this is exactly Anthropic's own documented format.
  - `ConverseStream`/`InvokeModelWithResponseStream` use AWS's real `application/vnd.amazon.eventstream` binary framing, built with `aws-sdk-go-v2/aws/protocol/eventstream`'s own `Encoder` — not a simplified substitute.
- **Vertex AI / Gemini** (`internal/adapters/inbound/http/vertex`): uses `google.golang.org/genai`'s types (`Content`, `Part`, `Candidate`, `GenerateContentResponse`, ...) directly — these do carry standard JSON tags, so no hand-rolled mirrors are needed for the response side. `google.golang.org/genai` was chosen over the legacy generated `aiplatform/v1` client because its types map cleanly onto the REST wire shape with plain JSON tags. The one hand-rolled piece is the top-level *request* envelope (`internal/adapters/inbound/http/vertex/types.go`, `generateContentRequest`), because `genai.GenerateContentConfig` is shaped for client-side use, not for decoding an incoming request body — but its `Contents`/`SystemInstruction` fields are `[]*genai.Content` / `*genai.Content` straight from the SDK. Vertex's Go SDK client has limited support for pointing at a custom base URL, so `test/contract/vertex_contract_test.go` uses a plain `http.Client` against the mock and unmarshals into `genai.GenerateContentResponse` directly — proving the JSON shape round-trips through Google's own generated struct, which is the fidelity property that matters.
  - `streamGenerateContent` uses Vertex's real framing: a single JSON **array** of response objects sent incrementally over chunked transfer encoding (not SSE, unlike OpenAI).

## Running the server

```bash
make run
# or: go run ./cmd/server
```

Env vars (all optional):
- `MOCK_HTTP_PORT` (default `9000`) — provider-facing mock API port.
- `MOCK_ADMIN_PORT` (default `9001`) — admin API port.
- `MOCK_SCENARIO_FILE` — path to a YAML scenario file (see `admin/scenarios.example.yaml`) loaded at startup.

`9000`/`9001` were picked to match this machine's nginx-on-Windows -> WSL stream proxy
(`streams.conf`), which forwards ports 9000-9099 and 2222 straight through to WSL's `eth0`
address. That means the mock is reachable from Windows at `localhost:9000`/`9001` with no
extra port-forwarding setup, as long as the WSL `eth0` IP in `streams.conf` still matches
`ip addr show eth0`.

## Pointing each provider's SDK at the mock

- **OpenAI**: `option.WithBaseURL("http://localhost:9000/v1/")` on `openai.NewClient(...)`, or set `OPENAI_BASE_URL=http://localhost:9000/v1`.
- **AWS Bedrock**: `bedrockruntime.New(bedrockruntime.Options{BaseEndpoint: aws.String("http://localhost:9000"), Region: "us-east-1", Credentials: ...})`, or `AWS_ENDPOINT_URL_BEDROCK_RUNTIME=http://localhost:9000`.
- **Vertex AI**: point requests at `http://localhost:9000/v1/projects/<project>/locations/<location>/publishers/google/models/<model>:generateContent` (the Go genai/aiplatform SDKs have limited custom-endpoint support; a plain `http.Client` works reliably, as the contract test demonstrates).

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
make test-bdd         # BDD (godog) suite against the live docker-compose stack, see below
```

The contract tests in `test/contract/` spin up the real server with `httptest.NewServer` and
call it with each provider's real SDK client (OpenAI, Bedrock) or by unmarshaling into the
SDK's own generated structs (Vertex), proving round-trip fidelity rather than asserting on
hand-written expectations of the wire format.

### BDD tests (godog) against the live stack

`test/bdd/` is a [godog](https://github.com/cucumber/godog) (the Go equivalent of Python's
`behave`) Gherkin suite that drives **real HTTP calls through the running LiteLLM proxy** — not
the in-process test server the contract tests use — and asserts on side effects in the other
services (OpenMeter events, raw responses in Mongo). It requires `make litellm-up` to already be
running, and is gated behind a `bdd` build tag so `go test ./...`/`make test` never touch it:

```bash
make litellm-up
make test-bdd
```

`test/bdd/features/openmeter_events.feature` covers one scenario per OpenMeter-event-producing
route (chat completions, Responses API, image generation, audio transcription, video generation)
plus the audio/speech case that's *expected* to fail OpenMeter's own event but still get its raw
response captured (see "Known LiteLLM limitation" above) — each ending with a step that prints
the current MongoDB event contents, so a run is easy to eyeball rather than just trust green.

`make test-bdd` always passes `-count=1`: this suite has real side effects on a live stack (HTTP
calls through LiteLLM), so `go test`'s result cache — which would silently skip re-running it and
report a stale "ok (cached)" — must stay disabled here.

## Testing against LiteLLM

`docker-compose.yaml` at the repo root brings up a full local stack to exercise the mock through
a **real LiteLLM proxy**, not just direct HTTP calls — the same way an application would actually
use it:

```
localhost:4000 (litellm) ──┬─ openai/*      ──▶ mock-api:9000  (real openai-go structs)
                            ├─ bedrock/*      ──▶ mock-api:9000  (real bedrockruntime client)
                            └─ vertex_ai/*    ──▶ mock-api:9000  (real genai structs)
                                   │
                                   ▼ (service-account OAuth token exchange)
                            nginx:8080 /oauth2/ ──▶ wiremock:8080  (fakes oauth2.googleapis.com)
```

- **`mock-api`** — this project, built from `Dockerfile` using a Linux binary compiled on the
  host (`make build-linux`) rather than inside the Docker build — this environment's Docker
  network has flaky/slow HTTPS to `proxy.golang.org`, and the host toolchain already works.
- **`wiremock`** (`wiremock/mappings/oauth-token.json`) — stubs `POST /token` to return a fake
  `{"access_token": ...}`, standing in for Google's real `oauth2.googleapis.com`.
- **`nginx`** (`nginx/nginx.conf`) — single front door on port 8080: `/llm/` → litellm,
  `/mock/` → mock-api directly, `/oauth2/` → wiremock.
- **`litellm`** (`litellm/config.yaml`) — the real LiteLLM proxy, configured to route
  `openai/*` and `vertex_ai/*` via `api_base` and `bedrock/*` via
  `aws_bedrock_runtime_endpoint`, all pointed at `mock-api`.

The tricky part is Vertex: LiteLLM's Vertex integration goes through Google's real
`google-auth` library, which insists on a service-account JSON and, before making any network
call at all, validates its shape locally. `scripts/gen-fake-vertex-credentials.py` generates a
throwaway RSA key (never registered with Google, generated fresh on this machine, gitignored
under `litellm/secrets/`) with its `token_uri` field pointed at `http://nginx:8080/oauth2/token`
instead of the real Google endpoint — so the entire OAuth token exchange happens locally against
wiremock, and Vertex chat completions reach the mock without any real GCP credentials or network
access.

```bash
make litellm-up     # builds the linux binary, generates the fake credentials, docker compose up
make litellm-logs   # tail all four services
make litellm-down   # tear the stack down

# Then, e.g.:
curl -X POST localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer sk-litellm-local-test" -H "Content-Type: application/json" \
  -d '{"model":"vertex_ai/gemini-2.5-pro","messages":[{"role":"user","content":"oi"}]}'
```

One LiteLLM quirk worth knowing if you add more Vertex models to `litellm/config.yaml`: it
appends `:generateContent` straight onto `api_base` without inserting the model id itself, so
each Vertex model entry's `api_base` needs to already end in that specific model's path segment
(`.../models/<model-id>`), not just `.../models`.

## Testing OpenMeter usage events

Every successful call through LiteLLM also logs a usage event to a mock OpenMeter collector,
persisted into MongoDB so it can actually be inspected — not just trusted to have fired.

```
litellm ──POST /api/v1/events──▶ openmeter-mock:9010 ──▶ mongo:27017 (db: openmeter_mock)
   (litellm/integrations/openmeter.py,                        exposed on host :9017
    the real LiteLLM code, unmodified)
```

- **`cmd/openmeter-mock`** (`internal/openmeter/...`, same hexagonal shape as the rest of this
  repo: `domain.Event`, `ports.EventRepository`/`EventIngestUseCase`, a thin `usecase.Ingest`,
  an `internal/adapters/outbound/mongo` repository, and an
  `internal/adapters/inbound/http/openmeter` HTTP adapter) imitates the one route LiteLLM's
  OpenMeter integration calls: `POST /api/v1/events`, CloudEvents-shaped JSON. It also exposes
  `GET /api/v1/events?limit=N` as a convenience for inspecting recent events without a Mongo
  client.
- **`mongo`** stores every event. It's on host port **9017** — inside the 9000-9099 block this
  machine's Windows-side nginx already TCP-proxies into WSL (see "Running the server" above),
  so a native Mongo client on Windows (mongosh, Compass) can connect to `localhost:9017` with
  no extra setup.
- LiteLLM's OpenMeter integration is registered the normal, documented way — `callbacks:
  ["openmeter"]` under `litellm_settings` in `litellm/config.yaml`, with `OPENMETER_API_ENDPOINT`
  and `OPENMETER_API_KEY` set as env vars on the `litellm` service in `docker-compose.yaml`.
  **The request must include a `user`** (top-level in the JSON body, or resolved from the
  calling API key's metadata) — LiteLLM's `OpenMeterLogger` raises otherwise, since the event
  needs a subject.
- `litellm` is pinned to **`v1.88.4`**, not `main-latest`: an earlier version this stack was
  tested against had a real bug where the module-level `openMeterLogger` instance the built-in
  `"openmeter"` callback depends on was never actually constructed on the *proxy server* code
  path (only the plain-Python `litellm.completion(...)` SDK path triggered it) — the callback
  registered without error but silently never fired anything, for any config. v1.88.4 does not
  have this problem; confirmed by actually watching an event land in Mongo, not just by the
  callback list printing correctly at startup.

```bash
curl -X POST localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer sk-litellm-local-test" -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o","user":"someone@example.com","messages":[{"role":"user","content":"oi"}]}'

curl localhost:9010/api/v1/events?limit=10   # or connect a Mongo client to localhost:9017
```

### Known LiteLLM limitation: no event for `/v1/audio/speech`

LiteLLM's own `"openmeter"` callback (v1.88.4) throws an uncaught exception internally when it
tries to read token usage off an `audio/speech` response — that route returns raw audio bytes,
which the callback's usage-extraction code doesn't expect. This is a bug in LiteLLM itself, not
the mock (confirmed by reading `litellm/integrations/openmeter.py` inside the running container).
No OpenMeter event is emitted for `/v1/audio/speech` calls as a result — see the raw-response
capture below for how this mock still records that call anyway.

## Capturing the full raw response (custom LiteLLM callback)

OpenMeter's event only carries `{model, cost, prompt_tokens, completion_tokens, total_tokens}` —
useful for usage/billing, useless for actually inspecting what the mock returned. A second,
independent mechanism captures the **full** provider-shaped response for every call:

```
litellm ──POST /api/v1/raw-responses──▶ openmeter-mock:9010 ──▶ mongo:27017 (collection: raw_responses)
   (litellm/custom_callback.py, our own code,
    reading kwargs["original_response"])
```

- **`litellm/custom_callback.py`** — a `CustomLogger` (`raw_response_logger`) registered in
  `litellm/config.yaml` (`callbacks: [..., "custom_callback.raw_response_logger"]`) and mounted
  into the `litellm` container alongside `config.yaml`. On every successful call it forwards
  `kwargs["original_response"]` (the exact JSON/dict LiteLLM got back from the mock) to
  `openmeter-mock`'s `POST /api/v1/raw-responses`.
- **`openmeter-mock`** persists it into MongoDB (collection `raw_responses`, upserted by
  `call_id` so retried deliveries don't duplicate) and exposes `GET /api/v1/raw-responses?limit=N`
  to inspect it — same pattern as `GET /api/v1/events`, just a different collection.
- This is intentionally **independent** of the `"openmeter"` callback: it has no assumption about
  token usage, so it keeps working even for routes (like `/v1/audio/speech`, above) where
  OpenMeter's own integration silently fails.

```bash
curl localhost:9010/api/v1/raw-responses?limit=5
```

## Observability: OpenTelemetry traces via Jaeger

LiteLLM's built-in `"otel"` callback exports one span per proxied call over OTLP/gRPC to a local
all-in-one **Jaeger** instance (`docker-compose.yaml`), UI at **`localhost:9018`**
(within the 9000-9099 block, see "Running the server" above).

```
litellm ──OTLP/gRPC:4317──▶ jaeger (all-in-one, OTLP receiver + UI on :9018)
```

- Configured entirely via env vars on the `litellm` service: `OTEL_EXPORTER=otlp_grpc`,
  `OTEL_ENDPOINT=jaeger:4317`, `OTEL_EXPORTER_OTLP_TRACES_INSECURE=true` (Jaeger's OTLP receiver
  here is plaintext; LiteLLM's exporter defaults to TLS, and without this flag Jaeger just logs a
  "bogus greeting" and drops every span), plus `litellm_settings.callbacks: [..., "otel"]` in
  `litellm/config.yaml`.
- Each trace has one span per proxy stage (`router`, `auth`, `proxy_pre_call`,
  `litellm_request`, ...), plus a **`raw_gen_ai_request`** span carrying the original
  provider-shaped request/response as `llm.<provider>.*` attributes (e.g. `llm.openai.choices`,
  `llm.openai.usage`) — this is a *different* place than the `raw_responses` Mongo collection
  above to see the same data, useful for quickly eyeballing one call's shape in the Jaeger UI
  without hitting `GET /api/v1/raw-responses`.
- Jaeger's all-in-one image stores traces in memory only — they're gone on container restart,
  unlike the OpenMeter events and raw responses, which are in MongoDB.

```bash
open http://localhost:9018   # or curl http://localhost:9018/api/traces?service=litellm-mock-proxy
```

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
- **Tool/function calling and multimodal (image) chat content** — domain types have room for
  multimodal `ContentPart`, but the chat usecase/adapters currently only exercise `text` parts.
  (Image *generation*, not multimodal chat input, is implemented — see "Fidelity strategy" above.)
- **OpenAI video generation (Sora) has no Go SDK backing** — see "Fidelity strategy" above; it's
  implemented against hand-kept structs mirroring `openai-python`, not `openai-go`, since the
  latter doesn't have video support yet. Revisit once it does.
- **Video edit/extend/remix/character endpoints** — only create/status/content are implemented;
  LiteLLM and the OpenAI API also expose `/v1/videos/{id}/remix`, `/v1/videos/edits`, character
  management, etc., none of which are wired up.
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
