Feature: OpenMeter usage events via LiteLLM
  LiteLLM logs a usage event to OpenMeter after each successful call it proxies to the mock.
  These scenarios drive real requests through the local LiteLLM proxy (docker-compose stack)
  and verify a matching event lands in the OpenMeter mock, proving the callback fires for
  every route the mock implements — not just plain chat completions.

  Background:
    Given the LiteLLM proxy and the OpenMeter mock are reachable

  Scenario: Chat completion emits an OpenMeter event
    When I request a chat completion from LiteLLM for model "openai/gpt-4o" with prompt "say hi"
    Then an OpenMeter event for model "gpt-4o" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Bedrock chat completion emits an OpenMeter event
    # Same /v1/chat/completions call as the OpenAI scenario above — LiteLLM routes it to the
    # mock's Bedrock Converse implementation based on the "bedrock/" model prefix in
    # litellm/config.yaml, not a different endpoint.
    When I request a chat completion from LiteLLM for model "bedrock/claude-sonnet-4-5" with prompt "say hi"
    Then an OpenMeter event for model "anthropic.claude-sonnet-4-5-20250929-v1:0" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Vertex chat completion emits an OpenMeter event
    # Same /v1/chat/completions call, routed to the mock's Vertex generateContent
    # implementation based on the "vertex_ai/" model prefix.
    When I request a chat completion from LiteLLM for model "vertex_ai/gemini-2.5-pro" with prompt "say hi"
    Then an OpenMeter event for model "gemini-2.5-pro" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Responses API emits an OpenMeter event
    When I request a response from LiteLLM for model "openai/gpt-4o" with prompt "say hi"
    Then an OpenMeter event for model "gpt-4o" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Image generation emits an OpenMeter event
    When I request an image generation from LiteLLM for model "openai/dall-e-3" with prompt "a red bicycle on the moon"
    Then an OpenMeter event for model "dall-e-3" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Audio transcription emits an OpenMeter event
    When I request an audio transcription from LiteLLM for model "openai/whisper-1"
    Then an OpenMeter event for model "whisper-1" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Audio speech does not emit an OpenMeter event, but its raw response is still captured
    # LiteLLM's own "openmeter" callback (v1.88.4) throws internally when it tries to read
    # token usage off a raw-bytes audio/speech response — a LiteLLM limitation, not the mock's.
    # Our own custom_callback.py raw-response logger has no such assumption, so it still
    # captures the call even where OpenMeter's integration can't.
    When I request audio speech from LiteLLM for model "openai/tts-1" with input "hello world"
    Then a raw response for model "tts-1" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB

  Scenario: Video generation emits an OpenMeter event
    # Sora only exists in openai-python (openai-go has no video support at all yet), and
    # video generation is asynchronous in the real API (create -> poll -> download). This
    # mock always completes the job synchronously, so a single create call is enough here.
    When I request a video generation from LiteLLM for model "openai/sora-2" with prompt "a red bicycle on the moon"
    Then an OpenMeter event for model "sora-2" should appear within 10 seconds
    And I print the OpenMeter events stored in MongoDB
